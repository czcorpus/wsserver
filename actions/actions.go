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

	"github.com/czcorpus/cnc-gokit/uniresp"
	"github.com/czcorpus/scollector/storage"
	"github.com/czcorpus/wsserver/corpora"
	"github.com/czcorpus/wsserver/model"
	"github.com/gin-gonic/gin"
)

// ----

// ActionHandler wraps all the HTTP actions of word-sim-service
type ActionHandler struct {
	modelProvider *model.Provider
	corpora       map[string]corpora.Info
	collDBs       CollDBMap
}

// HandleModelList provides listing of all the configured w2v models for a specified corpus
func (a *ActionHandler) HandleModelList(ctx *gin.Context) {
	corpusID := ctx.Param("corpusId")
	ans, err := a.modelProvider.ListModels(corpusID)
	if err != nil {
		uniresp.RespondWithErrorJSON(ctx, err, http.StatusInternalServerError)
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, ans)
}

// HandleCorporaList provides list of all the configured corpora
func (a *ActionHandler) HandleCorporaList(ctx *gin.Context) {
	ans, err := a.modelProvider.ListCorpora()
	if err != nil {
		uniresp.RespondWithErrorJSON(ctx, err, http.StatusInternalServerError)
		return
	}
	uniresp.WriteJSONResponse(ctx.Writer, ans)
}

func (a *ActionHandler) HandleModelInfo(ctx *gin.Context) {
	corpusID := ctx.Param("corpusId")
	info, ok := a.corpora[corpusID]
	if !ok {
		uniresp.RespondWithErrorJSON(
			ctx,
			fmt.Errorf("corpus not found"),
			http.StatusNotFound,
		)
		return
	}
	resp := corpora.NewInfoResponse(info, "cs-CZ") // TOOD locale
	uniresp.WriteJSONResponse(ctx.Writer, resp)
}

// NewActionHandler is a recommended factory function for creating ActionHandler instance
func NewActionHandler(
	dataDir string,
	models []model.ModelConf,
	corpora map[string]corpora.Info,
) (*ActionHandler, error) {
	collDbs := make(CollDBMap)
	for _, conf := range models {
		if conf.SyntaxDatabasePath != "" {
			db, err := storage.OpenDB(conf.SyntaxDatabasePath)
			if err != nil {
				return nil, fmt.Errorf("failed to instantiate coll database for %s: %w", conf.Corpname, err)
			}
			collDbs[conf.Corpname] = db
		}
	}
	return &ActionHandler{
		modelProvider: model.NewProvider(dataDir, models),
		corpora:       corpora,
		collDBs:       collDbs,
	}, nil
}
