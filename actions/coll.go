package actions

import (
	"fmt"
	"net/http"

	"github.com/czcorpus/cnc-gokit/unireq"
	"github.com/czcorpus/cnc-gokit/uniresp"
	"github.com/czcorpus/scollector/scoll"
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

func (a *ActionHandler) Collocations(ctx *gin.Context) {

	corpusID := ctx.Param("corpusId")
	word := ctx.Param("word")
	pos := ctx.Param("pos")

	limit, ok := unireq.GetURLIntArgOrFail(ctx, "limit", 10)
	if !ok {
		uniresp.RespondWithErrorJSON(ctx, fmt.Errorf("invalid value of 'limit'"), http.StatusUnprocessableEntity)
		return
	}
	minScore, ok := unireq.GetURLFloatArgOrFail(ctx, "minScore", 0)
	if !ok {
		uniresp.RespondWithErrorJSON(ctx, fmt.Errorf("invalid value of 'limit'"), http.StatusUnprocessableEntity)
		return
	}

	result, err := a.searcher.Collocations(ctx, corpusID, word, pos, limit, int(minScore))
	if !err.IsZero() {
		uniresp.RespondWithErrorJSON(ctx, err, mapError(err))
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, result)
}

func (a *ActionHandler) CollocationsOfType(ctx *gin.Context) {

	corpusID := ctx.Param("corpusId")
	word := ctx.Param("word")
	pos := ctx.Param("pos")
	collType := ctx.Param("type")

	limit, ok := unireq.GetURLIntArgOrFail(ctx, "limit", 10)
	if !ok {
		uniresp.RespondWithErrorJSON(ctx, fmt.Errorf("invalid value of 'limit'"), http.StatusUnprocessableEntity)
		return
	}
	minScore, ok := unireq.GetURLFloatArgOrFail(ctx, "minScore", 0)
	if !ok {
		uniresp.RespondWithErrorJSON(ctx, fmt.Errorf("invalid value of 'limit'"), http.StatusUnprocessableEntity)
		return
	}

	result, err := a.searcher.CollocationsOfType(
		ctx, corpusID, word, pos, scoll.PredefinedSearch(collType), limit, int(minScore))
	if !err.IsZero() {
		uniresp.RespondWithErrorJSON(ctx, err, mapError(err))
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, result)

}
