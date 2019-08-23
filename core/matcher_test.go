package core

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/target/flottbot/models"
)

func TestCraftResponse(t *testing.T) {
	type args struct {
		rule models.Rule
		msg  models.Message
		bot  *models.Bot
	}

	// Init test variables
	testBot := new(models.Bot)

	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantErr    bool
	}{
		{
			"Successful craft response (no templates, no var substitution)",
			args{
				rule: models.Rule{
					FormatOutput:      "test output",
					DirectMessageOnly: true,
					OutputToRooms:     []string{"not_a_real_room_1", "not_a_real_room_2"},
				},
				msg: models.Message{
					OutputToRooms: []string{"not_a_real_room_1", "not_a_real_room_2"},
					Vars:          map[string]string{},
				},
				bot: testBot,
			},
			"test output",
			false,
		},
		{"Empty rule format output", args{rule: models.Rule{FormatOutput: ""}, msg: models.Message{}, bot: testBot}, "test output", true},
		{
			"Successful craft response (no templates, with var substitution)",
			args{
				rule: models.Rule{
					FormatOutput:      "here is ${test_var}",
					DirectMessageOnly: true,
					OutputToRooms:     []string{"not_a_real_room_1", "not_a_real_room_2"},
				},
				msg: models.Message{
					OutputToRooms: []string{"not_a_real_room_1", "not_a_real_room_2"},
					Vars: map[string]string{
						"test_var": "some value",
					},
				},
				bot: testBot,
			},
			"here is some value",
			false,
		},
		{
			"Successful craft response (with templates, with var substitution)",
			args{
				rule: models.Rule{
					FormatOutput:      `{{ if (eq "${_test_status}" "ok") }}hello{{ else }}hi{{ end }}`,
					DirectMessageOnly: true,
					OutputToRooms:     []string{"not_a_real_room_1", "not_a_real_room_2"},
				},
				msg: models.Message{
					OutputToRooms: []string{"not_a_real_room_1", "not_a_real_room_2"},
					Vars: map[string]string{
						"_test_status": "ok",
					},
				},
				bot: testBot,
			},
			"hello",
			false,
		},
		{
			"Successful craft response (with templates, with var substitution)",
			args{
				rule: models.Rule{
					FormatOutput:      `{{ if (eq "${_test_status}" "ok") }}hello{{ else }}hi{{ end }}`,
					DirectMessageOnly: true,
					OutputToRooms:     []string{"not_a_real_room_1", "not_a_real_room_2"},
				},
				msg: models.Message{
					OutputToRooms: []string{"not_a_real_room_1", "not_a_real_room_2"},
					Vars: map[string]string{
						"_test_status": "not_ok",
					},
				},
				bot: testBot,
			},
			"hi",
			false,
		},
		{
			"Successful craft response (none of the rooms exist and OutputToUsers empty)",
			args{
				rule: models.Rule{
					FormatOutput:      `{{ if (eq "${_test_status}" "ok") }}hello{{ else }}hi{{ end }}`,
					DirectMessageOnly: false,
					OutputToRooms:     []string{"not_a_real_room_1", "not_a_real_room_2"},
					OutputToUsers:     []string{},
				},
				msg: models.Message{
					OutputToRooms: []string{},
					Vars: map[string]string{
						"_test_status": "not_ok",
					},
				},
				bot: testBot,
			},
			"hi",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := craftResponse(tt.args.rule, tt.args.msg, tt.args.bot)
			if (err != nil) != tt.wantErr {
				t.Errorf("craftResponse() error = \"%v\", wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr { // all happy paths (i.e. no errors) go here
				if got != tt.wantOutput {
					t.Errorf("craftResponse() got = \"%s\", wantOutput %s", got, tt.wantOutput)
				}
			}
		})
	}
}

