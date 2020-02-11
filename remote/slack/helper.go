package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
	"github.com/target/flottbot/models"
	"github.com/target/flottbot/utils"
)

/*
======================================================================
Slack helper functions (anything that uses the 'nlopes/slack' package)
======================================================================
*/

// constructInteractiveComponentMessage creates a message specifically for a matched rule from the Interactive Components server
func constructInteractiveComponentMessage(callback slack.AttachmentActionCallback, bot *models.Bot) models.Message {
	text := ""
	if len(callback.ActionCallback.AttachmentActions) > 0 {
		for _, action := range callback.ActionCallback.AttachmentActions {
			if action.Value != "" {
				text = fmt.Sprintf("<@%s> %s", bot.ID, action.Value)
				break
			}
		}
	}
	message := models.NewMessage()
	messageType, err := getMessageType(callback.Channel.ID)
	if err != nil {
		bot.Log.Debug(err.Error())
	}

	userNames := strings.Split(callback.User.Name, ".")
	user := &slack.User{
		ID:       callback.User.ID,
		TeamID:   callback.User.TeamID,
		Name:     callback.User.Name,
		Color:    callback.User.Color,
		RealName: callback.User.RealName,
		TZ:       callback.User.TZ,
		TZLabel:  callback.User.TZLabel,
		TZOffset: callback.User.TZOffset,
		Profile: slack.UserProfile{
			FirstName:             userNames[0],
			LastName:              userNames[len(userNames)-1],
			RealNameNormalized:    callback.User.Profile.RealNameNormalized,
			DisplayName:           callback.User.Profile.DisplayName,
			DisplayNameNormalized: callback.User.Profile.DisplayName,
			Email:                 callback.User.Profile.Email,
			Skype:                 callback.User.Profile.Skype,
			Phone:                 callback.User.Profile.Phone,
			Title:                 callback.User.Profile.Title,
			StatusText:            callback.User.Profile.StatusText,
			StatusEmoji:           callback.User.Profile.StatusEmoji,
			Team:                  callback.User.Profile.Team,
		},
	}
	channel := callback.Channel.Name
	if callback.Channel.IsPrivate {
		channel = callback.Channel.ID
	}

	msgType, err := getMessageType(callback.Channel.ID)
	if err != nil {
		bot.Log.Debug(err.Error())
	}

	if msgType == models.MsgTypePrivateChannel {
		channel = callback.Channel.ID
	}
	contents, mentioned := removeBotMention(text, bot.ID)

	return populateMessage(message, messageType, channel, contents, callback.MessageTs, callback.MessageTs, "", mentioned, user, bot)
}

// getEventsAPIHealthHandler creates and returns the handler for health checks on the Slack Events API reader
func getEventsAPIHealthHandler(bot *models.Bot) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			bot.Log.Errorf("getEventsAPIHealthHandler: Received invalid method: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		bot.Log.Info("Bot event health endpoint hit!")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func sendHTTPResponse(status int, contentType string, message string, w http.ResponseWriter, r *http.Request) {
	if contentType == "" {
		contentType = "text/plain"
	}
	w.WriteHeader(status)
	w.Header().Set("Content-Type", contentType)
	w.Write([]byte(message))
}

func handleURLVerification(body string, w http.ResponseWriter, r *http.Request) {
	var slackResponse *slackevents.ChallengeResponse
	statusCode := http.StatusOK
	err := json.Unmarshal([]byte(body), &slackResponse)
	if err != nil {
		statusCode = http.StatusInternalServerError
	}

	sendHTTPResponse(statusCode, "", slackResponse.Challenge, w, r)
}

