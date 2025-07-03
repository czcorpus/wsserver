package actions

import (
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/czcorpus/cnc-gokit/unireq"
	"github.com/czcorpus/cnc-gokit/uniresp"
	"github.com/czcorpus/wsserver/model"
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

type lemmaInfo struct {
	Value         string `json:"value"`
	SyntacticFunc string `json:"syntacticFunc"`
}

type simpleCollocation struct {
	SearchMatch lemmaInfo `json:"searchMatch"`
	Collocate   lemmaInfo `json:"collocate"`
	LogDice     float64   `json:"logDice"`
	TScore      float64   `json:"tscore"`
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
	/*
		minScore, ok := unireq.GetURLFloatArgOrFail(ctx, "minScore", 0)
		if !ok {
			uniresp.RespondWithErrorJSON(ctx, fmt.Errorf("invalid value of 'limit'"), http.StatusUnprocessableEntity)
			return
		}
	*/

	if !a.collDBs.Contains(corpusID) {
		uniresp.RespondWithErrorJSON(ctx, fmt.Errorf("dataset %s not found", corpusID), http.StatusNotFound)
		return
	}
	modelInfo, err := a.modelProvider.FindModel(corpusID, "nce") // TODO
	var errNotFound model.ModelNotFoundError
	if errors.As(err, &errNotFound) {
		uniresp.RespondWithErrorJSON(ctx, err, http.StatusNotFound)
		return

	} else if err != nil {
		uniresp.RespondWithErrorJSON(ctx, err, http.StatusInternalServerError)
	}
	db := a.collDBs[corpusID]
	if syntaxFn != "" {
		word = word + "-" + syntaxFn
	}
	result, err := db.CalculateMeasures(word, int(modelInfo.CorpusSize), limit, "tscore") // TODO
	if err != nil {
		uniresp.RespondWithErrorJSON(ctx, err, http.StatusInternalServerError)
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
	ans := make([]simpleCollocation, len(result))
	for i, v := range result {
		lm, lmf := v.LemmaAndFn()
		col, colf := v.CollocateAndFn()
		ans[i] = simpleCollocation{
			SearchMatch: lemmaInfo{
				Value:         lm,
				SyntacticFunc: lmf,
			},
			Collocate: lemmaInfo{
				Value:         col,
				SyntacticFunc: colf,
			},
			LogDice: math.Round(v.LogDice*100) / 100,
			TScore:  math.Round(v.TScore*100) / 100,
		}
	}
	uniresp.WriteJSONResponse(ctx.Writer, ans)
}
