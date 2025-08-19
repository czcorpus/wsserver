package queries

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/czcorpus/scollector/scoll"
	"github.com/czcorpus/wsserver/core"
	"github.com/czcorpus/wsserver/model"
	"github.com/sajari/word2vec"
)

func isNotFound(err error) bool {
	_, ok := err.(*word2vec.NotFoundError)
	return ok
}

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

// ---------

type SearchProvider struct {
	collDBs       CollDBMap
	modelProvider W2VModelProvider
}

func (wss *SearchProvider) SimilarlyUsedWords(
	ctx context.Context,
	datasetID, modelID, posOrSfn, word string,
	limit int,
	minScore float32,
) ([]ResultRow, core.AppError) {

	var syntaxFnMatches []string

	modelConf, err := wss.modelProvider.FindModel(datasetID, modelID)
	if err == model.ErrModelConfNotFound || err == model.ErrModelNotFound {
		return []ResultRow{}, core.NewAppError(
			"failed to get requested model",
			core.ErrorTypeNotFound,
			err,
		)
	}
	if err != nil {
		return []ResultRow{}, core.NewAppError(
			"failed to get requested model",
			core.ErrorTypeInternalError,
			err,
		)
	}

	if !modelConf.ContainsPoS {
		if posOrSfn != "" {
			return []ResultRow{}, core.NewAppError(
				"The model does not support setting PoS",
				core.ErrorTypeInvalidArguments,
				nil,
			)
		}
		syntaxFnMatches = []string{""}

	} else if posOrSfn != "" {
		syntaxFnMatches = []string{posOrSfn}

	} else if wss.collDBs.Contains(datasetID) {
		db := wss.collDBs[datasetID]
		variants, err := db.GetLemmaIDsByPrefix(word)
		if err != nil {
			return []ResultRow{}, core.NewAppError(
				"failed to get matching variants",
				core.ErrorTypeInternalError,
				err,
			)
		}
		for _, v := range variants {
			_, deprel := splitByLastUnderscore(v.Value)
			syntaxFnMatches = append(syntaxFnMatches, strings.TrimSpace(deprel)) // TODO workaround, should be solved in the lib
		}

	} else {
		syntaxFnMatches = posIDs
	}

	ans := make([]ResultRow, 0, len(syntaxFnMatches)*limit)
	for _, posItem := range syntaxFnMatches {
		matches, err := wss.modelProvider.Query(modelConf, word, posItem, limit+1)
		if err != nil && !isNotFound(err) {
			return []ResultRow{}, core.NewAppError(
				"problem evaluation word similarity query",
				core.ErrorTypeInternalError,
				err,
			)
		}
		ans = append(ans, exportResult(matches, minScore)...)
	}
	ans = mergeByFunc(ans, word)
	sort.Slice(ans, func(i, j int) bool {
		return ans[i].Score > ans[j].Score
	})
	return ans, core.AppError{}
}

func (wss *SearchProvider) CollocationsOfType(
	ctx context.Context,
	datasetID, word, pos string,
	collType scoll.PredefinedSearch,
	limit, minScore int,
) ([]simpleCollocation, core.AppError) {

	if collType != scoll.ModifiersOf && collType != scoll.NounsModifiedBy &&
		collType != scoll.VerbsSubject && collType != scoll.VerbsObject {

		return []simpleCollocation{}, core.NewAppError("unsupported collocation type", core.ErrorTypeInvalidArguments, nil)
	}

	if !wss.collDBs.Contains(datasetID) {
		return []simpleCollocation{}, core.NewAppError(
			fmt.Sprintf("collocations dataset %s not found", datasetID),
			core.ErrorTypeNotFound,
			nil,
		)
	}
	db := wss.collDBs[datasetID]

	result, err := scoll.FromDatabase(db).GetCollocations(
		word,
		scoll.WithLimit(limit),
		scoll.WithSortBy("rrf"),
		scoll.WithPredefinedSearch(collType),
		scoll.WithMaxAvgCollocateDist(1.499),
	)
	if err != nil {
		return []simpleCollocation{}, core.NewAppError(
			fmt.Sprintf("collocations dataset %s not found", datasetID),
			core.ErrorTypeInternalError,
			err,
		)
	}

	ans := make([]simpleCollocation, len(result))
	for i, v := range result {
		ans[i] = simpleCollocation{
			SearchMatch: lemmaInfo{
				Value: v.Lemma.Value,
				PoS:   v.Lemma.PoS,
			},
			Collocate: lemmaInfo{
				Value: v.Collocate.Value,
				PoS:   v.Collocate.PoS,
			},
			Deprel:     v.Deprel,
			LogDice:    math.Round(v.LogDice*100) / 100,
			TScore:     math.Round(v.TScore*100) / 100,
			LMI:        math.Round(v.LMI*100) / 100,
			RRF:        math.Round(v.RRFScore*1000) / 1000,
			MutualDist: v.MutualDist,
		}
	}

	return ans, core.AppError{}
}

func (wss *SearchProvider) Collocations(
	ctx context.Context,
	datasetID, word, pos string,
	limit, minScore int,
) ([]simpleCollocation, core.AppError) {

	if !wss.collDBs.Contains(datasetID) {
		return []simpleCollocation{}, core.NewAppError(
			fmt.Sprintf("collocations dataset %s not found", datasetID),
			core.ErrorTypeNotFound,
			nil,
		)
	}
	db := wss.collDBs[datasetID]

	result, err := scoll.FromDatabase(db).GetCollocations(
		word,
		scoll.WithPoS(pos),
		scoll.WithLimit(limit),
		scoll.WithSortBy("rrf"),
	)
	if err != nil {
		return []simpleCollocation{}, core.NewAppError(
			fmt.Sprintf("collocations dataset %s not found", datasetID),
			core.ErrorTypeInternalError,
			err,
		)
	}

	ans := make([]simpleCollocation, len(result))
	for i, v := range result {
		ans[i] = simpleCollocation{
			SearchMatch: lemmaInfo{
				Value: v.Lemma.Value,
				PoS:   v.Lemma.PoS,
			},
			Collocate: lemmaInfo{
				Value: v.Collocate.Value,
				PoS:   v.Collocate.PoS,
			},
			Deprel:     v.Deprel,
			LogDice:    math.Round(v.LogDice*100) / 100,
			TScore:     math.Round(v.TScore*100) / 100,
			LMI:        math.Round(v.LMI*100) / 100,
			RRF:        math.Round(v.RRFScore*1000) / 1000,
			MutualDist: v.MutualDist,
		}
	}

	return ans, core.AppError{}
}

func NewSearchProvider(
	dataDir string,
	collDbs CollDBMap,
	w2vModels W2VModelProvider,
) (*SearchProvider, error) {

	return &SearchProvider{
		collDBs:       collDbs,
		modelProvider: w2vModels,
	}, nil
}