func handleCallBack(api *slack.Client, event slackevents.EventsAPIInnerEvent, bot *models.Bot, inputMsgs chan<- models.Message, w http.ResponseWriter, r *http.Request) {
	// write back to the event to ensure the event does not trigger again
	sendHTTPResponse(http.StatusOK, "", "{}", w, r)

	// process the event
	bot.Log.Debugf("getEventsAPIEventHandler: Received event '%s'", event.Type)
	switch ev := event.Data.(type) {
	// There are Events API specific MessageEvents
	// https://api.slack.com/events/message.channels
	case *slackevents.MessageEvent:
		senderID := ev.User
		// Only process messages that aren't from the bot itself
		if senderID != "" && bot.ID != senderID {
			channel := ev.Channel
			msgType, err := getMessageType(channel)
			if err != nil {
				bot.Log.Debug(err.Error())
			}
			text, mentioned := removeBotMention(ev.Text, bot.ID)
			user, err := api.GetUserInfo(senderID)
			if err != nil && senderID != "" { // we only care if senderID is not empty and there's an error (senderID == "" could be a thread from a message)
				bot.Log.Errorf("getEventsAPIEventHandler: Did not get Slack user info: %s", err.Error())
			}
			timestamp := ev.TimeStamp
			threadTimestamp := ev.ThreadTimeStamp

			link, err := api.GetPermalink(&slack.PermalinkParameters{Channel: channel, Ts: timestamp})
			if err != nil {
				link = ""
			}
			inputMsgs <- populateMessage(models.NewMessage(), msgType, channel, text, timestamp, threadTimestamp, link, mentioned, user, bot)
		}
	// This is an Event shared between RTM and the Events API
	case *slack.MemberJoinedChannelEvent:
		// get bot rooms
		bot.Rooms = getRooms(api)
		bot.Log.Debugf("%s has joined the channel %s", bot.Name, bot.Rooms[ev.Channel])
	case *slack.MemberLeftChannelEvent:
		// remove room
		delete(bot.Rooms, ev.Channel)
		bot.Log.Debugf("%s has left the channel %s", bot.Name, bot.Rooms[ev.Channel])
	default:
		bot.Log.Errorf("getEventsAPIEventHandler: Unrecognized event type: %v", ev)
	}
}

// getEventsAPIEventHandler creates and returns the handler for events coming from the the Slack Events API reader
func getEventsAPIEventHandler(api *slack.Client, vToken string, inputMsgs chan<- models.Message, bot *models.Bot) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			bot.Log.Errorf("Slack API Server: invalid method %s", r.Method)
			message := fmt.Sprintf("Oops! I encountered an unexpected HTTP request method: %s. It should be POST.", r.Method)
			sendHTTPResponse(http.StatusMethodNotAllowed, "", message, w, r)
			return
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		body := buf.String()

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: vToken}))
		if err != nil {
			bot.Log.Errorf("Slack API Server: There was an error reading an event: %s", err)
			sendHTTPResponse(http.StatusInternalServerError, "", "Oops! There was an error with the Slack events API", w, r)
			return
		}

		// accept challenge response
		if eventsAPIEvent.Type == slackevents.URLVerification {
			bot.Log.Debug("Slack API Server: Received Slack challenge request. Sending challenge acceptance.")
			handleURLVerification(body, w, r)
		}

		// process the event
		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			handleCallBack(api, eventsAPIEvent.InnerEvent, bot, inputMsgs, w, r)
		}
	}
}

// getInteractiveComponentHealthHandler creates and returns the handler for health checks on the Interactive Component server
func getInteractiveComponentHealthHandler(bot *models.Bot) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			bot.Log.Errorf("getInteractiveComponentHealthHandler: Received invalid method %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		bot.Log.Info("Bot interaction health endpoint hit!")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

// getInteractiveComponentRuleHandler creates and returns the handler for processing and sending out messages from the Interactive Component server
func getInteractiveComponentRuleHandler(verificationToken string, inputMsgs chan<- models.Message, message *models.Message, rule models.Rule, bot *models.Bot) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			bot.Log.Errorf("getInteractiveComponentRuleHandler: Received invalid method: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(fmt.Sprintf("Oops! I encountered an unexpected HTTP request method: %s", r.Method)))
			return
		}

		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			bot.Log.Errorf("getInteractiveComponentRuleHandler: Failed to read request body: %s", err.Error())
		}

		contents, err := sanitizeContents(buff)
		if err != nil {
			bot.Log.Debugf("failed to sanitize content: %v", err)
		}

		var callback slack.AttachmentActionCallback
		if err := json.Unmarshal([]byte(contents), &callback); err != nil {
			bot.Log.Errorf("getInteractiveComponentRuleHandler: Failed to decode callback json\n %s\n because %s", contents, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Oops! Looks like I failed to decode some JSON in the backend. Please contact admins for more info!"))
			return
		}

		// Only accept message from slack with valid token
		if callback.Token != verificationToken {
			bot.Log.Errorf("getInteractiveComponentRuleHandler: Invalid token %s", callback.Token)
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Sorry, but I didn't recognize your verification token! Perhaps check if it's a valid token."))
			return
		}

		// Construct and send out message
		message := constructInteractiveComponentMessage(callback, bot)
		inputMsgs <- message

		// Respond
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Rodger that!"))

		bot.Log.Debugf("getInteractiveComponentRuleHandler: triggering rule: %s", rule.Name)
	}
}

