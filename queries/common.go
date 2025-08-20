package queries

import (
	"fmt"
	"strings"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/czcorpus/scollector/storage"
	"github.com/czcorpus/wsserver/model"
	"github.com/sajari/word2vec"
)

var (
	posIDs = []string{"N", "A", "P", "C", "V", "D", "R", "J", "T", "I", "Z", "X"}
)

type W2VModelProvider interface {
	FindModel(corpusName string, modelName string) (*model.ModelConf, error)
	Query(conf *model.ModelConf, word, pos string, limit int) ([]word2vec.Match, error)
	ListModels(corpname string) ([]model.ModelInfo, error)
}

// ResultRow represents a single result item for "similar words"
type ResultRow struct {
	Word     string   `json:"word"`
	SyntaxFn []string `json:"syntaxFn"`
	Score    float32  `json:"score"`
}

// --------------------------------

type CollDBMap map[string]*storage.DB

func (dbmap CollDBMap) Contains(key string) bool {
	_, ok := dbmap[key]
	return ok
}

func NewCollDbMap(modelConfigs []model.ModelConf) (CollDBMap, error) {
	collDbs := make(CollDBMap)
	for _, conf := range modelConfigs {
		if conf.SyntaxDatabasePath != "" {
			db, err := storage.OpenDB(conf.SyntaxDatabasePath)
			if err != nil {
				return collDbs, fmt.Errorf("failed to instantiate coll database for %s: %w", conf.Corpname, err)
			}
			collDbs[conf.Corpname] = db
		}
	}
	return collDbs, nil
}

func splitByLastUnderscore(s string) (string, string) {
	lastIndex := strings.LastIndex(s, "_")
	if lastIndex == -1 {
		return s, ""
	}
	return s[:lastIndex], s[lastIndex+1:]
}

func mergeByFunc(data []ResultRow, srchWord string) []ResultRow {
	merged := collections.NewMultidict[ResultRow]()
	ans := make([]ResultRow, 0, len(data))
	for _, item := range data {
		if item.Word == srchWord {
			continue
		}
		merged.Add(item.Word, item)
	}
	for k, v := range merged.Iterate {
		newItem := ResultRow{
			Word: k,
		}
		var avg float32
		for _, v2 := range v {
			newItem.SyntaxFn = append(newItem.SyntaxFn, v2.SyntaxFn...)
			avg += v2.Score
		}
		newItem.Score = avg
		ans = append(ans, newItem)
	}
	return ans
}

// ----------------------------

type LemmaInfo struct {
	Value string `json:"value"`
	PoS   string `json:"pos"`
}

type SimpleCollocation struct {
	SearchMatch LemmaInfo `json:"searchMatch"`
	Collocate   LemmaInfo `json:"collocate"`
	Deprel      string    `json:"deprel"`
	LogDice     float64   `json:"logDice"`
	TScore      float64   `json:"tscore"`
	LMI         float64   `json:"lmi"`
	RRF         float64   `json:"rrf"`
	MutualDist  float64   `json:"mutualDist"`
}
