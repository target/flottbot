// Copyright (c) 2023 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package gchat

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/chat/v1"
	"google.golang.org/api/option"

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
	Credentials    string
	ProjectID      string
	SubscriptionID string
}

// validate that Client adheres to remote interface.
var _ remote.Remote = (*Client)(nil)

// instantiate a new client.
func (c *Client) new() *pubsub.Client {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, c.ProjectID, option.WithCredentialsFile(c.Credentials))
	if err != nil {
		log.Error().Msgf("google_chat unable to authenticate: %s", err.Error())
	}

	return client
}

// Name returns the name of the remote.
func (c *Client) Name() string {
	return "google_chat"
}

// Read messages from Google Chat.
func (c *Client) Read(inputMsgs chan<- models.Message, _ map[string]models.Rule, _ *models.Bot) {
	ctx := context.Background()

	// init client
	client := c.new()

	sub := client.Subscription(c.SubscriptionID)

	err := sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
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
		ctx, option.WithCredentialsFile(c.Credentials),
		option.WithScopes("https://www.googleapis.com/auth/chat.bot"),
	)
	if err != nil {
		log.Fatal().Msgf("google_chat unable to create chat service: %s", err.Error())
	}

	msgService := chat.NewSpacesMessagesService(service)

	// Best effort. If the instance goes away, so be it.
	msg := &chat.Message{
		Text: message.Output,
	}

	if message.ThreadID != "" {
		msg.Thread = &chat.Thread{
			Name: message.ThreadID,
		}
	}

	_, err = msgService.Create(message.ChannelID, msg).Do()
	if err != nil {
		log.Error().Msgf("google_chat failed to create message: %s", err.Error())
	}
}

// Reaction implementation to satisfy remote interface.
func (c *Client) Reaction(_ models.Message, _ models.Rule, _ *models.Bot) {
	// Not implemented for Google Chat
}
