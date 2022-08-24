// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/utils"
)

/*
======================================================================
Slack helper functions (anything that uses the 'slack-go/slack' package)
======================================================================
*/

// constructInteractiveComponentMessage creates a message specifically for a matched rule from the Interactive Components server.
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
		log.Error().Msg(err.Error())
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
		log.Error().Msg(err.Error())
	}

	if msgType == models.MsgTypePrivateChannel {
		channel = callback.Channel.ID
	}

	contents, mentioned := removeBotMention(text, bot.ID)

	return populateMessage(message, messageType, channel, contents, callback.MessageTs, callback.MessageTs, "", mentioned, user, bot)
}

// getEventsAPIHealthHandler creates and returns the handler for health checks on the Slack Events API reader.
func getEventsAPIHealthHandler(bot *models.Bot) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Error().Msgf("received invalid method: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		log.Debug().Msg("bot event health endpoint hit")

		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Error().Msgf("failed to send health response: %v", err)
		}
	}
}

func sendHTTPResponse(status int, message string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain")

	_, err := w.Write([]byte(message))
	if err != nil {
		log.Error().Msgf("failed to send response: %v", err)
	}
}

func handleURLVerification(body []byte, w http.ResponseWriter, r *http.Request) {
	var slackResponse *slackevents.ChallengeResponse

	statusCode := http.StatusOK

	err := json.Unmarshal(body, &slackResponse)
	if err != nil {
		statusCode = http.StatusInternalServerError
	}

	sendHTTPResponse(statusCode, slackResponse.Challenge, w, r)
}

func handleCallBack(api *slack.Client, event slackevents.EventsAPIInnerEvent, bot *models.Bot, inputMsgs chan<- models.Message, w http.ResponseWriter, r *http.Request) {
	// write back to the event to ensure the event does not trigger again
	sendHTTPResponse(http.StatusOK, "{}", w, r)

	// process the event
	log.Info().Msgf("received event: %s", event.Type)

	switch ev := event.Data.(type) {
	// Ignoring app_mention events
	case *slackevents.AppMentionEvent:
	// There are Events API specific MessageEvents
	// https://api.slack.com/events/message.channels
	case *slackevents.MessageEvent:
		senderID := ev.User

		// check if message originated from a bot
		// and whether we should respond to other bot messages
		if ev.BotID != "" && bot.RespondToBots {
			// get bot information to get
			// the associated user id
			user, err := api.GetBotInfo(ev.BotID)
			if err != nil {
				log.Error().Msgf("unable to retrieve bot info for %#q", ev.BotID)

				return
			}

			// use the bot's user id as the senderID
			senderID = user.UserID
		}

		// only process messages that aren't from our bot
		if senderID != "" && bot.ID != senderID {
			channel := ev.Channel

			msgType, err := getMessageType(channel)
			if err != nil {
				log.Error().Msg(err.Error())
			}

			text, mentioned := removeBotMention(ev.Text, bot.ID)

			// get the full user object for the given ID
			user, err := api.GetUserInfo(senderID)
			if err != nil {
				log.Error().Msgf("error getting slack user info: %v", err)
			}

			timestamp := ev.TimeStamp
			threadTimestamp := ev.ThreadTimeStamp

			// get the link to the message, will be empty string if there's an error
			link, err := api.GetPermalink(&slack.PermalinkParameters{Channel: channel, Ts: timestamp})
			if err != nil {
				log.Error().Msgf("unable to retrieve link to message: %#q", err.Error())
			}

			inputMsgs <- populateMessage(models.NewMessage(), msgType, channel, text, timestamp, threadTimestamp, link, mentioned, user, bot)
		}
	case *slackevents.MemberJoinedChannelEvent:
		// limit to our bot
		if ev.User == bot.ID {
			// look up channel info, since 'ev' only gives us ID
			channel, err := api.GetConversationInfo(ev.Channel, false)
			if err != nil {
				log.Error().Msgf("unable to fetch channel info for channel joined event: %v", err)
			}

			// add the room to the lookup
			bot.Rooms[channel.Name] = channel.ID
			log.Info().Msgf("joined new channel - %s (%s) added to lookup", channel.Name, channel.ID)
		}
	default:
		log.Debug().Msgf("unrecognized event type: %v", ev)
	}
}

