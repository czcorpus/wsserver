// Copyright 2025 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2025 Institute of the Czech National Corpus,
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

package corpora

type Keyword string

type CitationInfo struct {
	DefaultRef        string   `json:"default_ref"`
	ArticleRef        []string `json:"article_ref"`
	OtherBibliography string   `json:"other_bibliography"`
}

type Info struct {
	Corpname     string       `json:"corpname"`
	Size         int          `json:"size"`
	Description  string       `json:"description"`
	WebURL       string       `json:"webUrl"`
	CitationInfo CitationInfo `json:"citationInfo"`
	SrchKeywords []Keyword    `json:"srchKeywords"`
}

type corpusData struct {
	Data Info `json:"data"`
}

type infoResponse struct {
	Corpus corpusData `json:"corpus"`
	Locale string     `json:"locale"`
}

func NewInfoResponse(info Info, locale string) infoResponse {
	return infoResponse{
		Corpus: corpusData{
			Data: info,
		},
		Locale: locale,
	}
}
