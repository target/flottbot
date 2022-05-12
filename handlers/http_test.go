// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/target/flottbot/models"
)

func Test_extractFields(t *testing.T) {
	type args struct {
		raw []byte
	}

	JSONTest := make(map[string]interface{})
	JSONTest["testing"] = "test"

	JSONArrTest := make([]map[string]interface{}, 0)
	JSONArrTest = append(JSONArrTest, map[string]interface{}{"testing": "test"})

	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{"JSON Test", args{raw: []byte(`{ "testing": "test" }`)}, JSONTest, false},
		{"JSON Arr Test", args{raw: []byte(`[{ "testing": "test" }]`)}, JSONArrTest, false},
		{"String Test", args{raw: []byte(`testing`)}, "testing", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractFields(tt.args.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPReq(t *testing.T) {
	type args struct {
		args models.Action
		msg  *models.Message
	}

	tsOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer tsOK.Close()

	tsError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer tsError.Close()

	customHeader := make(map[string]string)
	customHeader["testHeader"] = "testHeaderContent"

	customQueryData := make(map[string]interface{})
	customQueryData["testQuery"] = "${testValues}"

	customQueryDataWithVars := make(map[string]interface{})
	customQueryDataWithVars["testQuery"] = "${test}"

	TestMessage := models.NewMessage()
	TestMessage.Vars["testValues"] = "test"

	TestGETAction := models.Action{
		Name:          "Test Action",
		Type:          "GET",
		URL:           tsOK.URL,
		CustomHeaders: customHeader,
		QueryData:     customQueryData,
	}
	TestPOSTAction := models.Action{
		Name:          "Test Action",
		Type:          "POST",
		URL:           tsOK.URL,
		CustomHeaders: customHeader,
		QueryData:     customQueryData,
	}
	TestEmptyQueryAction := models.Action{
		Name:          "Test Action",
		Type:          "GET",
		URL:           tsOK.URL,
		CustomHeaders: customHeader,
	}
	TestErrorResponseAction := models.Action{
		Name: "Test Action",
		Type: "GET",
		URL:  tsError.URL,
	}
	TestQueryWithSubsAction := models.Action{
		Name:      "Test Action",
		Type:      "GET",
		URL:       tsOK.URL,
		QueryData: customQueryDataWithVars,
	}
	TestWithError := models.Action{
		Name: "Error Case",
		Type: "GET",
		URL:  "/%zz",
	}

	tests := []struct {
		name    string
		args    args
		want    *models.HTTPResponse
		wantErr bool
	}{
		{"HTTPReq GET", args{args: TestGETAction, msg: &TestMessage}, &models.HTTPResponse{Status: 200, Raw: "", Data: ""}, false},
		{"HTTPReq POST", args{args: TestPOSTAction, msg: &TestMessage}, &models.HTTPResponse{Status: 200, Raw: "", Data: ""}, false},
		{"HTTPReq No Query", args{args: TestEmptyQueryAction, msg: &TestMessage}, &models.HTTPResponse{Status: 200, Raw: "", Data: ""}, false},
		{"HTTPReq Error Response", args{args: TestErrorResponseAction, msg: &TestMessage}, &models.HTTPResponse{Status: 502, Raw: "", Data: ""}, false},
		{"HTTPReq with Sub", args{args: TestQueryWithSubsAction, msg: &TestMessage}, nil, true},
		{"HTTPReq with Error", args{args: TestWithError, msg: &TestMessage}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HTTPReq(tt.args.args, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPReq() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPReq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_prepRequestData(t *testing.T) {
	type args struct {
		url        string
		actionType string
		data       map[string]interface{}
		msg        *models.Message
	}

	tests := []struct {
		name    string
		args    args
		want    string
		want1   io.Reader
		wantErr bool
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := prepRequestData(tt.args.url, tt.args.actionType, tt.args.data, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("prepRequestData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("prepRequestData() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("prepRequestData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_createGetQuery(t *testing.T) {
	type args struct {
		data map[string]interface{}
		msg  *models.Message
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Simple Query", args{data: map[string]interface{}{"foo": "bar"}, msg: new(models.Message)}, "foo=bar", false},
		{"Query with Spaces", args{data: map[string]interface{}{"foo": "bar foo"}, msg: new(models.Message)}, `foo=bar%20foo`, false},
		{"Query with Plus", args{data: map[string]interface{}{"foo": "bar+foo"}, msg: new(models.Message)}, `foo=bar%2Bfoo`, false},
		{"Empty", args{data: make(map[string]interface{}), msg: new(models.Message)}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createGetQuery(tt.args.data, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("createGetQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("createGetQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createJSONPayload(t *testing.T) {
	type args struct {
		data map[string]interface{}
		msg  *models.Message
	}

	testMsg := new(models.Message)
	testData := make(map[string]interface{})

	// map[attachments:map[text:And here's an attachment!] channel:C9816S0B1 text:I am a test message http://slack.com]
	testData["attachments"] = map[interface{}]interface{}{"text": "And here's an attachment!"}
	testData["channel"] = "C9816S0B1"
	testData["text"] = "I am a test message http://slack.com"

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"YAML parser fail", args{data: testData, msg: testMsg}, `{"attachments":{"text":"And here's an attachment!"},"channel":"C9816S0B1","text":"I am a test message http://slack.com"}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createJSONPayload(tt.args.data, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("createJSONPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("createJSONPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}
