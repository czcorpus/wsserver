package queries

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/czcorpus/depreldb/scoll"
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
			if v.Value != word {
				continue
			}
			entries, err := db.GetLemmaDeprelValues(v.TokenID)
			if err != nil {
				return []ResultRow{}, core.NewAppError(
					"failed to get requested model",
					core.ErrorTypeInternalError,
					err,
				)
			}
			if len(entries) > 10 {
				entries = entries[:10]
			}
			for _, entry := range entries {
				syntaxFnMatches = append(syntaxFnMatches, entry.Value)
			}
		}

	} else {
		syntaxFnMatches = posIDs
	}

	ans := make([]ResultRow, 0, len(syntaxFnMatches)*limit)
	for _, posItem := range syntaxFnMatches {
		fmt.Println("POSITEM >> ", posItem)
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
	if len(ans) > limit {
		ans = ans[:limit]
	}
	return ans, core.AppError{}
}

func (wss *SearchProvider) Collocations(
	ctx context.Context,
	datasetID, word string,
	options ...func(opts *scoll.CalculationOptions),
) ([]SimpleCollocation, core.AppError) {

	if !wss.collDBs.Contains(datasetID) {
		return []SimpleCollocation{}, core.NewAppError(
			fmt.Sprintf("collocations dataset %s not found", datasetID),
			core.ErrorTypeNotFound,
			nil,
		)
	}
	db := wss.collDBs[datasetID]

	result, err := scoll.FromDatabase(db).GetCollocations(
		word,
		options...,
	)
	if err != nil {
		return []SimpleCollocation{}, core.NewAppError(
			fmt.Sprintf("collocations dataset %s not found", datasetID),
			core.ErrorTypeInternalError,
			err,
		)
	}

	ans := make([]SimpleCollocation, len(result))
	for i, v := range result {
		ans[i] = SimpleCollocation{
			SearchMatch: LemmaInfo{
				Value: v.Lemma.Value,
				PoS:   v.Lemma.PoS,
			},
			Collocate: LemmaInfo{
				Value: v.Collocate.Value,
				PoS:   v.Collocate.PoS,
			},
			Deprel:     v.Deprel,
			LogDice:    SafeFloat(math.Round(v.LogDice*100) / 100),
			TScore:     SafeFloat(math.Round(v.TScore*100) / 100),
			LMI:        SafeFloat(math.Round(v.LMI*100) / 100),
			LL:         SafeFloat(math.Round(v.LogLikelihood*100) / 100),
			RRF:        SafeFloat(math.Round(v.RRFScore*1000) / 1000),
			MutualDist: SafeFloat(v.MutualDist),
		}
	}

	return ans, core.AppError{}
}

type dictItem struct {
	Lemma    string    `json:"lemma"`
	PoS      string    `json:"pos"`
	Freq     SafeFloat `json:"freq"`
	TextType string    `json:"textType"`
}

func (wss *SearchProvider) Dictionary(datasetID, word string) ([]dictItem, core.AppError) {
	fmt.Println("DATASET: ", datasetID)
	db, ok := wss.collDBs[datasetID]
	if !ok {
		return []dictItem{}, core.NewAppError(
			fmt.Sprintf("unknown dataset: %s", datasetID), core.ErrorTypeNotFound, nil)
	}
	variants, err := db.GetLemmaIDsByPrefix(word)
	fmt.Println("VARIANRTS: ", variants)
	if err != nil {
		return []dictItem{}, core.NewAppError(
			"failed to get matching lemmas",
			core.ErrorTypeInternalError,
			err,
		)
	}
	ans := make([]dictItem, 0, 20)
	for _, v := range variants {
		if v.Value != word {
			continue
		}
		entries, err := db.GetMatchingLemmaProps(v.TokenID)
		if err != nil {
			return []dictItem{}, core.NewAppError(
				"failed to get requested model",
				core.ErrorTypeInternalError,
				err,
			)
		}
		for _, entry := range entries {
			ans = append(ans, dictItem{
				Lemma:    v.Value,
				PoS:      entry.Pos,
				Freq:     SafeFloat(entry.Freq),
				TextType: entry.TextType,
			})
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
