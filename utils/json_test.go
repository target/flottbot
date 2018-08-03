package utils

import (
	"reflect"
	"testing"
)

func TestMakeNiceJSON(t *testing.T) {
	type args struct {
		in map[string]interface{}
	}

	testData := make(map[string]interface{})
	testData["channel"] = "C9816S0B1"
	testData["text"] = "I am a test message http://slack.com"
	testData["attachments"] = map[interface{}]interface{}{"text": "And here's an attachment!"}

	testDataResult := make(map[string]interface{})
	testDataResult["channel"] = "C9816S0B1"
	testDataResult["text"] = "I am a test message http://slack.com"
	testDataResult["attachments"] = map[string]interface{}{"text": "And here's an attachment!"}

	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{"Nested Object", args{in: testData}, testDataResult},
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
