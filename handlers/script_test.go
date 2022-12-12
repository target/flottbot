// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package handlers

import (
  "reflect"
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
		action models.Action
		msg    *models.Message
	}

	simpleScriptMessage := models.NewMessage()
	simpleScriptMessage.Vars["test"] = "echo"

	simpleScriptAction := newExecAction(`echo "hi there"`)

	slowScriptAction := newExecAction(`sleep 2`)

	errorScriptAction := newExecAction(`false`)

	varExistsScriptAction := newExecAction(`echo "${test}"`)

	varMissingScriptAction := newExecAction(`echo "${notest}"`)

	msgBeforeExit := newExecAction(`/bin/sh ../testdata/fail.sh`)

	notExists := newExecAction(`./trap.sh`)

	notExistsInPath := newExecAction(`trap.sh`)

	// assumes existence of /bin/sh
	fileNotFound := newExecAction(`/bin/sh ./this/is/a/trap.sh`)

	tests := []struct {
		name    string
		args    args
		want    *models.ScriptResponse
		wantErr bool
	}{
		{
			"Simple Script",
			args{
				action: simpleScriptAction,
				msg:    &simpleScriptMessage,
			},
			&models.ScriptResponse{
				Status: 0,
				Output: "hi there",
			},
			false,
		},
		{
			"Slow Script",
			args{
				action: slowScriptAction,
				msg:    &simpleScriptMessage,
			},
			&models.ScriptResponse{
				Status: 1,
				Output: "Hmm, the command timed out. Please try again.",
			},
			true,
		},
		{
			"Error Script",
			args{
				action: errorScriptAction,
				msg:    &simpleScriptMessage,
			},
			&models.ScriptResponse{
				Status: 1,
				Output: "",
			},
			true,
		},
		{
			"Existing Var Script",
			args{
				action: varExistsScriptAction,
				msg:    &simpleScriptMessage,
			},
			&models.ScriptResponse{
				Status: 0,
				Output: "echo",
			},
			false,
		},
		{
			"Missing Var Script",
			args{
				action: varMissingScriptAction,
				msg:    &simpleScriptMessage,
			},
			&models.ScriptResponse{
				Status: 1,
				Output: "",
			},
			true,
		},
		{
			"StdOut before exit code 1",
			args{
				action: msgBeforeExit,
				msg:    &simpleScriptMessage,
			},
			&models.ScriptResponse{
				Status: 1,
				Output: "error is coming",
			},
			true,
		},
		{
			"Script doesn't exist",
			args{
				action: notExists,
				msg:    &simpleScriptMessage,
			},
			&models.ScriptResponse{
				Status: 1,
				Output: "file not found: ./trap.sh",
			},
			true,
		},
		{
			"Script doesn't exist in PATH",
			args{
				action: notExistsInPath,
				msg:    &simpleScriptMessage,
			},
			&models.ScriptResponse{
				Status: 1,
				Output: "file not found: trap.sh",
			},
			true,
		},
		{
			"Calling file doesn't exist",
			args{
				action: fileNotFound,
				msg:    &simpleScriptMessage,
			},
			&models.ScriptResponse{
				Status: 127,
				Output: "file not found: /bin/sh ./this/is/a/trap.sh",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ScriptExec(tt.args.action, tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScriptExec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
      
      // special error handling
      if err != nil {
        // if there's an error, the status should not be 0
        // we're not matching the exit code exactly due to
        // architecture/os not being consistent here
        if got.Status == 0 {
          t.Errorf("Status = %v, want %v", got.Status, tt.want.Status)
        }
  
        // if there's an error, the output should still match
        if got.Output != tt.want.Output {
          t.Errorf("Output = %v, want %v", got.Status, tt.want.Output)
        }

        // exit
        return
      }

      // for non-error, check to make sure
      // response is as expected
      if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScriptExec() = %v, want %v", got, tt.want)
			}

		})
	}
}