// getEventsAPIEventHandler creates and returns the handler for events coming from the the Slack Events API reader.
func getEventsAPIEventHandler(api *slack.Client, signingSecret string, inputMsgs chan<- models.Message, bot *models.Bot) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// silently throw away anything that's not a POST
		if r.Method != http.MethodPost {
			log.Error().Msg("slack: method not allowed")
			sendHTTPResponse(http.StatusMethodNotAllowed, "method not allowed", w, r)

			return
		}

		// read in the body of the incoming payload
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error().Msg("slack: error reading request body")
			sendHTTPResponse(http.StatusBadRequest, "error reading request body", w, r)

			return
		}

		// create a new secrets verifier with
		// the request header and signing secret
		sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			log.Error().Msg("slack: error creating secrets verifier")
			sendHTTPResponse(http.StatusBadRequest, "error creating secrets verifier", w, r)

			return
		}

		// write the request body's hash
		if _, err := sv.Write(body); err != nil {
			log.Error().Msg("slack: error while writing body")
			sendHTTPResponse(http.StatusInternalServerError, "error while writing body", w, r)

			return
		}

		// validate signing secret with computed hash
		if err := sv.Ensure(); err != nil {
			log.Error().Msg("slack: request unauthorized")
			sendHTTPResponse(http.StatusUnauthorized, "request unauthorized", w, r)

			return
		}

		// parse the event from the request
		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			log.Error().Msg("slack: error while parsing event")
			sendHTTPResponse(http.StatusInternalServerError, "error while parsing event", w, r)

			return
		}

		// validate a URLVerification event with signing secret
		if eventsAPIEvent.Type == slackevents.URLVerification {
			log.Debug().Msg("slack: received slack challenge request - sending challenge response...")

			handleURLVerification(body, w, r)
		}

		// process regular Callback events
		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			handleCallBack(api, eventsAPIEvent.InnerEvent, bot, inputMsgs, w, r)
		}
	}
}

// getInteractiveComponentHealthHandler creates and returns the handler for health checks on the Interactive Component server.
func getInteractiveComponentHealthHandler(bot *models.Bot) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Error().Msgf("received invalid method: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		log.Debug().Msg("bot interaction health endpoint hit")

		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Error().Msgf("failed to handle interactive component: %v", err)
		}
	}
}

// getInteractiveComponentRuleHandler creates and returns the handler for processing and sending out messages from the Interactive Component server.
func getInteractiveComponentRuleHandler(signingSecret string, inputMsgs chan<- models.Message, message *models.Message, rule models.Rule, bot *models.Bot) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Error().Msgf("received invalid method: %s", r.Method)

			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Header().Set("Content-Type", "text/plain")

			_, err := w.Write([]byte("Oops! I encountered an unexpected HTTP verb"))
			if err != nil {
				log.Error().Msgf("failed to send response for interactive component handler: %v", err)
			}

			return
		}

		buff, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error().Msgf("failed to read request body: %v", err)
		}

		contents, err := sanitizeContents(buff)
		if err != nil {
			log.Error().Msgf("failed to sanitize content: %v", err)
		}

		var callback slack.AttachmentActionCallback
		if err := json.Unmarshal([]byte(contents), &callback); err != nil {
			log.Error().Msgf("failed to decode callback json %#q: %v", contents, err)

			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "text/plain")

			_, err := w.Write([]byte("Oops! Looks like I failed to decode some JSON in the backend. Please contact admins for more info!"))
			if err != nil {
				log.Error().Msgf("failed to send response for error during unmarshal process: %v", err)
			}

			return
		}

		// Only accept message from slack with valid token
		if callback.Token != bot.SlackSigningSecret {
			log.Error().Msg("invalid 'slack_signing_secret'")

			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "text/plain")

			_, err := w.Write([]byte("Sorry, but I didn't recognize your signing secret! Perhaps check if it's a valid secret."))
			if err != nil {
				log.Error().Msg("failed to send response for validating secret.")
			}

			return
		}

		// Construct and send out message
		message := constructInteractiveComponentMessage(callback, bot)
		inputMsgs <- message

		// Respond
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")

		_, err = w.Write([]byte("Rodger that!"))
		if err != nil {
			log.Error().Msgf("failed to send response: %v", err)
		}

		log.Info().Msgf("triggering rule: %s", rule.Name)
	}
}