// getRooms - return a map of rooms
func getRooms(api *slack.Client) map[string]string {
	rooms := make(map[string]string)
	// get public channels
	channels, _ := api.GetChannels(true)
	for _, channel := range channels {
		rooms[channel.Name] = channel.ID
	}
	// get private channels
	groups, _ := api.GetGroups(true)
	for _, group := range groups {
		rooms[group.Name] = group.ID
	}
	return rooms
}

// getSlackUsers gets Slack user objects for each user listed in messages 'output_to_users' field
func getSlackUsers(api *slack.Client, message models.Message) ([]slack.User, error) {
	slackUsers := []slack.User{}
	// grab list of users to message if 'output_to_users' was specified
	if len(message.OutputToUsers) > 0 {
		res, err := api.GetUsers()
		if err != nil {
			return []slack.User{}, fmt.Errorf("did not find any users listed in 'output_to_users': %s", err.Error())
		}
		slackUsers = res
	}
	return slackUsers, nil
}

// getUserID - returns the user's Slack user ID via email
func getUserID(email string, users []slack.User, bot *models.Bot) string {
	email = strings.ToLower(email)
	for _, u := range users {
		if strings.Contains(strings.ToLower(u.Profile.Email), email) {
			return u.ID
		}
	}
	bot.Log.Errorf("Could not find user '%s'", email)
	return ""
}

// handleDirectMessage - handle sending logic for direct messages
func handleDirectMessage(api *slack.Client, message models.Message, bot *models.Bot) error {
	// Is output to rooms set?
	if len(message.OutputToRooms) > 0 {
		bot.Log.Warn("You have specified 'direct_message_only' as 'true' and provided 'output_to_rooms'." +
			" Messages will not be sent to listed rooms. If you want to send messages to these rooms," +
			" please set 'direct_message_only' to 'false'.")
	}
	// Is output to users set?
	if len(message.OutputToUsers) > 0 {
		bot.Log.Warn("You have specified 'direct_message_only' as 'true' and provided 'output_to_users'." +
			" Messages will not be sent to the listed users (other than you). If you want to send messages to other users," +
			" please set 'direct_message_only' to 'false'.")
	}
	// Respond back to user via direct message
	return sendDirectMessage(api, message.Vars["_user.id"], message)
}

// handleNonDirectMessage - handle sending logic for non direct messages
func handleNonDirectMessage(api *slack.Client, users []slack.User, message models.Message, bot *models.Bot) error {
	// 'direct_message_only' is either 'false' OR
	// 'direct_message_only' was probably never set
	// Is output to rooms set?
	if len(message.OutputToRooms) > 0 {
		for _, roomID := range message.OutputToRooms {
			err := sendChannelMessage(api, roomID, message)
			if err != nil {
				return err
			}
		}
	}
	// Is output to users set?
	if len(message.OutputToUsers) > 0 {
		for _, u := range message.OutputToUsers {
			// Get users Slack user ID
			userID := getUserID(u, users, bot)
			if userID != "" {
				// If 'direct_message_only' is 'false' but the user listed himself in the 'output_to_users'
				if userID == message.Vars["_user.id"] && !message.DirectMessageOnly {
					bot.Log.Warn("You have specified 'direct_message_only' as 'false' but listed yourself in 'output_to_users'")
				}
				// Respond back to these users via direct message
				err := sendDirectMessage(api, userID, message)
				if err != nil {
					return err
				}
			}
		}
	}
	// Was there no specified output set?
	// Send message back to original channel
	if len(message.OutputToRooms) == 0 && len(message.OutputToUsers) == 0 {
		err := sendBackToOriginMessage(api, message)
		if err != nil {
			return err
		}
	}
	return nil
}

