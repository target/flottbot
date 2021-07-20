package core

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"html/template"
	"strconv"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/mohae/deepcopy"

	"github.com/target/flottbot/handlers"
	"github.com/target/flottbot/models"
	"github.com/target/flottbot/utils"
)

// Matcher will search through the map of loaded rules, determine if a rule was hit, and process said rule to be sent out as a message
func Matcher(inputMsgs <-chan models.Message, outputMsgs chan<- models.Message, rules map[string]models.Rule, hitRule chan<- models.Rule, bot *models.Bot) {
	for {
		message := <-inputMsgs
		matcherLoop(message, outputMsgs, rules, hitRule, bot)
	}
}

func matcherLoop(message models.Message, outputMsgs chan<- models.Message, rules map[string]models.Rule, hitRule chan<- models.Rule, bot *models.Bot) {
	match := false

RuleSearch:
	// Look through rules to see if we can find a match
	for _, rule := range rules {
		// Only check active rules.
		if rule.Active {
			// Init some variables for use below
			processedInput, hit := getProccessedInputAndHitValue(message.Input, rule.Respond, rule.Hear)
			// Determine what service we are processing the rule for
			switch message.Service {
			case models.MsgServiceChat, models.MsgServiceCLI:
				foundMatch, stopSearch := handleChatServiceRule(outputMsgs, message, hitRule, rule, processedInput, hit, bot)
				match = foundMatch
				if stopSearch {
					break RuleSearch
				}
			case models.MsgServiceScheduler:
				foundMatch, stopSearch := handleSchedulerServiceRule(outputMsgs, message, hitRule, rule, bot)
				match = foundMatch
				if stopSearch {
					break RuleSearch
				}
			}
		}
	}
	// No rule was matched
	if !match {
		handleNoMatch(outputMsgs, message, hitRule, rules, bot)
	}
}

// getProccessedInputAndHitValue gets the processed input from the message input and the true/false if it was a successfully hit rule
func getProccessedInputAndHitValue(messageInput, ruleRespondValue, ruleHearValue string) (string, bool) {
	processedInput, hit := "", false
	if ruleRespondValue != "" {
		processedInput, hit = utils.Match(ruleRespondValue, messageInput, true)
	} else if ruleHearValue != "" { // Are we listening to everything?
		_, hit = utils.Match(ruleHearValue, messageInput, false)
	}
	return processedInput, hit
}

// handleChatServiceRule handles the processing logic for a rule that came from either the chat application or CLI remote
func handleChatServiceRule(outputMsgs chan<- models.Message, message models.Message, hitRule chan<- models.Rule, rule models.Rule, processedInput string, hit bool, bot *models.Bot) (bool, bool) {
	match, stopSearch := false, false
	if rule.Respond != "" || rule.Hear != "" {
		// You can only use 'respond' OR 'hear'
		if rule.Respond != "" && rule.Hear != "" {
			bot.Log.Debug().Msgf("rule '%s' has both 'hear' and 'match' or 'respond' defined. please choose one or the other", rule.Name)
		}
		// Args are not implemented for 'hear'
		if rule.Hear != "" && len(rule.Args) > 0 {
			bot.Log.Debug().Msgf("rule '%s' has both 'args' and 'hear' set. to use 'args', use 'respond' instead of 'hear'", rule.Name)
		}

		if hit && message.ThreadTimestamp != "" && rule.IgnoreThreads {
			bot.Log.Debug().Msg("response suppressed due to 'ignore_threads' being set")
			return true, true
		}

		// if it's a 'respond' rule, make sure the bot was mentioned
		if hit && rule.Respond != "" && !message.BotMentioned && message.Type != models.MsgTypeDirect {
			return match, stopSearch
		}

		if hit {
			bot.Log.Info().Msgf("found rule match '%s' for input '%s'", rule.Name, message.Input)
			// Don't go through more rules if rule is matched
			match, stopSearch = true, true
			// Publish metric to prometheus - metricname will be combination of bot name and rule name
			Prommetric(bot.Name+"-"+rule.Name, bot)
			// Capture untouched user input

			message.Vars["_raw_user_input"] = message.Input
			message.Vars["_is_thread_message"] = strconv.FormatBool(message.ThreadTimestamp != "")

			// Do additional checks on the rule before running
			if !isValidHitChatRule(&message, rule, processedInput, bot) {
				outputMsgs <- message
				hitRule <- models.Rule{}
				// prevent actions from being run; exit early
				return match, stopSearch
			}
			msg := deepcopy.Copy(message).(models.Message)
			go doRuleActions(msg, outputMsgs, rule, hitRule, bot)
			return match, stopSearch
		}
	}
	return match, stopSearch
}

