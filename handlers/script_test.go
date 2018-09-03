package handlers

import (
	"reflect"
	"testing"

	"github.com/target/flottbot/models"
)

func TestScriptExec(t *testing.T) {
	type args struct {
		args models.Action
		msg  *models.Message
		bot  *models.Bot
	}

	bot := new(models.Bot)

	simpleScriptMessage := models.NewMessage()
	simpleScriptMessage.Vars["test"] = "echo"

	simpleScriptAction := models.Action{
		Name: "Simple",
		Type: "exec",
		Cmd:  `echo "hi there"`,
	}

	slowScriptAction := models.Action{
		Name: "Simple",
		Type: "exec",
		Cmd:  `sleep 22`,
	}

	errorScriptAction := models.Action{
		Name: "Simple",
		Type: "exec",
		Cmd:  `false`,
	}

	varExistsScriptAction := models.Action{
		Name: "Simple",
		Type: "exec",
		Cmd:  `echo "${test}"`,
	}

	varMissingScriptAction := models.Action{
		Name: "Simple",
		Type: "exec",
		Cmd:  `echo "${notest}"`,
	}

	cmdNotFound := models.Action{
		Name: "Simple",
		Type: "exec",
		Cmd:  `/bin/sh ./this/is/a/trap.sh`,
	}

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
		{"Script does not exist", args{args: cmdNotFound, msg: &simpleScriptMessage, bot: bot}, &models.ScriptResponse{Status: 127, Output: "/bin/sh: ./this/is/a/trap.sh: No such file or directory"}, true},
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
