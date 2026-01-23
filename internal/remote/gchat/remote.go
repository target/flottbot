// SPDX-License-Identifier: Apache-2.0

package gchat

import (
	"context"

	"cloud.google.com/go/pubsub/v2"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/chat/v1"
	"google.golang.org/api/option"

	"github.com/target/flottbot/internal/models"
	"github.com/target/flottbot/internal/remote"
)

/*
=======================================
Implementation for the Remote interface
=======================================
*/

// Client struct.
type Client struct {
	Credentials        string
	ProjectID          string
	SubscriptionID     string
	ForceReplyToThread bool
}

// validate that Client adheres to remote interface.
var _ remote.Remote = (*Client)(nil)

// instantiate a new client.
func (c *Client) new() *pubsub.Client {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, c.ProjectID, option.WithAuthCredentialsFile(option.ServiceAccount, c.Credentials))
	if err != nil {
		log.Error().Msgf("google_chat unable to authenticate: %s", err.Error())
	}

	return client
}

// Name returns the name of the remote.
func (c *Client) Name() string {
	return models.ChatAppGoogleChat
}

// Read messages from Google Chat.
func (c *Client) Read(inputMsgs chan<- models.Message, _ map[string]models.Rule, _ *models.Bot) {
	ctx := context.Background()

	// init client
	client := c.new()

	sub := client.Subscriber(c.SubscriptionID)

	err := sub.Receive(ctx, func(_ context.Context, m *pubsub.Message) {
		defer m.Ack()

		// Convert Google Chat Message to Flottbot Message
		message, err := toMessage(m)
		if err != nil {
			log.Error().Msg(err.Error())
			return
		}

		// send to flotbot core for processing
		inputMsgs <- message
	})
	if err != nil {
		log.Fatal().Msgf("google_chat unable to create subscription against %s: %s", c.SubscriptionID, err.Error())
	}

	log.Info().Msgf("google_chat successfully subscribed to %s", c.SubscriptionID)
}

// Send messages to Google Chat.
func (c *Client) Send(message models.Message, _ *models.Bot) {
	ctx := context.Background()

	service, err := chat.NewService(
		ctx, option.WithAuthCredentialsFile(option.ServiceAccount, c.Credentials),
		option.WithScopes("https://www.googleapis.com/auth/chat.bot"),
	)
	if err != nil {
		log.Fatal().Msgf("google_chat unable to create chat service: %s", err.Error())
	}

	msgService := chat.NewSpacesMessagesService(service)

	// Best effort. If the instance goes away, so be it.
	msg := &chat.Message{
		Text: message.Output,
		Thread: &chat.Thread{
			Name: message.ThreadID,
		},
	}

	request := msgService.Create(message.ChannelID, msg)
	if c.ForceReplyToThread {
		request = request.MessageReplyOption("REPLY_MESSAGE_FALLBACK_TO_NEW_THREAD")
	}

	_, err = request.Do()
	if err != nil {
		log.Error().Msgf("google_chat failed to create message: %s", err.Error())
	}
}

// Reaction implementation to satisfy remote interface.
func (c *Client) Reaction(_ models.Message, _ models.Rule, _ *models.Bot) {
	// Not implemented for Google Chat
}