// handleSchedulerServiceRule handles the processing logic for a rule that came from the Scheduler remote
func handleSchedulerServiceRule(outputMsgs chan<- models.Message, message models.Message, hitRule chan<- models.Rule, rule models.Rule, bot *models.Bot) (bool, bool) {
	match, stopSearch := false, false
	if rule.Schedule != "" && rule.Name == message.Attributes["from_schedule"] {
		match, stopSearch = true, true // Don't go through more rules if rule is matched
		msg := deepcopy.Copy(message).(models.Message)
		go doRuleActions(msg, outputMsgs, rule, hitRule, bot)
		return match, stopSearch
	}
	return match, stopSearch
}

// handleNoMatch - handles logic for unmatched rule
func handleNoMatch(outputMsgs chan<- models.Message, message models.Message, hitRule chan<- models.Rule, rules map[string]models.Rule, bot *models.Bot) {
	// If bot was addressed or was private messaged, print help text by default
	if message.Type == models.MsgTypeDirect || message.BotMentioned {
		// Do not send help message if DisableNoMatchHelp is true
		if !bot.DisableNoMatchHelp {
			bot.Log.Info().Msg("bot was addressed, but no rule matched - showing help")
			// Publish metric as none
			Prommetric(bot.Name+"-None", bot)
			// Set custom_help_text if it is set in bot.yml
			helpMsg := bot.CustomHelpText
			// If custom_help_text is not set, use default Help Text, for each rule use help_text from rule file
			if helpMsg == "" {
				helpMsg = "I understand these commands: \n"
				// Go through all the rules and collect the help_text
				for _, rule := range rules {
					// Is the rule active and does the user want to expose the help for it? 'hear' rules don't show in help by default
					if rule.Active && rule.Hear == "" && rule.IncludeInHelp && rule.HelpText != "" {
						helpMsg = helpMsg + fmt.Sprintf("\n • %s", rule.HelpText)
					}
				}
			}
			// Populate output with help text defined above
			message.Output = helpMsg
			outputMsgs <- message
			hitRule <- models.Rule{}
		}
	}
}