// getRooms - return a map of rooms.
func getRooms(api *slack.Client) map[string]string {
	rooms := make(map[string]string)

	// we're getting all channel types by default
	// this can be controlled with permission scopes in Slack:
	// channels:read, groups:read, im:read, mpim:read
	cp := slack.GetConversationsParameters{
		Cursor:          "",
		ExcludeArchived: true,
		Limit:           1000, // this is the maximum value allowed
		Types:           []string{"public_channel", "private_channel", "mpim", "im"},
	}

	// there's a possibility we need to page through results
	// the results, so we're looping until there are no more pages
	for {
		channels, nc, err := api.GetConversations(&cp)
		if err != nil {
			break
		}

		// populate our channel map
		for _, channel := range channels {
			rooms[channel.Name] = channel.ID
		}

		// no more pages to process? quit the loop
		if len(nc) == 0 {
			break
		}

		// override the cursor
		cp.Cursor = nc
	}

	return rooms
}

// getSlackUsers gets Slack user objects for each user listed in messages 'output_to_users' field.
func getSlackUsers(api *slack.Client, message models.Message) ([]slack.User, error) {
	slackUsers := []slack.User{}
	// grab list of users to message if 'output_to_users' was specified
	if len(message.OutputToUsers) > 0 {
		res, err := api.GetUsers()
		if err != nil {
			return []slack.User{}, fmt.Errorf("did not find any users listed in 'output_to_users': %w", err)
		}

		slackUsers = res
	}

	return slackUsers, nil
}

// getUserID - returns the user's Slack user ID via email.
func getUserID(email string, users []slack.User, bot *models.Bot) string {
	email = strings.ToLower(email)
	for _, u := range users {
		if strings.Contains(strings.ToLower(u.Profile.Email), email) {
			return u.ID
		}
	}

	log.Error().Msgf("could not find user %#q", email)

	return ""
}

// handleDirectMessage - handle sending logic for direct messages.
func handleDirectMessage(api *slack.Client, message models.Message, bot *models.Bot) error {
	// Is output to rooms set?
	if len(message.OutputToRooms) > 0 {
		log.Warn().Msg("you have specified 'direct_message_only' as 'true' and provided 'output_to_rooms' -" +
			" messages will not be sent to listed rooms - if you want to send messages to these rooms," +
			" please set 'direct_message_only' to 'false'")
	}
	// Is output to users set?
	if len(message.OutputToUsers) > 0 {
		log.Warn().Msg("you have specified 'direct_message_only' as 'true' and provided 'output_to_users' -" +
			" messages will not be sent to the listed users (other than you) - if you want to send messages to other users," +
			" please set 'direct_message_only' to 'false'")
	}
	// Respond back to user via direct message
	return sendDirectMessage(api, message.Vars["_user.id"], message)
}

// handleNonDirectMessage - handle sending logic for non direct messages.
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
					log.Warn().Msg("you have specified 'direct_message_only' as 'false' but listed yourself in 'output_to_users'")
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

// populateUsers populates slack users.
func populateUsers(su []slack.User, bot *models.Bot) {
	users := make(map[string]string)

	// create a map of users
	for _, user := range su {
		users[user.Name] = user.ID
	}

	// add users to bot
	bot.Users = users
}

// populateUserGroups populates slack user groups.
func populateUserGroups(sug []slack.UserGroup, bot *models.Bot) {
	userGroups := make(map[string]string)

	// create a map of usergroups
	for _, usergroup := range sug {
		userGroups[usergroup.Handle] = usergroup.ID
	}

	// add usergroups to bot
	bot.UserGroups = userGroups
}

// populateMessage - populates the 'Message' object to be passed on for processing/sending.
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
		message.SourceLink = link

		// If the message read was not a dm, get the name of the channel it came from
		if msgType != models.MsgTypeDirect {
			name, ok := findKey(bot.Rooms, channel)
			if !ok {
				log.Error().Msgf("could not find name of channel %#q", channel)
			}

			message.ChannelName = name
		}

		// make channel variables available
		message.Vars["_channel.id"] = message.ChannelID
		message.Vars["_channel.name"] = message.ChannelName // will be empty if it came via DM

		// make link to trigger message available
		message.Vars["_source.link"] = message.SourceLink

		// make timestamp information available
		message.Vars["_source.timestamp"] = timeStamp
		message.Vars["_source.thread_timestamp"] = threadTimestamp

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

		return message
	default:
		log.Debug().Msgf("read message of unsupported type '%T' - unable to populate message attributes", msgType)
		return message
	}
}