func TestHandleExec(t *testing.T) {
	type args struct {
		action models.Action
		msg    *models.Message
		bot    *models.Bot
	}

	// Init test variables
	bot := new(models.Bot)

	testScriptMessage := models.NewMessage()
	testScriptMessage.Vars["test"] = "echo"

	testScriptAction := models.Action{
		Name: "Test",
		Type: "exec",
		Cmd:  `echo "hi there"`,
	}

	testPassScriptResponse := models.ScriptResponse{
		Status: 0,
		Output: "hi there",
	}

	testSlowScriptAction := models.Action{
		Name:    "Test",
		Type:    "exec",
		Cmd:     `sleep 5`,
		Timeout: 2,
	}

	testFailScriptResponse := models.ScriptResponse{
		Status: 1,
		Output: "hmm, something timed out. Please try again",
	}

	testNoCmdScriptAction := models.Action{
		Name: "Test",
		Type: "exec",
		Cmd:  ``,
	}

	tests := []struct {
		name               string
		args               args
		wantScriptResponse *models.ScriptResponse
		wantErr            bool
	}{
		{"Test echo script", args{action: testScriptAction, msg: &testScriptMessage, bot: bot}, &testPassScriptResponse, false},
		{"Slow Script", args{action: testSlowScriptAction, msg: &testScriptMessage, bot: bot}, &testFailScriptResponse, true},
		{"No Cmd Script", args{action: testNoCmdScriptAction, msg: &testScriptMessage, bot: bot}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleExec(tt.args.action, tt.args.msg, tt.args.bot)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleExec() error = \"%v\", wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr { // all happy paths (i.e. no errors) go here
				if tt.args.msg.Vars["_exec_output"] != tt.wantScriptResponse.Output {
					t.Errorf("handleExec() = \"%s\", want \"%v\"", tt.args.msg.Vars["_exec_output"], tt.wantScriptResponse.Output)
				}
				if tt.args.msg.Vars["_exec_status"] != strconv.Itoa(tt.wantScriptResponse.Status) {
					t.Errorf("handleExec() = %s, want %v", tt.args.msg.Vars["_exec_status"], tt.wantScriptResponse.Status)
				}
			}
		})
	}
}

func TestHandleHTTP(t *testing.T) {
	type args struct {
		action models.Action
		msg    *models.Message
		bot    *models.Bot
	}

	// Init test variables
	bot := new(models.Bot)

	tsOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	}))
	defer tsOK.Close()

	tsError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer tsError.Close()

	tsOKJSON := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test": "value"}`))
	}))
	defer tsOKJSON.Close()

	customHeader := make(map[string]string)
	customHeader["testHeader"] = "testHeaderContent"

	customQueryData := make(map[string]interface{})
	customQueryData["testQuery"] = "${testValues}"

	customJSONFields := make(map[string]string)
	customJSONFields["var"] = ".test"

	testMsg := models.NewMessage()
	testMsg.Vars["testValues"] = "test"

	TestEmptyURLAction := models.Action{
		Name: "Test Action",
		Type: "GET",
		URL:  "",
	}

	TestErrorResponseAction := models.Action{
		Name: "Test Action",
		Type: "GET",
		URL:  tsError.URL,
	}

	TestGETAction := models.Action{
		Name:          "Test Action",
		Type:          "GET",
		URL:           tsOK.URL,
		CustomHeaders: customHeader,
		QueryData:     customQueryData,
	}

	TestGETActionWithJSON := models.Action{
		Name:             "Test Action",
		Type:             "GET",
		URL:              tsOKJSON.URL,
		CustomHeaders:    customHeader,
		QueryData:        customQueryData,
		ExposeJSONFields: customJSONFields,
	}

	tests := []struct {
		name         string
		args         args
		wantResponse *models.HTTPResponse
		wantErr      bool
	}{
		{"No URL", args{action: TestEmptyURLAction, msg: &testMsg, bot: bot}, &models.HTTPResponse{}, true},
		{"HTTP GET 200", args{action: TestGETAction, msg: &testMsg, bot: bot}, &models.HTTPResponse{Status: 200, Raw: "hello", Data: ""}, false},
		{"HTTP GET 404", args{action: TestErrorResponseAction, msg: &testMsg, bot: bot}, &models.HTTPResponse{Status: 404, Raw: "not found", Data: ""}, false},
		{
			"HTTP GET 200 JSON",
			args{action: TestGETActionWithJSON, msg: &testMsg, bot: bot},
			&models.HTTPResponse{
				Status: 200,
				Raw:    `{"test": "value"}`,
				Data:   ""},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleHTTP(tt.args.action, tt.args.msg, tt.args.bot)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleHTTP() error = \"%v\", wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr { // all happy paths (i.e. no errors) go here
				if tt.args.msg.Vars["_raw_http_output"] != tt.wantResponse.Raw {
					t.Errorf("handleHTTP() = \"%s\", want \"%v\"", tt.args.msg.Vars["_raw_http_output"], tt.wantResponse.Raw)
				}
				if tt.args.msg.Vars["_raw_http_status"] != strconv.Itoa(tt.wantResponse.Status) {
					t.Errorf("handleHTTP() = %s, want %v", tt.args.msg.Vars["_raw_http_status"], tt.wantResponse.Status)
				}
			}
		})
	}
}