// populateBotUsers populates slack users
func populateBotUsers(slackUsers []slack.User, bot *models.Bot) {
	if len(slackUsers) > 0 {
		users := make(map[string]string)

		for _, user := range slackUsers {
			users[user.Name] = user.ID
		}

		bot.Users = users
	}
}

// populateUserGroups populates slack user groups
func populateUserGroups(bot *models.Bot) {
	if bot.SlackWorkspaceToken != "" {
		userGroups := make(map[string]string)
		wsAPI := slack.New(bot.SlackWorkspaceToken)
		ugroups, err := wsAPI.GetUserGroups()
		if err != nil {
			bot.Log.Debugf("Unable to retrieve usergroups: %s", err.Error())
			bot.Log.Debug("Please double check your Slack Workspace token")
		}
		for _, usergroup := range ugroups {
			userGroups[usergroup.Handle] = usergroup.ID
		}
		// we don't need API anymore
		wsAPI = nil
		bot.UserGroups = userGroups
	}
}

// populateMessage - populates the 'Message' object to be passed on for processing/sending
func populateMessage(message models.Message, msgType models.MessageType, channel, text, timeStamp, threadTimestamp, link string, mentioned bool, user *slack.User, bot *models.Bot) models.Message {
	switch msgType {
	case models.MsgTypeDirect, models.MsgTypeChannel, models.MsgTypePrivateChannel:
		// Populate message attributes
		message.Type = msgType
		message.Service = models.MsgServiceChat
		message.ChannelID = channel
		message.Input = text
		message.Output = ""
		message.Timestamp = timeStamp
		message.ThreadTimestamp = threadTimestamp
		message.BotMentioned = mentioned
		message.Attributes["ws_token"] = bot.SlackWorkspaceToken
		message.SourceLink = link

		// If the message read was not a dm, get the name of the channel it came from
		if msgType != models.MsgTypeDirect {
			name, ok := findKey(bot.Rooms, channel)
			if !ok {
				bot.Log.Warnf("populateMessage: Could not find name of channel '%s'.", channel)
			}
			message.ChannelName = name
		}

		// make channel variables available
		message.Vars["_channel.id"] = message.ChannelID
		message.Vars["_channel.name"] = message.ChannelName // will be empty if it came via DM

		// make link to trigger message available
		message.Vars["_source.link"] = message.SourceLink

		// Populate message with user information (i.e. who sent the message)
		// These will be accessible on rules via ${_user.email}, ${_user.id}, etc.
		if user != nil { // nil user implies a message from an api/bot (i.e. not an actual user)
			message.Vars["_user.id"] = user.ID
			message.Vars["_user.teamid"] = user.TeamID
			message.Vars["_user.name"] = user.Name
			message.Vars["_user.color"] = user.Color
			message.Vars["_user.realname"] = user.RealName
			message.Vars["_user.tz"] = user.TZ
			message.Vars["_user.tzlabel"] = user.TZLabel
			message.Vars["_user.tzoffset"] = strconv.Itoa(user.TZOffset)
			message.Vars["_user.firstname"] = user.Profile.FirstName
			message.Vars["_user.lastname"] = user.Profile.LastName
			message.Vars["_user.realnamenormalized"] = user.Profile.RealNameNormalized
			message.Vars["_user.displayname"] = user.Profile.DisplayName
			message.Vars["_user.displaynamenormalized"] = user.Profile.DisplayNameNormalized
			message.Vars["_user.email"] = user.Profile.Email
			message.Vars["_user.skype"] = user.Profile.Skype
			message.Vars["_user.phone"] = user.Profile.Phone
			message.Vars["_user.title"] = user.Profile.Title
			message.Vars["_user.statustext"] = user.Profile.StatusText
			message.Vars["_user.statusemoji"] = user.Profile.StatusEmoji
			message.Vars["_user.team"] = user.Profile.Team
		}

		message.Debug = true // TODO: is this even needed?
		return message
	default:
		bot.Log.Debugf("Read message of unsupported type '%T'. Unable to populate message attributes", msgType)
		return message
	}
}

