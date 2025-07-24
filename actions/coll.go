package actions

import (
	"fmt"
	"net/http"

	"github.com/czcorpus/cnc-gokit/unireq"
	"github.com/czcorpus/cnc-gokit/uniresp"
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
	syntaxFn := ctx.Param("fn")

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

	result, err := a.searcher.Collocations(ctx, corpusID, syntaxFn, word, limit, int(minScore))
	if !err.IsZero() {
		uniresp.RespondWithErrorJSON(ctx, err, mapError(err))
		return
	}

	/*
		ans := collGroupedResponse{
			Matches: make([]lemmaCollocates, 0, len(result)),
		}
			merged := collections.NewMultidict[storage.Collocation]()
			for _, v := range result {
				merged.Add(v.RawLemma, v)
			}
			for _, v := range merged.Iterate {
				lemma, lemmaFn := v[0].LemmaAndFn() // len(v) is always > 0
				colls := lemmaCollocates{
					Lemma:      lemma,
					SyntaxFn:   lemmaFn,
					Collocates: make([]collocate, 0, limit),
				}
				for _, coll := range v {
					lemma, lemmaFn := coll.CollocateAndFn()
					colls.Collocates = append(
						colls.Collocates,
						collocate{
							Lemma:    lemma,
							SyntaxFn: lemmaFn,
							Score:    float32(coll.LogDice),
						},
					)
				}
				ans.Matches = append(ans.Matches, colls)
			}
	*/

	uniresp.WriteJSONResponse(ctx.Writer, result)
}
