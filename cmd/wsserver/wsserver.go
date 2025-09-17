// Copyright 2017 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2017 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/czcorpus/wsserver/actions"
	"github.com/czcorpus/wsserver/config"
	"github.com/czcorpus/wsserver/model"
	"github.com/czcorpus/wsserver/queries"

	"github.com/czcorpus/cnc-gokit/logging"
	"github.com/czcorpus/cnc-gokit/uniresp"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var (
	version     string
	buildDate   string
	gitCommit   string
	versionInfo = config.VersionInfo{
		Version:   version,
		BuildDate: buildDate,
		GitCommit: gitCommit,
	}
)

func main() {
	flag.Parse()

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(logging.GinMiddleware())
	engine.Use(uniresp.AlwaysJSONContentType())

	if flag.Arg(0) != "" {
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

		searcher, err := queries.NewSearchProvider(
			conf.DataDir,
			collDbMap,
			w2vModels,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to instantiate searcher: %s\n", err)
			os.Exit(1)
		}

		log.Printf("INFO: starting to listen on %s:%d", conf.ListenAddress, conf.ListenPort)
		handler, err := actions.NewActionHandler(conf.DataDir, w2vModels, searcher)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to instantiate API handler: %s\n", err)
			os.Exit(1)
		}

		engine.GET(
			"/dataset/:corpusId/dictionary/:word",
			handler.Dictionary,
		)

		engine.GET(
			"/dataset/:corpusId/collocations/:word/:pos",
			handler.Collocations,
		)
		engine.GET(
			"/dataset/:corpusId/collocations/:word",
			handler.Collocations,
		)
		engine.GET(
			"/dataset/:corpusId/collocationsOfType/:type/:word/:pos",
			handler.CollocationsOfType,
		)
		engine.GET(
			"/dataset/:corpusId/collocationsOfType/:type/:word",
			handler.CollocationsOfType,
		)
		engine.GET(
			"/dataset/:corpusId/similarWords/:modelId",
			handler.HandleModelInfo,
		)

		engine.GET(
			"/dataset/:corpusId/similarWords/:modelId/:word/:fn",
			handler.WordSimilarity,
		)
		engine.GET(
			"/dataset/:corpusId/similarWords/:modelId/:word",
			handler.WordSimilarity,
		)
		engine.GET(
			"/dataset/:corpusId/similarWords",
			handler.HandleModelList,
		)

		srv := &http.Server{
			Handler:      engine,
			Addr:         fmt.Sprintf("%s:%d", conf.ListenAddress, conf.ListenPort),
			WriteTimeout: time.Duration(conf.ServerWriteTimeoutSecs) * time.Second,
			ReadTimeout:  time.Duration(conf.ServerReadTimeoutSecs) * time.Second,
		}

		go func() {
			err := srv.ListenAndServe()
			if err != nil {
				log.Error().Err(err).Send()
			}
		}()

		<-ctx.Done()
		log.Info().Err(err).Msg("Shutdown request error")

		ctxShutDown, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctxShutDown); err != nil {
			log.Fatal().Err(err).Msg("Server forced to shutdown")
		}

	} else {
		fmt.Printf("Word-Sim-Service %s\nbuild date: %s\nlast commit: %s\n",
			versionInfo.Version, versionInfo.BuildDate, versionInfo.GitCommit)
	}
}
