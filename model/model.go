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

package model

import (
	"errors"
	"os"

	"github.com/sajari/word2vec"
)

var (
	ErrModelNotFound     = errors.New("model not found")
	ErrModelConfNotFound = errors.New("model configuration not found")
)

type ModelInfo struct {
	Name        string `json:"name"`
	Size        int    `json:"size"`
	Description string `json:"description"`
	Error       string `json:"error"`
}

// isFile tests whether a provided path represents
// a file. If not or in case of an IO error,
// false is returned.
func isFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	finfo, err := f.Stat()
	if err != nil {
		return false
	}
	return finfo.Mode().IsRegular()
}

//  -------------------------------

// Provider is a wrapper around word2vec with support
// for multiple models. Please be aware though that each
// model is typically quite memory consuming.
type Provider struct {
	dataDir string
	models  map[string]*word2vec.Model
	configs []ModelConf
}

func (m *Provider) FindModel(corpusName string, modelName string) (*ModelConf, error) {
	for _, mc := range m.configs {
		if mc.Corpname == corpusName && mc.ID == modelName {
			return &mc, nil
		}
	}
	return nil, ErrModelConfNotFound
}

func (m *Provider) access(conf *ModelConf) (*word2vec.Model, error) {
	_, ok := m.models[conf.ModelKey()]
	if !ok {
		dataPath := conf.MkDataPath(m.dataDir)
		if !isFile(dataPath) {
			return nil, ErrModelNotFound
		}
		f, err := os.Open(dataPath)
		if err != nil {
			return nil, err
		}
		model, err := word2vec.FromReader(f)
		if err != nil {
			return nil, err
		}
		m.models[conf.ModelKey()] = model
	}
	return m.models[conf.ModelKey()], nil
}

func (m *Provider) Query(conf *ModelConf, word, pos string, limit int) ([]word2vec.Match, error) {
	model, err := m.access(conf)
	if err != nil {
		return nil, err
	}
	expr := word2vec.Expr{}
	if conf.ContainsPoS {
		expr.Add(1, word+"_"+pos)

	} else {
		expr.Add(1, word)
	}
	return model.CosN(expr, limit+1)
}

func (m *Provider) ListModels(corpname string) ([]ModelInfo, error) {

	ans := make([]ModelInfo, 0, len(m.configs))
	for _, modelConf := range m.configs {
		if modelConf.Corpname != corpname {
			continue
		}
		model, err := m.access(&modelConf)
		info := ModelInfo{
			Name:        modelConf.ID,
			Description: modelConf.Description,
		}

		if err != nil {
			info.Error = err.Error()
			info.Size = 0

		} else {
			info.Size = model.Size()
		}
		ans = append(ans, info)
	}
	return ans, nil
}

// NewProvider is a recommended factory function for Provider
func NewProvider(dataDir string, configs []ModelConf) *Provider {
	return &Provider{
		dataDir: dataDir,
		models:  make(map[string]*word2vec.Model),
		configs: configs,
	}
}