// isValidHitChatRule does additional checks on a successfully hit rule that came from the chat or CLI service
func isValidHitChatRule(message *models.Message, rule models.Rule, processedInput string, bot *models.Bot) bool {
	// Check to honor allow_users or allow_usergroups
	canRunRule := utils.CanTrigger(message.Vars["_user.name"], message.Vars["_user.id"], rule, bot)
	if !canRunRule {
		message.Output = fmt.Sprintf("You are not allowed to run the '%s' rule.", rule.Name)
		// forcing direct message
		// message.DirectMessageOnly = true
		message.Type = models.MsgTypeDirect
		message.IsEphemeral = true
		return false
	}
	// If this wasn't a 'hear' rule, handle the args
	if rule.Hear == "" {
		// Get all the args that the message sender supplied
		args := utils.RuleArgTokenizer(processedInput)
		var optionalArgs int
		var requiredArgs int
		var varArgs int
		// take note of all optional args that end with a '?'
		for _, arg := range rule.Args {
			if strings.HasSuffix(arg, "?") {
				optionalArgs++
			}
			if strings.HasSuffix(arg, "+") {
				varArgs++
			}
		}
		if varArgs > 1 {
			// error, can ony have 1
			message.Output = "you cannot specify more than 1 variable argument"
			return false
		}
		if len(rule.Args) > 0 && strings.HasSuffix(rule.Args[len(rule.Args)-1], "+") {
			if optionalArgs > 0 {
				// error, cannot combine optional and varargs
				message.Output = "you cannot combine optional arguments with variable arguments"
				return false
			}
		} else if varArgs == 1 {
			// error, vararg but not in last position
			message.Output = "you must specify the variable argument in the last argument position"
			return false
		}
		// ensure we only require args that don't end with '?'
		requiredArgs = len(rule.Args) - optionalArgs
		// Are we expecting a number of args but don't have as many as the rule defines? Send a helpful message
		if len(rule.Args) > 0 && requiredArgs > len(args) {
			message.Output = fmt.Sprintf("you might be missing an argument or two - this is what i'm looking for\n```%s```", rule.HelpText)
			return false
		}
		// Go through the supplied args and make them available as variables
		for index, arg := range rule.Args {
			// If this is a varag method, then join all the remaining args and end
			if strings.HasSuffix(arg, "+") {
				message.Vars[strings.TrimSuffix(arg, "+")] = strings.Join(args[index:], " ")
				break
			}
			// strip '?' from end of arg
			arg = strings.TrimSuffix(arg, "?")
			// index starts at 0 so we need to account for that
			if index > (len(args) - 1) {
				message.Vars[arg] = ""
			} else {
				message.Vars[arg] = args[index]
			}
		}
	}
	return true
}

// core handler routing for all allowed actions
func doRuleActions(message models.Message, outputMsgs chan<- models.Message, rule models.Rule, hitRule chan<- models.Rule, bot *models.Bot) {
	// React to message which triggered rule
	if rule.Reaction != "" {
		copyrule := deepcopy.Copy(rule).(models.Rule)
		copymessage := deepcopy.Copy(message).(models.Message)
		handleReaction(outputMsgs, &copymessage, hitRule, copyrule)
	}

	// Deal with the actions associated with the rule asynchronously
	for _, action := range rule.Actions {
		var err error

		switch strings.ToLower(action.Type) {
		// HTTP actions.
		case "get", "post", "put":
			bot.Log.Debug().Msgf("executing action '%s'...", action.Name)
			err = handleHTTP(action, &message, bot)
		// Exec (script) actions
		case "exec":
			bot.Log.Debug().Msgf("executing action '%s'...", action.Name)
			err = handleExec(action, &message, bot)
		// Normal message/log actions
		case "message", "log":
			bot.Log.Debug().Msgf("executing action '%s'...", action.Name)
			// Log actions cannot direct message users by default
			directive := rule.DirectMessageOnly
			if action.Type == "log" {
				directive = false
			}
			// Create copy of message so as to not overwrite other message action type messages
			copy := deepcopy.Copy(message).(models.Message)
			err = handleMessage(action, outputMsgs, &copy, directive, rule.StartMessageThread, hitRule, bot)
		// Fallback to error if action type is invalid
		default:
			bot.Log.Error().Msgf("the rule '%s' of type '%s' is not a supported action", action.Name, action.Type)
		}

		// Handle reaction update
		updateReaction(action, &rule, message.Vars, bot)

		// Handle error
		if err != nil {
			bot.Log.Error().Msg(err.Error())
		}
	}

	// Match supplied room names to IDs
	message.OutputToRooms = utils.GetRoomIDs(rule.OutputToRooms, bot)

	// Populate message output to users
	message.OutputToUsers = rule.OutputToUsers

	// Start a thread if the message is not already part of a thread and
	// start_message_thread was set for the Rule
	if rule.StartMessageThread && message.ThreadTimestamp == "" {
		message.ThreadTimestamp = message.Timestamp
	}

	// After running through all the actions, compose final message
	val, err := craftResponse(rule, message, bot)
	if err != nil {
		bot.Log.Error().Msg(err.Error())
		message.Output = err.Error()
		outputMsgs <- message
	} else {
		message.Output = val
		// Override out with an error message, if one was set
		if message.Error != "" {
			message.Output = message.Error
		}
		// Pass along whether the message should be a direct message
		message.DirectMessageOnly = rule.DirectMessageOnly
		outputMsgs <- message
	}
	// Channel completed rule
	hitRule <- rule
}

