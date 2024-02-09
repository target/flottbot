// SPDX-License-Identifier: Apache-2.0

package slack

import (
	"time"

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
func (c *Client) Reaction(message models.Message, rule models.Rule, _ *models.Bot) {
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
func (c *Client) Read(inputMsgs chan<- models.Message, _ map[string]models.Rule, bot *models.Bot) {
	// init api client
	api := c.new()

	// get bot id
	rat, err := api.AuthTest()
	if err != nil {
		log.Error().Msgf("the 'slack_token' is invalid or is unauthorized: %s", err)

		return
	}

	// fetch rooms async
	go func(b *models.Bot) {
		start := time.Now()

		// get bot rooms
		b.Rooms = getRooms(api)

		log.Info().Msgf("fetched %d rooms in %s", len(b.Rooms), time.Since(start).String())
	}(bot)

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
func (c *Client) Send(message models.Message, _ *models.Bot) {
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
		send(api, message)
	default:
		log.Warn().Msg("received unknown message type - no message to send")
	}
}
