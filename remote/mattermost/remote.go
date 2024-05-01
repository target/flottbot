// SPDX-License-Identifier: Apache-2.0

package mattermost

import (
	"encoding/json"
	"strconv"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/remote"
)

// Client struct.
type Client struct {
	Server string
	Token  string
	BotID  string
}

type mmLogger struct {
	zerolog.Logger
}

func (l *mmLogger) Output(_ int, s string) error {
	l.Logger.Info().Msg(s)
	return nil
}

// validate that Client adheres to remote interface.
var _ remote.Remote = (*Client)(nil)

// instantiate a new mattermost client.
func (c *Client) new() *model.Client4 {
	log.Info().Msgf("%#v", c)
	api := model.NewAPIv4Client("http://" + c.Server)
	api.SetToken(c.Token)

	return api
}

func (c *Client) Name() string { return "mattermost" }

func (c *Client) Reaction(_ models.Message, rule models.Rule, _ *models.Bot) {
	if rule.RemoveReaction != "" {
		log.Debug().Msg("Remove reaction not implemented for mattermost")
	}

	if rule.Reaction != "" {
		log.Debug().Msg("Reactions not implemented for mattermost")
	}

}

func (c *Client) Read(inputMsgs chan<- models.Message, _ map[string]models.Rule, bot *models.Bot) {
	api := c.new()
	if user, resp, err := api.GetUser("me", ""); err != nil {
		log.Fatal().Msgf("Could not login, %s", err)
	} else {
		log.Info().Interface("user", user.Username).Interface("resp", resp).Msg("")
		log.Info().Msg("Logged in to mattermost")

		c.BotID = user.Username
	}

	sock, err := model.NewWebSocketClient4("ws://"+c.Server, c.Token)
	if err != nil {
		log.Info().Msgf("%s", err)
		panic(1)
	}

	log.Info().Msg("Mattermost Websocket connected")

	sock.Listen()

	go func() {
		for event := range sock.EventChannel {
			switch event.EventType() {
			case model.WebsocketEventHello:
				continue
			case model.WebsocketEventPosted:
				post := &model.Post{}

				err := json.Unmarshal([]byte(event.GetData()["post"].(string)), &post)
				if err != nil {
					log.Err(err)
				}

				log.Debug().Msgf("%+v\n", post)

				// remove the bot mention from the user input
				message, mentioned := removeBotMention(post.Message, c.BotID)
				if mentioned {
					log.Info().Msg("Bot mentioned in post")
				}

				user, _, err := api.GetUser(post.UserId, "")
				if err != nil {
					log.Fatal().Msgf("Could not get username, %s", err)
				}

				channelName, _, err := api.GetChannel(post.ChannelId, "")
				if err != nil {
					log.Fatal().Msgf("Could not get channelName, %s", err)
				}

				inputMsgs <- populateMessage(
					models.NewMessage(),
					models.MsgTypeChannel,
					post.ChannelId,
					channelName.Name,
					message,
					strconv.Itoa(int(post.CreateAt)),
					mentioned,
					post.UserId,
					user.Username,
				)

			default:
				log.Debug().Msgf("No Action for %s Event", event.EventType())
				continue
			}
		}
	}()
}

func populateMessage(
	message models.Message,
	messageType models.MessageType,
	channelID, channelName, text, timeStamp string,
	mentioned bool,
	userID string,
	userName string,
) models.Message {
	message.Type = messageType
	message.Service = models.MsgServiceChat
	message.ChannelID = channelID
	message.Input = text
	message.Output = ""
	message.Timestamp = timeStamp
	message.BotMentioned = mentioned

	// make channel variables available
	message.Vars["_channel.id"] = message.ChannelID
	message.Vars["_channel.name"] = channelName

	// make timestamp information available
	message.Vars["_source.timestamp"] = timeStamp

	message.Vars["_user.id"] = userID
	message.Vars["_user.name"] = userName

	return message
}

func (c *Client) Send(message models.Message, _ *models.Bot) {
	api := c.new()
	if user, resp, err := api.GetUser("me", ""); err != nil {
		log.Fatal().Msgf("Could not login, %s", err)
	} else {
		log.Info().Interface("user", user.Username).Interface("resp", resp).Msg("")
		log.Info().Msg("Logged in to mattermost")

		c.BotID = user.Username
	}

	post := &model.Post{}
	post.ChannelId = message.ChannelID
	post.Message = message.Output

	if _, _, err := api.CreatePost(post); err != nil {
		log.Error().Err(err).Msg("Failed to create post")
	}
}
