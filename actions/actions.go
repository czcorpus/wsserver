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

	"github.com/czcorpus/cnc-gokit/unireq"
	"github.com/czcorpus/cnc-gokit/uniresp"
	"github.com/czcorpus/wsserver/model"
	"github.com/czcorpus/wsserver/queries"
	"github.com/gin-gonic/gin"
)

// ----

// ActionHandler wraps all the HTTP actions of word-sim-service
type ActionHandler struct {
	models   queries.W2VModelProvider
	searcher *queries.SearchProvider
}

// HandleModelList provides listing of all the configured w2v models for a specified corpus
func (a *ActionHandler) HandleModelList(ctx *gin.Context) {
	corpusID := ctx.Param("corpusId")
	ans, err := a.models.ListModels(corpusID)
	if err != nil {
		uniresp.RespondWithErrorJSON(ctx, err, http.StatusInternalServerError)
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, ans)
}

func (a *ActionHandler) HandleModelInfo(ctx *gin.Context) {
	corpusID := ctx.Param("corpusId")
	modelID := ctx.Param("modelId")
	info, err := a.models.FindModel(corpusID, modelID)
	if err == model.ErrModelConfNotFound {
		uniresp.RespondWithErrorJSON(
			ctx,
			err,
			http.StatusNotFound,
		)
		return

	} else if err != nil {
		uniresp.RespondWithErrorJSON(
			ctx,
			err,
			http.StatusInternalServerError,
		)
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, info)
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
	word := ctx.Param("word")
	posOrSfn := ctx.Param("fn")

	res, err := a.searcher.SimilarlyUsedWords(
		ctx, corpusID, modelID, posOrSfn, word, limit, float32(minScore),
	)
	if !err.IsZero() {
		uniresp.RespondWithErrorJSON(
			ctx, err, mapError(err),
		)
	}
	uniresp.WriteJSONResponse(ctx.Writer, res)
}

func (a *ActionHandler) CollsByDepType(ctx *gin.Context) {
	/*
			TODO:
		ModifiersOf
		// [p_lemma="team" & deprel="nmod" & upos="NOUN"]
		fx, err := cdb.GetFreq("", "NOUN", w.V, w.PoS, "nmod")


		NounsModifiedBy
		// [lemma="team" & deprel="nmod" & p_upos="NOUN"]
		fx, err := cdb.GetFreq(w.V, w.PoS, "", "NOUN", "nmod")


		VerbsSubject
		// [lemma="team" & deprel="nsubj" & p_upos="VERB"]
		fx, err := cdb.GetFreq(w.V, w.PoS, "", "VERB", "nsubj")


		VerbsObject
		// [lemma="team" & deprel="obj|iobj" & p_upos="VERB"]
		fx, err := cdb.GetFreq(w.V, w.PoS, "", "VERB", "obj|iobj")

	*/
}

// NewActionHandler is a recommended factory function for creating ActionHandler instance
func NewActionHandler(
	dataDir string,
	models queries.W2VModelProvider,
	searcher *queries.SearchProvider,
) (*ActionHandler, error) {

	return &ActionHandler{
		models:   models,
		searcher: searcher,
	}, nil
}
