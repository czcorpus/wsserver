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

package actions

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/czcorpus/scollector/storage"

	"github.com/czcorpus/cnc-gokit/unireq"
	"github.com/czcorpus/cnc-gokit/uniresp"
	"github.com/gin-gonic/gin"
	"github.com/sajari/word2vec"
)

var (
	posIDs = []string{"N", "A", "P", "C", "V", "D", "R", "J", "T", "I", "Z", "X"}
)

// ResultRow represents a single result item for "similar words"
type ResultRow struct {
	Word     string   `json:"word"`
	SyntaxFn []string `json:"syntaxFn"`
	Score    float32  `json:"score"`
}

// ----

type CollDBMap map[string]*storage.DB

func (dbmap CollDBMap) Contains(key string) bool {
	_, ok := dbmap[key]
	return ok
}

// ----

func exportResult(matches []word2vec.Match, minScore float32) []ResultRow {
	ans := make([]ResultRow, 0, len(matches))
	if len(matches) < 2 {
		return ans
	}
	for _, v := range matches {
		if v.Score >= minScore {
			var pos string
			word := v.Word
			witems := strings.Split(v.Word, "_")
			if len(witems) > 1 {
				pos = witems[len(witems)-1]
				word = witems[len(witems)-2]
			}
			ans = append(ans, ResultRow{Word: word, SyntaxFn: []string{pos}, Score: v.Score})
		}
	}
	return ans
}

// WordSimilarity handles search actions for similar words
func (a *ActionHandler) WordSimilarity(ctx *gin.Context) {
	corpusID := ctx.Param("corpusId")
	modelID := ctx.Param("modelId")

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

	modelConf, err := a.modelProvider.FindModel(corpusID, modelID)
	if err != nil {
		uniresp.RespondWithErrorJSON(ctx, err, http.StatusInternalServerError)
		return
	}

	word := ctx.Param("word")
	syntaxFn := ctx.Param("fn")
	var syntaxFnMatches []string

	if !modelConf.ContainsPoS {
		if syntaxFn != "" {
			uniresp.RespondWithErrorJSON(
				ctx, fmt.Errorf("The model does not support setting PoS"), http.StatusBadRequest)
			return
		}
		syntaxFnMatches = []string{""}

	} else if syntaxFn != "" {
		syntaxFnMatches = []string{syntaxFn}

	} else if a.collDBs.Contains(corpusID) {
		db := a.collDBs[corpusID]
		variants, err := db.GetLemmaIDsByPrefix(word)
		if err != nil {
			uniresp.RespondWithErrorJSON(
				ctx,
				err,
				http.StatusInternalServerError,
			)
			return
		}
		for _, v := range variants {
			_, deprel := splitByLastUnderscore(v.Value)
			syntaxFnMatches = append(syntaxFnMatches, strings.TrimSpace(deprel)) // TODO workaround, should be solved in the lib
		}

	} else {
		syntaxFnMatches = posIDs
	}

	ans := make([]ResultRow, 0, len(syntaxFnMatches)*limit)
	var lastNotFoundErr error
	for _, posItem := range syntaxFnMatches {
		matches, err := a.modelProvider.Query(modelConf, word, posItem, limit+1)
		if err != nil {
			if isNotFound(err) {
				lastNotFoundErr = err

			} else {
				uniresp.RespondWithErrorJSON(ctx, err, http.StatusInternalServerError)
				return
			}
		}
		ans = append(ans, exportResult(matches, float32(minScore))...)
	}
	if lastNotFoundErr != nil && len(ans) == 0 {
		uniresp.RespondWithErrorJSON(
			ctx, lastNotFoundErr, http.StatusNotFound)
		return
	}
	ans = mergeByFunc(ans, word)
	sort.Slice(ans, func(i, j int) bool {
		return ans[i].Score > ans[j].Score
	})
	uniresp.WriteJSONResponse(ctx.Writer, ans[:limit])
}
