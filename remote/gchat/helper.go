// SPDX-License-Identifier: Apache-2.0

package gchat

import (
	"encoding/json"
	"fmt"
	"strings"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/chat/v1"

	"github.com/target/flottbot/models"
)

type DomainEvent struct {
	User struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		AvatarURL   string `json:"avatarUrl"`
		Email       string `json:"email"`
		Type        string `json:"type"`
		DomainID    string `json:"domainId"`
	} `json:"user"`
}

// HandleOutput handles input messages for this remote.
func HandleRemoteInput(inputMsgs chan<- models.Message, rules map[string]models.Rule, bot *models.Bot) {
	c := &Client{
		Credentials:        bot.GoogleChatCredentials,
		ProjectID:          bot.GoogleChatProjectID,
		SubscriptionID:     bot.GoogleChatSubscriptionID,
		ForceReplyToThread: bot.GoogleChatForceReplyToThread,
	}

	// Read messages from Google Chat
	go c.Read(inputMsgs, rules, bot)
}

// HandleRemoteOutput handles output messages for this remote.
func HandleRemoteOutput(message models.Message, bot *models.Bot) {
	c := &Client{
		Credentials:        bot.GoogleChatCredentials,
		ProjectID:          bot.GoogleChatProjectID,
		SubscriptionID:     bot.GoogleChatSubscriptionID,
		ForceReplyToThread: bot.GoogleChatForceReplyToThread,
	}

	// Send messages to Google Chat
	go c.Send(message, bot)
}

// toMessage converts a PubSub message to Flottbot Message.
func toMessage(m *pubsub.Message) (models.Message, error) {
	message := models.NewMessage()

	var event chat.DeprecatedEvent

	err := json.Unmarshal(m.Data, &event)
	if err != nil {
		return message, fmt.Errorf("google_chat was unable to parse event %s: %w", m.ID, err)
	}

	msgType, err := getMessageType(event)
	if err != nil {
		return message, err
	}

	message.Type = msgType
	message.Timestamp = event.EventTime

	if event.Type == "MESSAGE" {
		message.Input = strings.TrimPrefix(event.Message.ArgumentText, " ")
		message.ID = event.Message.Name
		message.Service = models.MsgServiceChat
		message.ChannelName = event.Space.DisplayName
		message.ChannelID = event.Space.Name
		message.BotMentioned = true // Google Chat only supports @bot mentions
		message.DirectMessageOnly = event.Space.SingleUserBotDm
		message.ThreadID = event.Message.Thread.Name
		message.ThreadTimestamp = event.EventTime

		// make channel variables available
		message.Vars["_channel.name"] = message.ChannelName // will be empty if it came via DM
		message.Vars["_channel.id"] = message.ChannelID
		message.Vars["_thread.id"] = message.ThreadID

		// make timestamp information available
		message.Vars["_source.timestamp"] = event.EventTime
	}

	if event.User != nil {
		message.Vars["_user.name"] = event.User.DisplayName
		message.Vars["_user.id"] = event.User.Name
		message.Vars["_user.displayname"] = event.User.DisplayName

		// Try parsing as a domain message to get user email
		var domainEvent DomainEvent
		if err := json.Unmarshal(m.Data, &domainEvent); err == nil {
			message.Vars["_user.email"] = domainEvent.User.Email
		}
	}

	return message, nil
}