func TestHandleMessage(t *testing.T) {
	type args struct {
		action         models.Action
		outputMsgs     chan<- models.Message
		msg            *models.Message
		direct         bool
		startMsgThread bool
		hitRule        chan<- models.Rule
		bot            *models.Bot
	}

	// Init test variables
	testAction := new(models.Action)
	testAction.LimitToRooms = []string{}
	testMsg := new(models.Message)
	testMsg.Attributes = make(map[string]string)
	testMsg.OutputToRooms = []string{}
	testActiveRooms := make(map[string]string)
	testActiveRooms["flottbot-room1"] = "12345"
	testActiveRooms["flottbot-room2"] = "54321"
	bot := new(models.Bot)
	bot.Rooms = testActiveRooms

	tests := []struct {
		name              string
		args              args
		wantLimitToRooms  []string
		wantOutputToRooms []string
		wantActionMessage string
		wantOutputMessage string
		wantErr           bool
	}{
		{
			"Send non-direct message",
			args{*testAction, nil, testMsg, false, false, nil, bot},
			[]string{},
			[]string{"flottbot-room1", "flottbot-room2"},
			"Message from action",
			"Message from action",
			false,
		},
		{
			"Send direct message but limit_to_rooms is set",
			args{*testAction, nil, testMsg, true, false, nil, bot},
			[]string{"flottbot-room1", "flottbot-room2"},
			[]string{},
			"Message from action",
			"Message from action",
			false,
		},
		{
			"Send non-direct message but limit_to_rooms is set",
			args{*testAction, nil, testMsg, true, false, nil, bot},
			[]string{"flottbot-room1", "flottbot-room2"},
			[]string{},
			"Message from action",
			"Message from action",
			false,
		},
		{
			"Send non-direct message but limit_to_rooms is not set",
			args{*testAction, nil, testMsg, true, false, nil, bot},
			[]string{},
			[]string{},
			"Message from action",
			"Message from action",
			false,
		},
		{
			"Send non-direct start-thread message",
			args{*testAction, nil, testMsg, false, true, nil, bot},
			[]string{},
			[]string{},
			"Message from action",
			"Message from action",
			false,
		},
		{
			"Empty action message",
			args{*testAction, nil, testMsg, false, false, nil, bot},
			[]string{},
			[]string{"flottbot-room1", "flottbot-room2"},
			"",
			"",
			true,
		},
		{
			"Error on Substitute()",
			args{*testAction, nil, testMsg, false, false, nil, bot},
			[]string{},
			[]string{"flottbot-room1", "flottbot-room2"},
			"${NOT_A_REAL_VAR}",
			"",
			true,
		},
		{
			"Rooms in limit_to_rooms don't exist",
			args{*testAction, nil, testMsg, false, false, nil, bot},
			[]string{"test", "test2"},
			[]string{},
			"Message",
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test variables
			var testOutputMsgs chan models.Message
			var testHitRule chan models.Rule
			tt.args.action.LimitToRooms = tt.wantLimitToRooms
			tt.args.action.Message = tt.wantActionMessage
			tt.args.msg.OutputToRooms = tt.wantOutputToRooms
			tt.args.msg.Output = tt.wantOutputMessage
			if !tt.wantErr { // all happy paths (i.e. no errors) go here
				testOutputMsgs = make(chan models.Message, 1)
				testHitRule = make(chan models.Rule, 1)
				tt.args.outputMsgs = testOutputMsgs
				tt.args.hitRule = testHitRule
			}
			// Do test
			err := handleMessage(tt.args.action, tt.args.outputMsgs, tt.args.msg, tt.args.direct, tt.args.startMsgThread, tt.args.hitRule, tt.args.bot)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleMessage() error = %v, wantErr %v", err, tt.wantErr)
			} else if (err == nil) == tt.wantErr {
				t.Errorf("handleMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr { // all happy paths (i.e. no errors) go here // this is actually pretty dangerous - this could cause a blocking wait and a panic... maybe look to test this better
				resultMsg := <-testOutputMsgs
				if tt.wantOutputMessage != resultMsg.Output {
					t.Errorf("handleMessage() wanted message \"%s\", but got \"%s\"", tt.wantOutputMessage, resultMsg.Output)
				}
			}
		})
	}
}

