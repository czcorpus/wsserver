package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/czcorpus/wsserver/core"
	"github.com/czcorpus/wsserver/queries"
)

type HTTPCLient struct {
}

func (c *HTTPCLient) GET(reqURL string, args map[string]string, headers map[string]string) (string, error) {
	parsedURL, err := url.Parse(reqURL)
	if err != nil {
		return "", err
	}
	params := url.Values{}
	for key, value := range args {
		params.Add(key, value)
	}
	parsedURL.RawQuery = params.Encode()
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return "", err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return "", nil
}

func (c *HTTPCLient) POST(path string, body string) (string, error) {
	// Implement the POST request logic here
	return "", nil
}

func NewHTTPClient() *HTTPCLient {
	return &HTTPCLient{}
}

// --------

type HTTPClientSearcher struct {
	client  *HTTPCLient
	baseURL string
}

func (searcher *HTTPClientSearcher) SimilarlyUsedWords(
	ctx context.Context,
	datasetID, modelID, posOrSfn, word string,
	limit int,
	minScore float32,
) ([]queries.ResultRow, core.AppError) {
	var path string
	if posOrSfn != "" {
		path = fmt.Sprintf("/dataset/%s/similarWords/%s/%s/%s", datasetID, modelID, word, posOrSfn)

	} else {
		path = fmt.Sprintf("/dataset/%s/similarWords/%s/%s", datasetID, modelID, word)
	}
	baseURL, err := url.JoinPath(searcher.baseURL, path)
	if err != nil {
		return []queries.ResultRow{}, core.NewAppError("failed to create API URL", core.ErrorTypeInternalError, err)
	}
	args := map[string]string{
		"minScore": fmt.Sprintf("%01.2f", minScore),
		"limit":    strconv.Itoa(limit),
	}
	resp, err := searcher.client.GET(baseURL, args, map[string]string{})
	ans := make([]queries.ResultRow, 0, min(50, limit))

	return nil, core.AppError{}
}

func NewHTTPClientSearcher() *HTTPClientSearcher {
	return &HTTPClientSearcher{}
}