// craftResponse handles format_output to make the final message from the bot user-friendly
func craftResponse(rule models.Rule, msg models.Message, bot *models.Bot) (string, error) {
	// The user removed the 'format_output' field, or it's not set
	if rule.FormatOutput == "" {
		return "", errors.New("hmm, the 'format_output' field in your configuration is empty")
	}

	// None of the rooms specified in 'output_to_rooms' exist
	if !rule.DirectMessageOnly && len(rule.OutputToRooms) > 0 && len(msg.OutputToRooms) == 0 {
		msg := fmt.Sprintf("Could not find any of the rooms specified in 'output_to_rooms' while 'direct_message_only' is set to false. "+
			"Please check rule '%s'", rule.Name)
		if len(rule.OutputToUsers) == 0 {
			return "", errors.New(msg)
		}
		bot.Log.Warn().Msg(msg)
	}

	// Simple warning that we will ignore 'output_to_rooms' when 'direct_message_only' is set
	if rule.DirectMessageOnly && len(rule.OutputToRooms) > 0 {
		bot.Log.Debug().Msgf("the rule '%s' has 'direct_message_only' set, 'output_to_rooms' will be ignored", rule.Name)
	}

	// Use FormatOutput as source for output and find variables and replace content the variable exists
	output, err := utils.Substitute(rule.FormatOutput, msg.Vars)

	// Check if the value contains html/template code, for advanced formatting
	if strings.Contains(output, "{{") {
		var i interface{}

		t, err := template.New("output").Funcs(sprig.FuncMap()).Parse(output)
		if err != nil {
			return "", err
		}
		buf := new(bytes.Buffer)

		err = t.Execute(buf, i)
		if err != nil {
			return "", err
		}

		output = buf.String()
	}

	return output, err
}

// Handle script execution actions
func handleExec(action models.Action, msg *models.Message, bot *models.Bot) error {
	if action.Cmd == "" {
		return fmt.Errorf("no command was supplied for the '%s' action named: %s", action.Type, action.Name)
	}

	resp, err := handlers.ScriptExec(action, msg, bot)

	// Set explicit variables to make script output, script status code accessible in rules
	msg.Vars["_exec_output"] = resp.Output
	msg.Vars["_exec_status"] = strconv.Itoa(resp.Status)

	if err != nil {
		return err
	}

	return nil
}

// Handle HTTP call actions
func handleHTTP(action models.Action, msg *models.Message, bot *models.Bot) error {
	if action.URL == "" {
		return fmt.Errorf("no URL was supplied for the '%s' action named: %s", action.Type, action.Name)
	}

	resp, err := handlers.HTTPReq(action, msg, bot)
	if err != nil {
		msg.Error = fmt.Sprintf("error in request made by action '%s' - see bot admin for more information", action.Name)
		return err
	}

	// Just a friendly debugger warning on failed requests
	if resp.Status >= 400 {
		bot.Log.Debug().Msgf("error in request made by action '%s' - '%s' returned '%d' with response: %s", action.Name, action.URL, resp.Status, resp.Raw)
	}

	// Always store raw response
	bot.Log.Debug().Msgf("successfully executed action '%s'", action.Name)
	// Set explicit variables to make raw response output, http status code accessible in rules
	msg.Vars["_raw_http_output"] = resp.Raw
	msg.Vars["_raw_http_status"] = strconv.Itoa(resp.Status)

	// Do we need to expose any fields?
	if len(action.ExposeJSONFields) > 0 {
		for k, v := range action.ExposeJSONFields {
			var t *template.Template

			v, err = utils.Substitute(v, msg.Vars)
			if err != nil {
				return err
			}

			// Check if the value contains html/template code
			if strings.Contains(v, "{{") {
				t, err = template.New(k).Funcs(sprig.FuncMap()).Parse(v)
			} else {
				t, err = template.New(k).Funcs(sprig.FuncMap()).Parse(fmt.Sprintf(`{{%s}}`, v))
			}
			if err != nil {
				return err
			}

			buf := new(bytes.Buffer)

			err := t.Execute(buf, resp.Data)
			if err != nil {
				return err
			}

			msg.Vars[k] = html.UnescapeString(buf.String())
		}
	}

	return nil
}

