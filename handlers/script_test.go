package handlers

import (
	"reflect"
	"testing"

	"github.com/target/flottbot/models"
)

func newExecAction(cmd string) models.Action {
	return models.Action{
		Name: "Simple",
		Type: "exec",
		Cmd:  cmd,
	}
}

func TestScriptExec(t *testing.T) {
	type args struct {
		args models.Action
		msg  *models.Message
		bot  *models.Bot
	}

	bot := new(models.Bot)

	simpleScriptMessage := models.NewMessage()
	simpleScriptMessage.Vars["test"] = "echo"

	simpleScriptAction := newExecAction(`echo "hi there"`)

	slowScriptAction := newExecAction(`sleep 22`)

	errorScriptAction := newExecAction(`false`)

	varExistsScriptAction := newExecAction(`echo "${test}"`)

	varMissingScriptAction := newExecAction(`echo "${notest}"`)

	cmdNotFound := newExecAction(`/bin/sh ./this/is/a/trap.sh`)

	// TODO: fix this test for travis
	// msgBeforeExit := newExecAction(`printf "error is coming"; exit 1;`)

	tests := []struct {
		name    string
		args    args
		want    *models.ScriptResponse
		wantErr bool
	}{
		{"Simple Script", args{args: simpleScriptAction, msg: &simpleScriptMessage, bot: bot}, &models.ScriptResponse{Status: 0, Output: "hi there"}, false},
		{"Slow Script", args{args: slowScriptAction, msg: &simpleScriptMessage, bot: bot}, &models.ScriptResponse{Status: 1, Output: "Hmm, something timed out. Please try again."}, true},
		{"Error Script", args{args: errorScriptAction, msg: &simpleScriptMessage, bot: bot}, &models.ScriptResponse{Status: 1, Output: ""}, true},
		{"Existing Var Script", args{args: varExistsScriptAction, msg: &simpleScriptMessage, bot: bot}, &models.ScriptResponse{Status: 0, Output: "echo"}, false},
		{"Missing Var Script", args{args: varMissingScriptAction, msg: &simpleScriptMessage, bot: bot}, &models.ScriptResponse{Status: 1, Output: ""}, true},
		{"Script does not exist", args{args: cmdNotFound, msg: &simpleScriptMessage, bot: bot}, &models.ScriptResponse{Status: 127, Output: "/bin/sh: 0: Can't open ./this/is/a/trap.sh"}, true},
		// {"StdOut before exit code 1", args{args: msgBeforeExit, msg: &simpleScriptMessage, bot: bot}, &models.ScriptResponse{Status: 1, Output: "error is coming"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ScriptExec(tt.args.args, tt.args.msg, tt.args.bot)
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
