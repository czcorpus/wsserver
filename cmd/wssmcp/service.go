package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/czcorpus/cnc-gokit/logging"
	"github.com/czcorpus/wsserver/config"
	"github.com/czcorpus/wsserver/core"
	"github.com/czcorpus/wsserver/model"
	"github.com/czcorpus/wsserver/queries"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type GeneralSearcher interface {
	SimilarlyUsedWords(
		ctx context.Context,
		datasetID, modelID, posOrSfn, word string,
		limit int,
		minScore float32,
	) ([]queries.ResultRow, core.AppError)
}

func main() {
	flag.Parse()
	conf, err := config.Load(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %s\n", err)
		os.Exit(1)
	}
	logging.SetupLogging(conf.Logging)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		stop()
	}()

	collDbMap, err := queries.NewCollDbMap(conf.Models)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to instantiate collocation databases: %s\n", err)
		os.Exit(1)
	}

	w2vModels := model.NewProvider(conf.DataDir, conf.Models)

	var searcher GeneralSearcher
	if conf.MCP.SelfContained {
		searcher, err = queries.NewSearchProvider(
			conf.DataDir,
			collDbMap,
			w2vModels,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to instantiate searcher: %s\n", err)
			os.Exit(1)
		}
	} else {
		searcher = NewHTTPClientSearcher()

	}

	// Create a new MCP server
	s := server.NewMCPServer(
		"WSServer",
		"0.0.2",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	similarWordsTool := mcp.NewTool("similarly_used_words",
		mcp.WithDescription("Find words that are similarly used in the dataset"),
		mcp.WithString("dataset_id",
			mcp.Required(),
			mcp.Description("The dataset ID to search in"),
		),
		mcp.WithString("model_id",
			mcp.Required(),
			mcp.Description("The model ID to use"),
		),
		mcp.WithString("pos_or_sfn",
			mcp.Description("Part of speech or surface form"),
		),
		mcp.WithString("word",
			mcp.Required(),
			mcp.Description("The word to find similar usage for"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results to return"),
			// Optional parameter - no mcp.Required()
		),
		mcp.WithNumber("min_score",
			mcp.Description("Minimum similarity score threshold"),
			// Optional parameter - no mcp.Required()
		),
	)

	s.AddTool(
		similarWordsTool,
		func(ctx context.Context, request mcp.CallToolRequest,
		) (*mcp.CallToolResult, error) {
			fmt.Fprintln(os.Stderr, "method invoked: similarly_used_words")
			// Extract required parameters using the helper methods
			datasetID, err := request.RequireString("dataset_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			modelID, err := request.RequireString("model_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			posOrSfn := request.GetString("pos_or_sfn", "")

			word, err := request.RequireString("word")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Extract optional parameters with defaults
			limit := request.GetInt("limit", 10)

			minScore := request.GetFloat("min_score", 0.5)

			// Call your domain function
			results, appErr := searcher.SimilarlyUsedWords(ctx, datasetID, modelID, posOrSfn, word, limit, float32(minScore))
			fmt.Fprintln(os.Stderr, "RESULTS:", results)
			if !appErr.IsZero() {
				return mcp.NewToolResultError(fmt.Sprintf("Error: %v", appErr)), nil
			}
			var formattedRes strings.Builder
			for i, res := range results {
				formattedRes.WriteString(fmt.Sprintf("%d. %s (score %01.2f)\n", i, res.Word, res.Score))
			}
			res := mcp.NewToolResultText(formattedRes.String())
			// Format and return results
			return res, nil
		})

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
	}
}
