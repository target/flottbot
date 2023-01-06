// Copyright (c) 2023 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package utils

import (
	"reflect"
	"testing"
)

func TestMakeNiceJSON(t *testing.T) {
	type args struct {
		in map[string]any
	}

	testData := make(map[string]any)
	testData["channel"] = "C9816S0B1"
	testData["text"] = "I am a test message http://slack.com"
	testData["attachments"] = map[any]any{"text": "And here's an attachment!"}

	testDataResult := make(map[string]any)
	testDataResult["channel"] = "C9816S0B1"
	testDataResult["text"] = "I am a test message http://slack.com"
	testDataResult["attachments"] = map[string]any{"text": "And here's an attachment!"}

	testDataArray := make(map[string]any)
	testDataArray["foo"] = []any{map[any]any{"text": "And here's an attachment!"}}

	testDataArrayResult := make(map[string]any)
	testDataArrayResult["foo"] = []any{map[string]any{"text": "And here's an attachment!"}}

	tests := []struct {
		name string
		args args
		want map[string]any
	}{
		{"Nested Object", args{in: testData}, testDataResult},
		{"Nested Array", args{in: testDataArray}, testDataArrayResult},
		{"No Change", args{in: testDataResult}, testDataResult},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeNiceJSON(tt.args.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeNiceJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
