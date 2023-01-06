// Copyright (c) 2023 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package telegram

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"

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
	Token string
}

// validate that Client adheres to remote interface.
var _ remote.Remote = (*Client)(nil)

// creates a new telegram bot.
func (c *Client) new() *tgbotapi.BotAPI {
	telegramAPI, err := tgbotapi.NewBotAPI(c.Token)
	if err != nil {
		return nil
	}

	return telegramAPI
}

// Name returns the name of the remote.
func (c *Client) Name() string {
	return "telegram"
}

// Reaction implementation to satisfy remote interface.
func (c *Client) Reaction(message models.Message, rule models.Rule, bot *models.Bot) {
	// not implemented for Telegram
}

// Read implementation to satisfy remote interface.
func (c *Client) Read(inputMsgs chan<- models.Message, rules map[string]models.Rule, bot *models.Bot) {
	telegramAPI := c.new()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	botuser, err := telegramAPI.GetMe()
	if err != nil {
		log.Error().Msg("failed to initialize telegram client")
		return
	}

	bot.Name = botuser.UserName

	updates := telegramAPI.GetUpdatesChan(u)

	for update := range updates {
		var m *tgbotapi.Message

		if update.Message != nil {
			m = update.Message
		}

		if update.ChannelPost != nil {
			m = update.ChannelPost
		}

		if m == nil {
			continue
		}

		// check if we should respond to other bot messages
		if m.From != nil && m.From.IsBot && !bot.RespondToBots {
			continue
		}

		// only process messages not from our bot
		// Note: it looks like "From" is not being populated currently
		if m.From != nil && m.From.ID == botuser.ID {
			continue
		}

		msg, mentioned := processMessageText(m.Text, bot.Name)

		// support slash commands
		if len(m.Command()) > 0 {
			fullCmd := fmt.Sprintf("%s %s", m.Command(), m.CommandArguments())
			mentioned = true
			msg = strings.TrimSpace(fullCmd)
		}

		message := models.NewMessage()
		message.Timestamp = strconv.FormatInt(m.Time().Unix(), 10)
		message.Type = mapMessageType(*m)
		message.Input = msg
		message.Output = ""
		message.ID = strconv.Itoa(m.MessageID)
		message.Service = models.MsgServiceChat
		message.BotMentioned = mentioned
		message.ChannelID = strconv.FormatInt(m.Chat.ID, 10)

		// populate message with metadata
		if m.From != nil {
			message.Vars["_user.name"] = m.From.UserName
			message.Vars["_user.firstname"] = m.From.FirstName
			message.Vars["_user.lastname"] = m.From.LastName
			message.Vars["_user.id"] = strconv.FormatInt(m.From.ID, 10)
			message.Vars["_user.realnamenormalized"] = fmt.Sprintf("%s %s", m.From.FirstName, m.From.LastName)
			message.Vars["_user.displayname"] = m.From.UserName
			message.Vars["_user.displaynamenormalized"] = m.From.UserName
		}

		message.Vars["_channel.id"] = message.ChannelID
		message.Vars["_channel.name"] = m.Chat.Title

		message.Vars["_source.timestamp"] = strconv.Itoa(m.Date)

		inputMsgs <- message
	}
}

// Send implementation to satisfy remote interface.
func (c *Client) Send(message models.Message, bot *models.Bot) {
	telegramAPI := c.new()

	chatID, err := strconv.ParseInt(message.ChannelID, 10, 64)
	if err != nil {
		log.Error().Msgf("unable to retrieve chat id %#q", message.ChannelID)

		return
	}

	// handle directive to only send direct message to user
	// instead of sending back to originating channel
	if message.DirectMessageOnly {
		chatID, err = strconv.ParseInt(message.Vars["_user.id"], 10, 64)
		if err != nil {
			log.Error().Msgf("unable to retrieve user id %#q for direct message", message.Vars["_user.id"])

			return
		}
	}

	msg := tgbotapi.NewMessage(chatID, message.Output)

	_, err = telegramAPI.Send(msg)
	if err != nil {
		log.Error().Msgf("unable to send message: %v", err)
	}
}
