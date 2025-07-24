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

package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/czcorpus/cnc-gokit/logging"
	"github.com/czcorpus/wsserver/corpora"
	"github.com/czcorpus/wsserver/model"
	"github.com/rs/zerolog/log"
)

const (
	dfltServerWriteTimeoutSecs = 10
	dfltServerReadTimeoutSecs  = 10
)

// VersionInfo provides a detailed information about the actual build
type VersionInfo struct {
	Version   string `json:"version"`
	BuildDate string `json:"buildDate"`
	GitCommit string `json:"gitCommit"`
}

type MCPConfig struct {
	SelfContained bool `json:"selfContained"`
}

type Config struct {
	ListenAddress          string                  `json:"listenAddress"`
	ListenPort             int                     `json:"listenPort"`
	ServerWriteTimeoutSecs int                     `json:"serverWriteTimeoutSecs"`
	ServerReadTimeoutSecs  int                     `json:"serverReadTimeoutSecs"`
	DataDir                string                  `json:"dataDir"`
	Models                 []model.ModelConf       `json:"models"`
	Corpora                map[string]corpora.Info `json:"corpora"`
	Logging                logging.LoggingConf     `json:"logging"`
	MCP                    MCPConfig               `json:"mcp"`
}

func ApplyDefaults(conf *Config) {
	if conf.ServerWriteTimeoutSecs == 0 {
		conf.ServerWriteTimeoutSecs = dfltServerWriteTimeoutSecs
		log.Warn().Msgf(
			"serverWriteTimeoutSecs not specified, using default: %d",
			dfltServerWriteTimeoutSecs,
		)
	}
	if conf.ServerReadTimeoutSecs == 0 {
		conf.ServerReadTimeoutSecs = dfltServerReadTimeoutSecs
		log.Warn().Msgf(
			"serverReadTimeoutSecs not specified, using default: %d",
			dfltServerReadTimeoutSecs,
		)
	}

}

func Load(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("Config path not specified")
	}
	rawData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var conf Config
	json.Unmarshal(rawData, &conf)
	return &conf, nil
}
