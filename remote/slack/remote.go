// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package slack

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/remote"
)

/*
=======================================
Implementation for the Remote interface
=======================================
*/

// Client struct.
type Client struct {
	ListenerPort  string
	Token         string
	AppToken      string
	SigningSecret string
}

// validate that Client adheres to remote interface.
var _ remote.Remote = (*Client)(nil)

// instantiate a new slack client.
func (c *Client) new() *slack.Client {
	api := slack.New(c.Token)
	return api
}

type slackLogger struct {
	zerolog.Logger
}

// Name returns the name of the remote.
func (c *Client) Name() string {
	return "slack"
}

func (l *slackLogger) Output(_ int, s string) error {
	l.Logger.Info().Msg(s)
	return nil
}

// Reaction implementation to satisfy remote interface.
func (c *Client) Reaction(message models.Message, rule models.Rule, bot *models.Bot) {
	if rule.RemoveReaction != "" {
		// Init api client
		api := c.new()
		// Grab a reference to the message
		msgRef := slack.NewRefToMessage(message.ChannelID, message.Timestamp)
		// Remove bot reaction from message
		if err := api.RemoveReaction(rule.RemoveReaction, msgRef); err != nil {
			log.Error().Msgf("could not add reaction: %v", err)
			return
		}

		log.Info().Msgf("removed reaction %#q for rule %#q", rule.RemoveReaction, rule.Name)
	}

	if rule.Reaction != "" {
		// Init api client
		api := c.new()
		// Grab a reference to the message
		msgRef := slack.NewRefToMessage(message.ChannelID, message.Timestamp)
		// React with desired reaction
		if err := api.AddReaction(rule.Reaction, msgRef); err != nil {
			log.Error().Msgf("could not add reaction: %v", err)
			return
		}

		log.Info().Msgf("added reaction %#q for rule %#q", rule.Reaction, rule.Name)
	}
}

// Read implementation to satisfy remote interface
// Utilizes the Slack API client to read messages from Slack.
func (c *Client) Read(inputMsgs chan<- models.Message, rules map[string]models.Rule, bot *models.Bot) {
	// init api client
	api := c.new()

	// get bot rooms
	bot.Rooms = getRooms(api)

	// get bot id
	rat, err := api.AuthTest()
	if err != nil {
		log.Error().Msgf("the 'slack_token' is invalid or is unauthorized: %s", err)

		return
	}

	// set the bot ID
	bot.ID = rat.UserID

	if c.AppToken != "" {
		// handle Socket Mode
		// assuming Socket Mode if slack_app_token is provided
		sm := slack.New(
			bot.SlackToken,
			slack.OptionDebug(bot.Debug),
			slack.OptionAppLevelToken(c.AppToken),
			// pass our custom logger through to slack
			slack.OptionLog(&slackLogger{Logger: log.Logger.With().Str("mode", "socket").Logger()}),
		)

		// move the above inside readFromSocketMode below :o

		readFromSocketMode(sm, inputMsgs, bot)
	} else if c.SigningSecret != "" {
		// handle Events API setup
		// assuming Events API setup if slack_signing_secret is provided
		readFromEventsAPI(api, c.SigningSecret, inputMsgs, bot)
	}

	// slack is not configured correctly and cli is set to false
	// TODO: move this out of the remote setup
	if c.AppToken == "" && c.SigningSecret == "" && !bot.CLI {
		log.Error().Msg("cli mode is disabled and tokens are not set up correctly to run the bot")
	}
}

// Send implementation to satisfy remote interface.
func (c *Client) Send(message models.Message, bot *models.Bot) {
	log.Debug().Msgf("sending message %#q", message.ID)

	api := c.new()

	// check message size and trim if necessary because
	// slack messages have a hard limit of 4000 characters
	if len(message.Output) > slack.MaxMessageTextLength {
		contents := message.Output
		message.Output = contents[:(slack.MaxMessageTextLength-3)] + "..."
	}

	// Timestamp message
	message.EndTime = models.MessageTimestamp()

	// send message  based on type
	switch message.Type {
	case models.MsgTypeDirect, models.MsgTypeChannel, models.MsgTypePrivateChannel:
		send(api, message, bot)
	default:
		log.Warn().Msg("received unknown message type - no message to send")
	}
}

var interactionsRouter *mux.Router

// InteractiveComponents implementation to satisfy remote interface
// It will serve as a way for your bot to handle advance messaging, such as message attachments.
// When your bot is up and running, it will have an http/https endpoint to handle rules for sending attachments.
func (c *Client) InteractiveComponents(inputMsgs chan<- models.Message, message *models.Message, rule models.Rule, bot *models.Bot) {
	if bot.InteractiveComponents && c.SigningSecret != "" {
		if bot.SlackInteractionsCallbackPath == "" {
			log.Error().Msg("need to specify a callback path for the 'slack_interactions_callback_path' field in the bot.yml (e.g. \"/slack_events/v1/mybot_dev-v1_interactions\")")
			log.Warn().Msg("closing interactions reader (will not be able to read interactive components)")

			return
		}

		if interactionsRouter == nil {
			// create router for the Interactive Components server
			interactionsRouter = mux.NewRouter()

			// interaction health check handler
			interactionsRouter.HandleFunc("/interaction_health", getInteractiveComponentHealthHandler(bot)).Methods("GET")

			// Rule handler and endpoint
			ruleHandle := getInteractiveComponentRuleHandler(c.SigningSecret, inputMsgs, message, rule, bot)

			// We use regex for interactions routing for any bot using this framework
			// e.g. /slack_events/v1/mybot_dev-v1_interactions
			if !isValidPath(bot.SlackInteractionsCallbackPath) {
				log.Error().Msg(`invalid events path - please double check your path value/syntax (e.g. "/slack_events/v1/mybot_dev-v1_interactions")`)
				log.Warn().Msg("closing interaction components reader (will not be able to read interactive components)")

				return
			}

			interactionsRouter.HandleFunc(bot.SlackInteractionsCallbackPath, ruleHandle).Methods("POST")

			// start Interactive Components server
			go func() {
				err := http.ListenAndServe(":4000", interactionsRouter)
				if err != nil {
					log.Error().Msgf("unable to start interactions endpoint: %v", err)
				}
			}()

			log.Info().Msgf("slack interactive components server is listening to %#q", bot.SlackInteractionsCallbackPath)
		}

		// Process the hit rule for Interactive Components, e.g. interactive messages
		processInteractiveComponentRule(rule, message, bot)
	}
}
