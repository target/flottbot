// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/utils"
)

// HTTPReq handles 'http' actions for rules.
func HTTPReq(args models.Action, msg *models.Message) (*models.HTTPResponse, error) {
	log.Info().Msgf("executing http request for action %#q", args.Name)

	if args.Timeout == 0 {
		// Default HTTP Timeout of 10 seconds
		args.Timeout = 10
	}

	client := &http.Client{
		Timeout: time.Duration(args.Timeout) * time.Second,
	}

	// check the URL string from defined action has a variable, try to substitute it
	url, err := utils.Substitute(args.URL, msg.Vars)
	if err != nil {
		log.Error().Msg("failed substituting variables in url parameter")
		return nil, err
	}

	// TODO: refactor querydata
	// this is a temp fix for scenarios where
	// substitution above may have introduced spaces in the URL
	url = strings.ReplaceAll(url, " ", "%20")

	url, payload, err := prepRequestData(url, args.Type, args.QueryData, msg)
	if err != nil {
		log.Error().Msg("failed preparing the request data for the http request")
		return nil, err
	}

	req, err := http.NewRequest(args.Type, url, payload)
	if err != nil {
		log.Error().Msg("failed to create a new http request")
		return nil, err
	}

	req.Close = true

	// Add custom headers to request
	for k, v := range args.CustomHeaders {
		value, err := utils.Substitute(v, msg.Vars)
		if err != nil {
			log.Error().Msg("failed substituting variables in custom headers")
			return nil, err
		}

		req.Header.Add(k, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error().Msg("failed to execute the http request")
		return nil, err
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Msg("failed to read response from http request")
		return nil, err
	}

	fields, err := extractFields(bodyBytes)
	if err != nil {
		log.Error().Msg("failed to extract the fields from the http response")
		return nil, err
	}

	result := models.HTTPResponse{
		Status: resp.StatusCode,
		Raw:    string(bodyBytes),
		Data:   fields,
	}

	log.Info().Msgf("http request for action %#q completed", args.Name)

	return &result, nil
}

// Depending on the type of request we want to deal with the payload accordingly.
func prepRequestData(url, actionType string, data map[string]any, msg *models.Message) (string, io.Reader, error) {
	if len(data) > 0 {
		if actionType == http.MethodGet {
			query, err := createGetQuery(data, msg)
			if err != nil {
				return url, nil, err
			}

			url = fmt.Sprintf("%s?%s", url, query)

			return url, nil, nil
		}

		query, err := createJSONPayload(data, msg)
		if err != nil {
			return url, nil, err
		}

		return url, strings.NewReader(query), nil
	}

	return url, nil, nil
}

// Unmarshal arbitrary JSON.
func extractFields(raw []byte) (any, error) {
	var resp map[string]any

	err := json.Unmarshal(raw, &resp)
	if err != nil {
		var arrResp []map[string]any

		err := json.Unmarshal(raw, &arrResp)
		if err != nil {
			return string(raw), nil
		}

		return arrResp, nil
	}

	return resp, nil
}

// Create GET query string.
func createGetQuery(data map[string]any, msg *models.Message) (string, error) {
	u := url.Values{}

	for k, v := range data {
		subv, err := utils.Substitute(v.(string), msg.Vars)
		if err != nil {
			return "", err
		}

		u.Add(k, subv)
	}

	encoded := u.Encode()                             // uses QueryEscape
	encoded = strings.ReplaceAll(encoded, "+", "%20") // replacing + with more reliable %20

	return encoded, nil
}

// Create querydata payload for non-GET requests.
func createJSONPayload(data map[string]any, msg *models.Message) (string, error) {
	dataNice := utils.MakeNiceJSON(data)

	str, err := json.Marshal(dataNice)
	if err != nil {
		return "", err
	}

	payload, err := utils.Substitute(string(str), msg.Vars)
	if err != nil {
		return "", err
	}

	return payload, nil
}
