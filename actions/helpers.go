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
	"net/http"
	"strings"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/sajari/word2vec"
)

func isNotFound(err error) bool {
	_, ok := err.(*word2vec.NotFoundError)
	return ok
}

func getClientAddress(req *http.Request) string {
	ans := req.Header.Get("X-Forwarded-For")
	if ans == "" {
		ans = req.RemoteAddr
	}
	return strings.Split(ans, ":")[0]
}

func splitByLastUnderscore(s string) (string, string) {
	lastIndex := strings.LastIndex(s, "_")
	if lastIndex == -1 {
		return s, ""
	}
	return s[:lastIndex], s[lastIndex+1:]
}

func mergeByFunc(data []ResultRow) []ResultRow {
	merged := collections.NewMultidict[ResultRow]()
	ans := make([]ResultRow, 0, len(data))
	for _, item := range data {
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
