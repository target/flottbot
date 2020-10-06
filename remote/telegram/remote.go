package telegram

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/target/flottbot/models"
	"github.com/target/flottbot/remote"
)

/*
=======================================
Implementation for the Remote interface
=======================================
*/

// Client struct
type Client struct {
	Token string
}

// validate that Client adheres to remote interface
var _ remote.Remote = (*Client)(nil)

// creates a new telegram bot
func (c *Client) new() *tgbotapi.BotAPI {
	telegramAPI, err := tgbotapi.NewBotAPI(c.Token)
	if err != nil {
		return nil
	}
	return telegramAPI
}

// Reaction implementation to satisfy remote interface
func (c *Client) Reaction(message models.Message, rule models.Rule, bot *models.Bot) {
	// not implemented for Telegram
}

// Read implementation to satisfy remote interface
func (c *Client) Read(inputMsgs chan<- models.Message, rules map[string]models.Rule, bot *models.Bot) {
	telegramAPI := c.new()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	botuser, err := telegramAPI.GetMe()
	if err != nil {
		bot.Log.Error("Failed to initialize Telegram client")
		return
	}
	bot.Name = botuser.UserName

	updates, err := telegramAPI.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		message := models.NewMessage()
		message.Timestamp = strconv.FormatInt(update.Message.Time().Unix(), 10)
		message.Type = models.MsgTypeDirect
		message.Input = update.Message.Text
		message.Output = ""
		message.ID = strconv.Itoa(update.Message.MessageID)
		message.Service = models.MsgServiceChat
		message.ChannelID = strconv.FormatInt(update.Message.Chat.ID, 10)
		message.Vars["_user.name"] = update.Message.Chat.UserName

		inputMsgs <- message
	}
}

// Send implementation to satisfy remote interface
func (c *Client) Send(message models.Message, bot *models.Bot) {
	telegramAPI := c.new()
	chatID, err := strconv.ParseInt(message.ChannelID, 10, 64)
	if err != nil {
		bot.Log.Errorf("unable to retrive chat ID %s", message.ChannelID)
		return
	}

	msg := tgbotapi.NewMessage(chatID, message.Output)
	telegramAPI.Send(msg)
}

// InteractiveComponents implementation to satisfy remote interface
func (c *Client) InteractiveComponents(inputMsgs chan<- models.Message, message *models.Message, rule models.Rule, bot *models.Bot) {
	// not implemented for Telegram
}