// processInteractiveComponentRule processes a rule that was triggered by an interactive component, e.g. Slack interactive messages
func processInteractiveComponentRule(rule models.Rule, message *models.Message, bot *models.Bot) {
	if &rule != nil {
		// Get slack attachments from hit rule and append to outgoing message
		config := rule.Remotes.Slack
		if config.Attachments != nil {
			bot.Log.Debugf("Found attachment for rule '%s'", rule.Name)
			config.Attachments[0].CallbackID = message.ID
			if len(config.Attachments[0].Actions) > 0 {
				for i, action := range config.Attachments[0].Actions {
					actionValue, err := utils.Substitute(action.Value, message.Vars)
					if err != nil {
						bot.Log.Warn(err)
					}
					config.Attachments[0].Actions[i].Value = actionValue
				}
			}
			message.Remotes.Slack.Attachments = config.Attachments
			message.IsEphemeral = true // We default Slack Message attachment's as ephemeral
		}
	}
}

// readFromEventsAPI utilizes the Slack API client to read event-based messages.
// This method of reading is preferred over the RTM method.
func readFromEventsAPI(api *slack.Client, vToken string, inputMsgs chan<- models.Message, bot *models.Bot) {
	// get the current users
	su, err := getSlackUsers(api, models.Message{})
	if err != nil {
		bot.Log.Error(err)
	}
	// populate users
	populateBotUsers(su, bot)
	// populate user groups
	populateUserGroups(bot)

	// Create router for the events server
	router := mux.NewRouter()

	// Add health check handler
	router.HandleFunc("/event_health", getEventsAPIHealthHandler(bot)).Methods("GET")

	// Add event handler
	router.HandleFunc(bot.SlackEventsCallbackPath, getEventsAPIEventHandler(api, vToken, inputMsgs, bot)).Methods("POST")

	// Start listening to Slack events
	go http.ListenAndServe(":3000", router)

	bot.Log.Infof("Slack Events API server is listening to %s", bot.SlackEventsCallbackPath)
}

