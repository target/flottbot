// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package handlers

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/target/flottbot/models"
)

func newExecAction(cmd string) models.Action {
	return models.Action{
		Name:    "Simple",
		Type:    "exec",
		Cmd:     cmd,
		Timeout: 1,
	}
}

func TestScriptExec(t *testing.T) {
	type args struct {
		args models.Action
		msg  *models.Message
	}

	simpleScriptMessage := models.NewMessage()
	simpleScriptMessage.Vars["test"] = "echo"

	simpleScriptAction := newExecAction(`echo "hi there"`)

	slowScriptAction := newExecAction(`sleep 2`)

	errorScriptAction := newExecAction(`false`)

	varExistsScriptAction := newExecAction(`echo "${test}"`)

	varMissingScriptAction := newExecAction(`echo "${notest}"`)

	msgBeforeExit := newExecAction(`/bin/sh ../testdata/fail.sh`)

	tests := []struct {
		name    string
		args    args
		want    *models.ScriptResponse
		wantErr bool
	}{
		{"Simple Script", args{args: simpleScriptAction, msg: &simpleScriptMessage}, &models.ScriptResponse{Status: 0, Output: "hi there"}, false},
		{"Slow Script", args{args: slowScriptAction, msg: &simpleScriptMessage}, &models.ScriptResponse{Status: 1, Output: "Hmm, the command timed out. Please try again."}, true},
		{"Error Script", args{args: errorScriptAction, msg: &simpleScriptMessage}, &models.ScriptResponse{Status: 1, Output: ""}, true},
		{"Existing Var Script", args{args: varExistsScriptAction, msg: &simpleScriptMessage}, &models.ScriptResponse{Status: 0, Output: "echo"}, false},
		{"Missing Var Script", args{args: varMissingScriptAction, msg: &simpleScriptMessage}, &models.ScriptResponse{Status: 1, Output: ""}, true},
		{"StdOut before exit code 1", args{args: msgBeforeExit, msg: &simpleScriptMessage}, &models.ScriptResponse{Status: 1, Output: "error is coming"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ScriptExec(tt.args.args, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScriptExec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScriptExec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScriptExecWithRegex(t *testing.T) {
	type args struct {
		args models.Action
		msg  *models.Message
	}

	simpleScriptMessage := models.NewMessage()

	cmdNotFound := newExecAction(`/bin/sh ./this/is/a/trap.sh`)

	tests := []struct {
		name       string
		args       args
		want       *models.ScriptResponse
		wantErr    bool
		wantRegexp *regexp.Regexp
	}{
		{"Script does not exist", args{args: cmdNotFound, msg: &simpleScriptMessage}, &models.ScriptResponse{Status: 2}, true, regexp.MustCompile(`No such file`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ScriptExec(tt.args.args, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScriptExec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantRegexp.MatchString(got.Output) {
				t.Errorf("ScriptExec() = %v, want %v", got.Output, "Regexp(`"+tt.wantRegexp.String()+"`)")
			}
		})
	}
}