// Handle standard message/logging actions
func handleMessage(action models.Action, outputMsgs chan<- models.Message, msg *models.Message, direct, startMsgThread bool, hitRule chan<- models.Rule, bot *models.Bot) error {
	if action.Message == "" {
		return fmt.Errorf("no message was set")
	}

	if action.Type == "message" && startMsgThread && msg.ThreadTimestamp == "" {
		msg.ThreadTimestamp = msg.Timestamp
	}

	// Get message output from action
	output, err := utils.Substitute(action.Message, msg.Vars)
	if err != nil {
		return err
	}

	msg.Output = output
	// Send to desired room(s)
	if direct && len(action.LimitToRooms) > 0 { // direct=true and limit_to_rooms is specified
		bot.Log.Debug().Msgf("'direct_message_only' is set - 'limit_to_rooms' field on the '%s' action will be ignored", action.Name)
	} else if !direct && len(action.LimitToRooms) > 0 { // direct=false and limit_to_rooms is specified
		msg.OutputToRooms = utils.GetRoomIDs(action.LimitToRooms, bot)

		if len(msg.OutputToRooms) == 0 {
			return errors.New("the rooms defined in 'limit_to_rooms' do not exist")
		}
	} else if !direct && len(action.LimitToRooms) == 0 { // direct=false and no limit_to_rooms is specified
		msg.OutputToRooms = []string{msg.ChannelID}
	}
	// Else: direct=true and no limit_to_rooms is specified

	// Set message directive
	msg.DirectMessageOnly = direct
	// Send out message
	outputMsgs <- *msg
	hitRule <- models.Rule{}
	return nil
}

// Handle initial emoji reaction when rule is matched
func handleReaction(outputMsgs chan<- models.Message, msg *models.Message, hitRule chan<- models.Rule, rule models.Rule) {
	outputMsgs <- *msg
	hitRule <- rule
}

// Update emoji reaction when specified
func updateReaction(action models.Action, rule *models.Rule, vars map[string]string, bot *models.Bot) {
	if action.Reaction != "" && rule.Reaction != "" {
		// Check if the value contains html/template code
		if strings.Contains(action.Reaction, "{{") {
			reaction, err := utils.Substitute(action.Reaction, vars)
			if err != nil {
				bot.Log.Error().Msg(err.Error())
				return
			}
			action.Reaction = reaction

			var t *template.Template
			var i interface{}

			t, err = template.New("update_reaction").Funcs(sprig.FuncMap()).Parse(action.Reaction)
			if err != nil {
				bot.Log.Error().Msgf("failed to update reaction '%s'", rule.Reaction)
				return
			}
			buf := new(bytes.Buffer)

			err = t.Execute(buf, i)
			if err != nil {
				return
			}
			rule.RemoveReaction = rule.Reaction
			action.Reaction = buf.String()
			action.Reaction = strings.TrimSpace(action.Reaction)
			rule.Reaction = action.Reaction
		} else {
			rule.RemoveReaction = rule.Reaction
			rule.Reaction = action.Reaction
		}
	}
}