// readFromRTM utilizes the Slack API client to read messages via RTM.
// This method of reading is not preferred and the event-based read should instead be used.
func readFromRTM(rtm *slack.RTM, inputMsgs chan<- models.Message, bot *models.Bot) {
	go rtm.ManageConnection()
	for {
		msg := <-rtm.IncomingEvents
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			senderID := ev.User
			// Sometimes message events in RTM don't have a User ID?
			// Also, only process messages that aren't from the bot itself
			if senderID != "" && bot.ID != senderID {
				channel := ev.Channel
				msgType, err := getMessageType(channel)
				if err != nil {
					bot.Log.Debug(err.Error())
				}
				text, mentioned := removeBotMention(ev.Text, bot.ID)
				user, err := rtm.GetUserInfo(senderID)
				if err != nil && senderID != "" { // we only care if senderID is not empty and there's an error (senderID == "" could be a thread from a message)
					bot.Log.Errorf("Did not get Slack user info: %s", err.Error())
				}
				timestamp := ev.Timestamp
				threadTimestamp := ev.ThreadTimestamp

				link, err := rtm.GetPermalink(&slack.PermalinkParameters{Channel: channel, Ts: timestamp})
				if err != nil {
					link = ""
				}

				inputMsgs <- populateMessage(models.NewMessage(), msgType, channel, text, timestamp, threadTimestamp, link, mentioned, user, bot)
			}
		case *slack.ConnectedEvent:
			// populate users
			users, err := rtm.GetUsers()
			if err != nil {
				bot.Log.Errorf("Unable to get users: %v", err)
			}
			populateBotUsers(users, bot)
			// populate user groups
			populateUserGroups(bot)
			bot.Log.Debugf("RTM connection established!")
		case *slack.GroupJoinedEvent:
			// when the bot joins a channel add it to the internal lookup
			// NOTE: looks like there is another unsupported event we could use
			//   Received unmapped event \"member_joined_channel\"
			// Maybe watch for an update to slack package for future support
			if bot.Rooms[ev.Channel.Name] == "" {
				bot.Rooms[ev.Channel.Name] = ev.Channel.ID
				bot.Log.Debugf("Joined new channel. %s(%s) added to lookup", ev.Channel.Name, ev.Channel.ID)
			}
		case *slack.HelloEvent:
			// ignore - this is the very first initial event sent when connecting to Slack
		case *slack.RTMError:
			bot.Log.Error(ev.Error())
		case *slack.ConnectionErrorEvent:
			bot.Log.Errorf("RTM connection error: %+v", ev)
		case *slack.InvalidAuthEvent:
			if !bot.CLI {
				bot.Log.Debug("Invalid Authorization. Please double check your Slack token.")
			}
		}
	}
}

// send - handles the sending logic of a message going to Slack
func send(api *slack.Client, message models.Message, bot *models.Bot) {
	users, err := getSlackUsers(api, message)
	if err != nil {
		bot.Log.Errorf("Problem sending message: %s", err.Error())
	}
	if message.DirectMessageOnly {
		err := handleDirectMessage(api, message, bot)
		if err != nil {
			bot.Log.Errorf("Problem sending message: %s", err.Error())
		}
	} else {
		err := handleNonDirectMessage(api, users, message, bot)
		if err != nil {
			bot.Log.Errorf("Problem sending message: %s", err.Error())
		}
	}
}

// sendBackToOriginMessage - sends a message back to where it came from in Slack; this is pretty much a catch-all among the other send functions
func sendBackToOriginMessage(api *slack.Client, message models.Message) error {
	return sendMessage(api, message.IsEphemeral, message.ChannelID, message.Vars["_user.id"], message.Output, message.ThreadTimestamp, message.Attributes["ws_token"], message.Remotes.Slack.Attachments)
}

// sendChannelMessage - sends a message to a Slack channel
func sendChannelMessage(api *slack.Client, channel string, message models.Message) error {
	return sendMessage(api, message.IsEphemeral, channel, message.Vars["_user.id"], message.Output, message.ThreadTimestamp, message.Attributes["ws_token"], message.Remotes.Slack.Attachments)
}

// sendDirectMessage - sends a message back to the user who dm'ed your bot
func sendDirectMessage(api *slack.Client, userID string, message models.Message) error {
	_, _, imChannelID, err := api.OpenIMChannel(userID)
	if err != nil {
		return err
	}
	return sendMessage(api, message.IsEphemeral, imChannelID, message.Vars["_user.id"], message.Output, message.ThreadTimestamp, message.Attributes["ws_token"], message.Remotes.Slack.Attachments)
}

// sendMessage - does the final send to Slack; adds any Slack-specific message parameters to the message to be sent out
func sendMessage(api *slack.Client, ephemeral bool, channel, userID, text, threadTimeStamp, wsToken string, attachments []slack.Attachment) error {
	// send ephemeral message is indicated
	if ephemeral {
		opt := slack.MsgOptionAttachments(attachments...)
		_, err := api.PostEphemeral(channel, userID, opt)
		return err
	}

	opts := []slack.MsgOption{
		slack.MsgOptionText(text, false),
		slack.MsgOptionAsUser(true),
		slack.MsgOptionAttachments(attachments...),
		slack.MsgOptionTS(threadTimeStamp),
	}
	_, _, err := api.PostMessage(channel, opts...)
	return err
}