func TestHandleReaction(t *testing.T) {
	type args struct {
		outputMsgs chan<- models.Message
		msg        *models.Message
		hitRule    chan<- models.Rule
		rule       models.Rule
	}

	// Init test variables
	testOutputMsgs := make(chan models.Message, 1)
	testHitRule := make(chan models.Rule, 1)
	testMsg := new(models.Message)
	testRule := new(models.Rule)

	test := struct {
		name         string
		args         args
		wantMessage  string
		wantRuleName string
	}{
		"Test handleReaction()",
		args{testOutputMsgs, testMsg, testHitRule, *testRule},
		"test output",
		"test rule",
	}
	t.Run(test.name, func(t *testing.T) {
		// Set test variables
		test.args.msg.Output = test.wantMessage
		test.args.rule.Name = test.wantRuleName
		// Do test
		handleReaction(test.args.outputMsgs, test.args.msg, test.args.hitRule, test.args.rule)
		resultMsg := <-testOutputMsgs
		resultRule := <-testHitRule
		if test.wantMessage != resultMsg.Output {
			t.Errorf("handReaction() wanted message \"%s\", but got \"%s\"", test.wantMessage, resultMsg.Output)
		}
		if test.wantRuleName != resultRule.Name {
			t.Errorf("handReaction() wanted rule '%s', but got '%s'", test.wantRuleName, resultRule.Name)
		}
	})
}

func TestUpdateReaction(t *testing.T) {
	type args struct {
		action models.Action
		rule   *models.Rule
		vars   map[string]string
		bot    *models.Bot
	}

	// Init test args
	testAction := new(models.Action)
	testRule := new(models.Rule)
	testVars := make(map[string]string)
	bot := new(models.Bot)

	// Set test variables
	testHTTPStatusTemplate := `
	{{ if (eq "${_raw_http_status}" "202") }}
	  check_mark
	{{ else }}
	  x
	{{ end }}`
	testExecStatusTemplate := `
	{{ if (eq "${_exec_status}" "0") }}
	  check_mark
	{{ else }}
	  x
	{{ end }}`

	// Construct and execute test cases
	tests := []struct {
		name           string
		args           args
		reaction       string
		updateReaction string
		want           string
	}{
		{"No reaction to update", args{*testAction, testRule, testVars, bot}, "wait", "", "wait"},
		{"Update wait to done", args{*testAction, testRule, testVars, bot}, "wait", "done", "done"},
		{"Update wait to check_mark with golang templating (http status)", args{*testAction, testRule, testVars, bot}, "wait", testHTTPStatusTemplate, "check_mark"},
		{"Update wait to x with golang templating (http status)", args{*testAction, testRule, testVars, bot}, "wait", testHTTPStatusTemplate, "x"},
		{"Update wait to check_mark with golang templating (exec status)", args{*testAction, testRule, testVars, bot}, "wait", testExecStatusTemplate, "check_mark"},
		{"Update wait to x with golang templating (exec status)", args{*testAction, testRule, testVars, bot}, "wait", testExecStatusTemplate, "x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.want {
			case "check_mark":
				testVars["_raw_http_status"] = "202"
				testVars["_exec_status"] = "0"
			case "x":
				testVars["_raw_http_status"] = "404"
				testVars["_exec_status"] = "1"
			default:
				break
			}
			tt.args.rule.Reaction = tt.reaction
			tt.args.action.Reaction = tt.updateReaction
			updateReaction(tt.args.action, tt.args.rule, tt.args.vars, tt.args.bot)
			if tt.args.rule.Reaction != tt.want {
				t.Errorf("updateReaction() wanted %s, but got %s", tt.want, tt.args.rule.Reaction)
			}
		})
	}
}