// processInteractiveComponentRule processes a rule that was triggered by an interactive component, e.g. Slack interactive messages.
func processInteractiveComponentRule(rule models.Rule, message *models.Message, bot *models.Bot) {
	// Get slack attachments from hit rule and append to outgoing message
	config := rule.Remotes.Slack
	if config.Attachments != nil {
		log.Debug().Msgf("found attachment for rule %#q", rule.Name)

		config.Attachments[0].CallbackID = message.ID

		if len(config.Attachments[0].Actions) > 0 {
			for i, action := range config.Attachments[0].Actions {
				actionValue, err := utils.Substitute(action.Value, message.Vars)
				if err != nil {
					log.Warn().Msg(err.Error())
				}

				config.Attachments[0].Actions[i].Value = actionValue
			}
		}

		message.Remotes.Slack.Attachments = config.Attachments
		message.IsEphemeral = true // We default Slack Message attachment's as ephemeral
	}
}

// readFromEventsAPI utilizes the Slack API client to read event-based messages.
// This method of reading is preferred over the RTM method.
func readFromEventsAPI(api *slack.Client, vToken string, inputMsgs chan<- models.Message, bot *models.Bot) {
	// get the current users
	users, err := api.GetUsers()
	if err != nil {
		log.Error().Msgf("error getting users: %v", err)
	}

	// add users to the bot
	if users != nil {
		populateUsers(users, bot)
	}

	// get the user groups
	usergroups, err := api.GetUserGroups()
	if err != nil {
		log.Error().Msgf("error getting user groups: %v", err)
	}

	// add user groups to the bot
	if usergroups != nil {
		populateUserGroups(usergroups, bot)
	}

	// Create router for the events server
	router := mux.NewRouter()

	// Add health check handler
	router.HandleFunc("/event_health", getEventsAPIHealthHandler(bot)).Methods("GET")

	// Add event handler
	router.HandleFunc(bot.SlackEventsCallbackPath, getEventsAPIEventHandler(api, vToken, inputMsgs, bot)).Methods("POST")

	// Start listening to Slack events
	maskedPort := fmt.Sprintf(":%s", bot.SlackListenerPort)

	go func() {
		err := http.ListenAndServe(maskedPort, router)
		if err != nil {
			log.Fatal().Msg("failed to run server")
		}
	}()

	log.Info().Msgf("slack events api server is listening to %#q on port %#q",
		bot.SlackEventsCallbackPath, bot.SlackListenerPort)
}

// readFromSocketMode reads messages from Slack's Socket Mode
//
// https://api.slack.com/apis/connections/socket
//
//nolint:gocyclo,funlen // needs refactor
func readFromSocketMode(sm *slack.Client, inputMsgs chan<- models.Message, bot *models.Bot) {
	// setup the client
	client := socketmode.New(sm)

	// spawn anonymous goroutine
	go func() {
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeHello:
				// handle "hello" event
				continue
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					log.Error().Msgf("ignored: %+v", evt)

					continue
				}

				log.Debug().Msgf("event received: %+v", eventsAPIEvent)

				// acknowledge event to Slack
				client.Ack(*evt.Request)

				switch eventsAPIEvent.Type {
				case slackevents.CallbackEvent:
					innerEvent := eventsAPIEvent.InnerEvent

					switch ev := innerEvent.Data.(type) {
					case *slackevents.AppMentionEvent, *slackevents.ReactionAddedEvent:
						continue
					case *slackevents.MessageEvent:
						senderID := ev.User

						// check if message originated from a bot
						// and whether we should respond to other bot messages
						if ev.BotID != "" && bot.RespondToBots {
							// get bot information to get
							// the associated user id
							user, err := sm.GetBotInfo(ev.BotID)
							if err != nil {
								log.Error().Msgf("unable to retrieve bot info for %#q", ev.BotID)

								return
							}

							// use the bot's user id as the senderID
							senderID = user.UserID
						}

						// only process message that are not from our bot
						if senderID != "" && bot.ID != senderID {
							channel := ev.Channel

							// determine the message type
							msgType, err := getMessageType(channel)
							if err != nil {
								log.Error().Msg(err.Error())
							}

							// remove the bot mention from the user input
							text, mentioned := removeBotMention(ev.Text, bot.ID)

							// get information on the user
							user, err := sm.GetUserInfo(senderID)
							if err != nil {
								log.Error().Msgf("did not get slack user info: %s", err.Error())
							}

							timestamp := ev.TimeStamp
							threadTimestamp := ev.ThreadTimeStamp

							// get the link to the message, will be empty string if there's an error
							link, err := sm.GetPermalink(&slack.PermalinkParameters{Channel: channel, Ts: timestamp})
							if err != nil {
								log.Error().Msgf("unable to retrieve link to message: %s", err.Error())
							}

							inputMsgs <- populateMessage(models.NewMessage(), msgType, channel, text, timestamp, threadTimestamp, link, mentioned, user, bot)
						}
					case *slackevents.MemberJoinedChannelEvent:
						// limit to our bot
						if ev.User == bot.ID {
							// look up channel info, since 'ev' only gives us ID
							channel, err := sm.GetConversationInfo(ev.Channel, false)
							if err != nil {
								log.Error().Msgf("unable to fetch channel info for channel joined event: %v", err)
							}

							// add the room to the lookup
							bot.Rooms[channel.Name] = channel.ID
							log.Info().Msgf("joined new channel - %s (%s) added to lookup", channel.Name, channel.ID)
						}
					}
				default:
					log.Warn().Msgf("unsupported events api event received: %s", eventsAPIEvent.Type)
				}
			case socketmode.EventTypeConnecting:
				log.Info().Msg("connecting to slack via socket mode...")
			case socketmode.EventTypeConnectionError:
				log.Error().Msg("connection failed - retrying later...")
			case socketmode.EventTypeConnected:
				log.Info().Msg("connected to slack with socket mode")

				// get users
				users, err := sm.GetUsers()
				if err != nil {
					log.Error().Msgf("unable to get users: %v", err)
				}

				// add users to bot
				if users != nil {
					populateUsers(users, bot)
				}

				// get user groups
				usergroups, err := sm.GetUserGroups()
				if err != nil {
					log.Error().Msgf("unable to get user groups: %v", err)
				}

				// add user groups to bot
				if usergroups != nil {
					populateUserGroups(usergroups, bot)
				}
			default:
				log.Warn().Msgf("unhandled event type received: %s", evt.Type)
			}
		}
	}()

	err := client.Run()
	if err != nil {
		log.Fatal().Msgf("unable to (re)connect to Slack: %v", err)
	}
}

