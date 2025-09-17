package actions

import (
	"fmt"
	"net/http"

	"github.com/czcorpus/cnc-gokit/unireq"
	"github.com/czcorpus/cnc-gokit/uniresp"
	"github.com/czcorpus/depreldb/scoll"
	"github.com/czcorpus/wsserver/queries"
	"github.com/gin-gonic/gin"
)

type collocate struct {
	Lemma    string  `json:"lemma"`
	SyntaxFn string  `json:"syntaxFn"`
	Score    float64 `json:"score"`
}

type lemmaCollocates struct {
	Lemma      string      `json:"lemma"`
	SyntaxFn   string      `json:"syntaxFn"`
	Collocates []collocate `json:"collocates"`
}

type collGroupedResponse struct {
	Matches []lemmaCollocates `json:"matches"`
	Error   error             `json:"error,omitempty"`
}

type collResponse struct {
	Items []queries.SimpleCollocation `json:"items"`
	Error error                       `json:"error,omitempty"`
}

func (a *ActionHandler) Collocations(ctx *gin.Context) {

	corpusID := ctx.Param("corpusId")
	word := ctx.Param("word")
	pos := ctx.Param("pos")
	tt := ctx.Query("tt")

	limit, ok := unireq.GetURLIntArgOrFail(ctx, "limit", 10)
	if !ok {
		uniresp.RespondWithErrorJSON(ctx, fmt.Errorf("invalid value of 'limit'"), http.StatusUnprocessableEntity)
		return
	}

	result, err := a.searcher.Collocations(
		ctx,
		corpusID,
		word,
		scoll.WithPoS(pos),
		scoll.WithLimit(limit),
		scoll.WithSortBy("rrf"),
		scoll.WithTextType(tt),
	)
	if !err.IsZero() {
		uniresp.RespondWithErrorJSON(ctx, err, mapError(err))
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, collResponse{Items: result})
}

func (a *ActionHandler) CollocationsOfType(ctx *gin.Context) {

	corpusID := ctx.Param("corpusId")
	word := ctx.Param("word")
	collType := ctx.Param("type")
	tt := ctx.Query("tt")

	limit, ok := unireq.GetURLIntArgOrFail(ctx, "limit", 10)
	if !ok {
		uniresp.RespondWithErrorJSON(ctx, fmt.Errorf("invalid value of 'limit'"), http.StatusUnprocessableEntity)
		return
	}

	result, err := a.searcher.Collocations(
		ctx,
		corpusID,
		word,
		scoll.WithLimit(limit),
		scoll.WithSortBy("rrf"),
		scoll.WithTextType(tt),
		scoll.WithPredefinedSearch(scoll.PredefinedSearch(collType)),
		scoll.WithMaxAvgCollocateDist(1.499),
	)
	if !err.IsZero() {
		uniresp.RespondWithErrorJSON(ctx, err, mapError(err))
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, collResponse{Items: result})
}