func Test_getProccessedInputAndHitValue(t *testing.T) {
	type args struct {
		messageInput     string
		ruleRespondValue string
		ruleHearValue    string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{"hit", args{"hello foo", "hello", "hello"}, "foo", true},
		{"hit no hear value", args{"hello foo", "hello", ""}, "foo", true},
		{"hit no respond value - drops args", args{"hello foo", "", "hello"}, "", true},
		{"no match", args{"hello foo", "", ""}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getProccessedInputAndHitValue(tt.args.messageInput, tt.args.ruleRespondValue, tt.args.ruleHearValue)
			if got != tt.want {
				t.Errorf("getProccessedInputAndHitValue() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getProccessedInputAndHitValue() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_isValidHitChatRule(t *testing.T) {
	type args struct {
		message        *models.Message
		rule           models.Rule
		processedInput string
		bot            *models.Bot
	}

	testBot := new(models.Bot)
	testRule := models.Rule{}
	testMessage := new(models.Message)
	happyVars := make(map[string]string)
	happyVars["_user.name"] = "fooUser"
	testMessage.Vars = happyVars

	testRuleFail := models.Rule{}
	testRuleFail.AllowUsers = []string{"barUser"}
	testMessageFail := new(models.Message)
	failVars := make(map[string]string)
	failVars["_user.name"] = "fooUser"
	testMessageFail.Vars = failVars

	testRuleUserAllowed := models.Rule{}
	testRuleUserAllowed.AllowUsers = []string{"fooUser"}
	testMessageUserAllowed := new(models.Message)
	userAllowedVars := make(map[string]string)
	userAllowedVars["_user.name"] = "fooUser"
	testMessageUserAllowed.Vars = userAllowedVars

	testRuleNeedArg := models.Rule{}
	testRuleNeedArg.AllowUsers = []string{"fooUser"}
	testRuleNeedArg.Args = []string{"arg1", "arg2"}
	testMessageNeedArg := new(models.Message)
	needArgVars := make(map[string]string)
	needArgVars["_user.name"] = "fooUser"
	testMessageNeedArg.Vars = needArgVars

	testRuleArgs := models.Rule{}
	testRuleArgs.AllowUsers = []string{"fooUser"}
	testRuleArgs.Args = []string{"arg1", "arg2"}
	testMessageArgs := new(models.Message)
	argsVars := make(map[string]string)
	argsVars["_user.name"] = "fooUser"
	testMessageArgs.Vars = argsVars

	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Happy", args{testMessage, testRule, "foo", testBot}, true},
		{"User not allowed", args{testMessageFail, testRuleFail, "foo", testBot}, false},
		{"User allowed, no user group restriction", args{testMessageUserAllowed, testRuleUserAllowed, "foo", testBot}, true},
		{"User allowed, not enough args for respond", args{testMessageNeedArg, testRuleNeedArg, "arg1", testBot}, false},
		{"User allowed, process args", args{testMessageArgs, testRuleArgs, "arg1 arg2", testBot}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidHitChatRule(tt.args.message, tt.args.rule, tt.args.processedInput, tt.args.bot); got != tt.want {
				t.Errorf("isValidHitChatRule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleChatServiceRule(t *testing.T) {
	type args struct {
		outputMsgs     chan<- models.Message
		message        models.Message
		hitRule        chan<- models.Rule
		rule           models.Rule
		processedInput string
		hit            bool
		bot            *models.Bot
	}

	rule := models.Rule{
		Name:     "Test Rule",
		Respond:  "foo",
		Args:     []string{"arg1", "arg2"},
		HelpText: "foo <arg1> <arg2>",
	}

	ruleOpt := models.Rule{
		Name:     "Test Rule with optional arg",
		Respond:  "foo",
		Args:     []string{"arg1", "arg2?"},
		HelpText: "foo <arg1> <arg2>",
	}

	ruleHearWithArgs := models.Rule{
		Name: "Hear rule with Args set",
		Hear: "/hi/",
		Args: []string{"arg1", "arg2"},
	}

	ruleIgnoreThread := models.Rule{
		Name:          "Test rule that ignores thread",
		Hear:          "/thread/",
		IgnoreThreads: true,
	}

	testBot := new(models.Bot)
	testBot.Name = "Testbot"

	testMessage := models.Message{
		Input:        "foo arg1 arg2",
		Vars:         map[string]string{},
		Attributes:   map[string]string{},
		BotMentioned: true,
	}

	testMessageBotNotMentioned := models.Message{
		Input:      "foo arg1 arg2",
		Vars:       map[string]string{},
		Attributes: map[string]string{},
	}

	testMessageNotEnoughArgs := models.Message{
		Input:        "foo arg1",
		Vars:         map[string]string{},
		BotMentioned: true,
	}

	testMessageOptionalArgs := models.Message{
		Input:        "foo arg1",
		Vars:         map[string]string{},
		BotMentioned: true,
	}

	testMessageIgnoreThread := models.Message{
		Input:           "we have a thread",
		Vars:            map[string]string{},
		Timestamp:       "x",
		ThreadTimestamp: "x",
	}

	tests := []struct {
		name      string
		args      args
		want      bool
		want1     bool
		expectMsg string
	}{
		{"basic", args{}, false, false, ""},
		{"respond + hear", args{rule: models.Rule{Respond: "hi", Hear: "/hi/"}, hit: false, bot: testBot, message: testMessage}, false, false, ""},
		{"hear + rule args", args{rule: ruleHearWithArgs, hit: false, bot: testBot, message: testMessage}, false, false, ""},
		{"respond rule - hit false", args{rule: rule, hit: false}, false, false, ""},
		{"respond rule - hit true - valid", args{rule: rule, hit: true, bot: testBot, message: testMessage, processedInput: "arg1 arg2"}, true, true, "hmm, the 'format_output' field in your configuration is empty"},
		{"respond rule - hit true - bot not mentioned", args{rule: rule, hit: true, bot: testBot, message: testMessageBotNotMentioned, processedInput: "arg1 arg2"}, false, false, ""},
		{"respond rule - hit true - valid - not enough args", args{rule: rule, hit: true, bot: testBot, message: testMessageNotEnoughArgs, processedInput: "arg1"}, true, true, "You might be missing an argument or two. This is what I'm looking for\n```foo <arg1> <arg2>```"},
		{"respond rule - hit true - valid optional arg", args{rule: ruleOpt, hit: true, bot: testBot, message: testMessageOptionalArgs, processedInput: "arg1"}, true, true, ""},
		{"respond rule - hit true - invalid", args{rule: rule, hit: true, bot: testBot, message: testMessage}, true, true, "You might be missing an argument or two. This is what I'm looking for\n```foo <arg1> <arg2>```"},
		{"hear rule - ignore thread", args{rule: ruleIgnoreThread, hit: true, bot: testBot, message: testMessageIgnoreThread}, true, true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testOutput := make(chan models.Message, 1)
			testHitRule := make(chan models.Rule, 1)
			tt.args.hitRule = testHitRule
			tt.args.outputMsgs = testOutput

			got, got1 := handleChatServiceRule(tt.args.outputMsgs, tt.args.message, tt.args.hitRule, tt.args.rule, tt.args.processedInput, tt.args.hit, tt.args.bot)

			select {
			case output := <-testOutput:
				if tt.expectMsg != output.Output {
					t.Errorf("Output message didn't match, got = %v, want %v", output.Output, tt.expectMsg)
				}
				if got != tt.want {
					t.Errorf("handleChatServiceRule() got = %v, want %v", got, tt.want)
				}
				if got1 != tt.want1 {
					t.Errorf("handleChatServiceRule() got1 = %v, want %v", got1, tt.want1)
				}
			default:
				if got != tt.want {
					t.Errorf("handleChatServiceRule() got = %v, want %v", got, tt.want)
				}
				if got1 != tt.want1 {
					t.Errorf("handleChatServiceRule() got1 = %v, want %v", got1, tt.want1)
				}
			}
		})
	}
}

func Test_handleSchedulerServiceRule(t *testing.T) {
	type args struct {
		outputMsgs chan<- models.Message
		message    models.Message
		hitRule    chan<- models.Rule
		rule       models.Rule
		bot        *models.Bot
	}

	testBot := new(models.Bot)

	testRuleValid := models.Rule{
		Schedule:      "@every 5s",
		Name:          "TestSchedule",
		Respond:       "foo",
		FormatOutput:  "Hello, from Scheduler 1!",
		Args:          []string{"arg1"},
		Active:        true,
		OutputToRooms: []string{"test-room1"},
	}

	testMessageValid := models.Message{
		Attributes: map[string]string{"from_schedule": "TestSchedule"},
		Input:      "foo arg1",
	}

	tests := []struct {
		name  string
		args  args
		want  bool
		want1 bool
	}{
		{"Basic", args{}, false, false},
		{"Valid Schedule", args{message: testMessageValid, rule: testRuleValid, bot: testBot}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := handleSchedulerServiceRule(tt.args.outputMsgs, tt.args.message, tt.args.hitRule, tt.args.rule, tt.args.bot)
			if got != tt.want {
				t.Errorf("handleSchedulerServiceRule() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("handleSchedulerServiceRule() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_handleNoMatch(t *testing.T) {
	type args struct {
		outputMsgs chan<- models.Message
		message    models.Message
		hitRule    chan<- models.Rule
		rules      map[string]models.Rule
		bot        *models.Bot
	}

	testBot := new(models.Bot)
	testMessage := models.Message{
		BotMentioned: true,
	}

	testRules := map[string]models.Rule{
		"test": {
			Name:          "testRule",
			Active:        true,
			IncludeInHelp: true,
			HelpText:      "testing",
		},
	}

	testBotCustomHelp := new(models.Bot)
	testBotCustomHelp.CustomHelpText = "This is help, foo. \n"

	tests := []struct {
		name         string
		args         args
		wantHelpText string
	}{
		{"Default help - no rules", args{message: testMessage, bot: testBot}, "I understand these commands: \n"},
		{"Custom help intro", args{message: testMessage, bot: testBotCustomHelp}, "This is help, foo. \n"},
		{"1 Rule", args{message: testMessage, bot: testBot, rules: testRules}, fmt.Sprintf("I understand these commands: \n\n • %s", testRules["test"].HelpText)},
		{"Custom help intro + 1 Rule", args{message: testMessage, bot: testBotCustomHelp, rules: testRules}, "This is help, foo. \n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testOutput := make(chan models.Message, 1)
			testHitRule := make(chan models.Rule, 1)
			tt.args.outputMsgs = testOutput
			tt.args.hitRule = testHitRule

			handleNoMatch(tt.args.outputMsgs, tt.args.message, tt.args.hitRule, tt.args.rules, tt.args.bot)
			output := <-testOutput
			if tt.wantHelpText != output.Output {
				t.Errorf("handleSchedulerServiceRule() wanted helpText to be = %s, but got %s", tt.wantHelpText, output.Output)
			}
		})
	}
}

func Test_doRuleActions(t *testing.T) {
	type args struct {
		message    models.Message
		outputMsgs chan<- models.Message
		rule       models.Rule
		hitRule    chan<- models.Rule
		bot        *models.Bot
	}

	testBot := new(models.Bot)

	testMessage := models.Message{
		Input:        "foo bar",
		BotMentioned: true,
		Vars:         make(map[string]string),
	}

	testAction := models.Action{
		Name: "message action",
		Type: "message",
	}

	testRule := models.Rule{
		Active: true,
		Actions: []models.Action{
			testAction,
		},
		Respond:      "foo",
		FormatOutput: "hi there from foo action",
	}

	execAction := models.Action{
		Name: "exec action",
		Type: "exec",
		Cmd:  `echo "hi"`,
	}

	execRule := models.Rule{
		Active: true,
		Actions: []models.Action{
			execAction,
		},
		Respond:      "foo",
		FormatOutput: "${_exec_output}",
	}

	failAction := models.Action{
		Name: "epic fail",
		Type: "fail",
	}

	failRule := models.Rule{
		Active: true,
		Actions: []models.Action{
			failAction,
		},
		Respond:      "foo",
		FormatOutput: "boo",
	}

	tsOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer tsOK.Close()

	httpAction := models.Action{
		Name: "http test",
		Type: "get",
		URL:  tsOK.URL,
	}

	httpRule := models.Rule{
		Active: true,
		Actions: []models.Action{
			httpAction,
		},
		Respond:      "foo",
		FormatOutput: "${_raw_http_output}",
	}

	// reactionAction := models.Action{
	// 	Name: "reaction",
	// 	Type: "get",
	// 	URL:  tsOK.URL,
	// }

	// reactionRule := models.Rule{
	// 	Active: true,
	// 	Name:   "Reaction Rule",
	// 	Actions: []models.Action{
	// 		reactionAction,
	// 	},
	// 	Reaction:     ":palm_face:",
	// 	Respond:      "foo",
	// 	FormatOutput: "${_raw_http_output}",
	// 	Remotes: models.Remotes{
	// 		Slack: models.SlackConfig{
	// 			Attachments: []slack.Attachment{},
	// 		},
	// 	},
	// }

	// testReactionMessage := models.Message{
	// 	Service:      models.MsgServiceChat,
	// 	Input:        "foo bar",
	// 	BotMentioned: true,
	// 	Vars:         make(map[string]string),
	// 	Attributes:   make(map[string]string),
	// 	ChannelID:    "DTEST",
	// 	Timestamp:    "74623874623",
	// }

	tests := []struct {
		name            string
		args            args
		expectedMessage string
	}{
		{"Missing format_output", args{message: models.Message{}, rule: models.Rule{}, bot: testBot}, "hmm, the 'format_output' field in your configuration is empty"},
		{"Message Action", args{message: testMessage, rule: testRule, bot: testBot}, "hi there from foo action"},
		{"Exec Action", args{message: testMessage, rule: execRule, bot: testBot}, "hi"},
		{"Http Action", args{message: testMessage, rule: httpRule, bot: testBot}, "OK"},
		// {"Reaction Action", args{message: testReactionMessage, rule: reactionRule, bot: testBot}, "OK"},
		{"Fail Action", args{message: testMessage, rule: failRule, bot: testBot}, "boo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testOutput := make(chan models.Message, 1)
			testHitRule := make(chan models.Rule, 1)
			tt.args.hitRule = testHitRule
			tt.args.outputMsgs = testOutput

			doRuleActions(tt.args.message, tt.args.outputMsgs, tt.args.rule, tt.args.hitRule, tt.args.bot)
			output := <-testOutput

			if output.Output != tt.expectedMessage {
				t.Errorf("Message expected to be: %s, but got: %s", tt.expectedMessage, output.Output)
			}
		})
	}
}

func Test_matcherLoop(t *testing.T) {
	type args struct {
		message    models.Message
		outputMsgs chan<- models.Message
		rules      map[string]models.Rule
		hitRule    chan<- models.Rule
		bot        *models.Bot
	}

	testBot := new(models.Bot)
	testRules := make(map[string]models.Rule)
	testRule := models.Rule{}
	testRules["test"] = testRule
	testMessage := models.Message{Input: "Hi there!", BotMentioned: true}

	testMessage2 := models.Message{
		Service:      models.MsgServiceChat,
		Input:        "foo test",
		BotMentioned: true,
		Vars:         make(map[string]string),
	}
	testRules2 := make(map[string]models.Rule)
	testRule2 := models.Rule{
		Active:        true,
		Respond:       "foo",
		Args:          []string{"arg1"},
		IncludeInHelp: true,
		HelpText:      "foo arg1",
		FormatOutput:  "output is foo ${arg1}",
	}
	testRules2["test"] = testRule2

	testMsgAttributes3 := make(map[string]string)
	testMsgAttributes3["from_schedule"] = "test-schedule"
	testMessage3 := models.Message{
		Service:      models.MsgServiceScheduler,
		Input:        "@every 5s",
		Attributes:   testMsgAttributes3,
		BotMentioned: true,
		Vars:         make(map[string]string),
	}
	testRules3 := make(map[string]models.Rule)
	testRule3 := models.Rule{
		Active:        true,
		Schedule:      "@every 5s",
		Name:          "test-schedule",
		Args:          []string{},
		IncludeInHelp: true,
		HelpText:      "foo arg1",
		FormatOutput:  "Hello, from Scheduler 1!",
	}
	testRules3["test"] = testRule3

	tests := []struct {
		name           string
		args           args
		expectedOutput string
	}{
		{"No Rule Match", args{message: testMessage, rules: testRules, bot: testBot}, "I understand these commands: \n"},
		{"Chat rule, no actions", args{message: testMessage2, rules: testRules2, bot: testBot}, "output is foo test"},
		{"Scheduler rule, no actions", args{message: testMessage3, rules: testRules3, bot: testBot}, "Hello, from Scheduler 1!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testOutput := make(chan models.Message, 1)
			testHitRule := make(chan models.Rule, 1)

			tt.args.outputMsgs = testOutput
			tt.args.hitRule = testHitRule

			matcherLoop(tt.args.message, tt.args.outputMsgs, tt.args.rules, tt.args.hitRule, tt.args.bot)

			output := <-testOutput
			if output.Output != tt.expectedOutput {
				t.Errorf("Message expected to be: %s, but got: %s", tt.expectedOutput, output.Output)
			}
		})
	}
}