// send - handles the sending logic of a message going to Slack.
func send(api *slack.Client, message models.Message, bot *models.Bot) {
	users, err := getSlackUsers(api, message)
	if err != nil {
		log.Error().Msgf("problem sending message: %v", err)
	}

	if message.DirectMessageOnly {
		err := handleDirectMessage(api, message, bot)
		if err != nil {
			log.Error().Msgf("problem sending message: %v", err)
		}
	} else {
		err := handleNonDirectMessage(api, users, message, bot)
		if err != nil {
			log.Error().Msgf("problem sending message: %v", err)
		}
	}
}

// sendBackToOriginMessage - sends a message back to where it came from in Slack; this is pretty much a catch-all among the other send functions.
func sendBackToOriginMessage(api *slack.Client, message models.Message) error {
	return sendMessage(api, message.IsEphemeral, message.ChannelID, message.Vars["_user.id"], message.Output, message.ThreadTimestamp, message.Remotes.Slack.Attachments)
}

// sendChannelMessage - sends a message to a Slack channel.
func sendChannelMessage(api *slack.Client, channel string, message models.Message) error {
	return sendMessage(api, message.IsEphemeral, channel, message.Vars["_user.id"], message.Output, message.ThreadTimestamp, message.Remotes.Slack.Attachments)
}

// sendDirectMessage - sends a message back to the user who dm'ed your bot.
func sendDirectMessage(api *slack.Client, userID string, message models.Message) error {
	params := &slack.OpenConversationParameters{
		Users: []string{userID},
	}

	imChannelID, _, _, err := api.OpenConversation(params)
	if err != nil {
		return err
	}

	return sendMessage(api, message.IsEphemeral, imChannelID.ID, message.Vars["_user.id"], message.Output, message.ThreadTimestamp, message.Remotes.Slack.Attachments)
}

// sendMessage - does the final send to Slack; adds any Slack-specific message parameters to the message to be sent out.
func sendMessage(api *slack.Client, ephemeral bool, channel, userID, text, threadTimeStamp string, attachments []slack.Attachment) error {
	// prepare the message options
	opts := []slack.MsgOption{
		slack.MsgOptionText(text, false),
		slack.MsgOptionAsUser(true),
		slack.MsgOptionAttachments(attachments...),
		slack.MsgOptionTS(threadTimeStamp),
	}

	// send as ephemeral
	if ephemeral {
		_, err := api.PostEphemeral(channel, userID, opts...)
		return err
	}

	// send as regular post
	_, _, err := api.PostMessage(channel, opts...)

	return err
}
