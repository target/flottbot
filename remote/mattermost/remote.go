// SPDX-License-Identifier: Apache-2.0

package mattermost

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/rs/zerolog/log"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/remote"
)

// Client struct.
type Client struct {
	Server   string
	Token    string
	BotID    string
	Insecure bool
}

// validate that Client adheres to remote interface.
var _ remote.Remote = (*Client)(nil)

// instantiate a new mattermost client.
func (c *Client) new() *model.Client4 {
	log.Info().Msgf("%#v", c)

	url := "https://" + c.Server
	if c.Insecure {
		url = "http://" + c.Server
	}

	log.Info().Msgf("connecting to instance with url: %s", url)
	api := model.NewAPIv4Client(url)
	api.SetToken(c.Token)

	return api
}

func (c *Client) Name() string { return "mattermost" }

func (c *Client) Reaction(_ models.Message, rule models.Rule, _ *models.Bot) {
	if rule.RemoveReaction != "" {
		log.Debug().Msg("remove reaction not implemented for mattermost")
	}

	if rule.Reaction != "" {
		log.Debug().Msg("reactions not implemented for mattermost")
	}
}

func (c *Client) Read(inputMsgs chan<- models.Message, _ map[string]models.Rule, _ *models.Bot) {
	api := c.new()

	ctx := context.Background()

	user, resp, err := api.GetUser(ctx, "me", "")
	if err != nil {
		log.Fatal().Msgf("could not login, %s", err)
	}

	log.Info().Interface("user", user.Username).Interface("resp", resp).Msg("")
	log.Info().Msg("logged in to mattermost")

	c.BotID = user.Username

	url := "wss://" + c.Server
	if c.Insecure {
		url = "ws://" + c.Server
	}

	sock, err := model.NewWebSocketClient4(url, c.Token)
	if err != nil {
		log.Info().Msgf("%s", err)
		panic(1)
	}

	log.Info().Msg("mattermost websocket connected")

	sock.Listen()

	go func(ctx context.Context) {
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
					log.Info().Msg("bot mentioned in post")
				}

				user, _, err := api.GetUser(ctx, post.UserId, "")
				if err != nil {
					log.Fatal().Msgf("could not get username, %s", err)
				}

				channelName, _, err := api.GetChannel(ctx, post.ChannelId, "")
				if err != nil {
					log.Fatal().Msgf("could not get channelName, %s", err)
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
				log.Debug().Msgf("no action for %s event", event.EventType())
				continue
			}
		}
	}(ctx)
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
	ctx := context.Background()

	if user, resp, err := api.GetUser(ctx, "me", ""); err != nil {
		log.Fatal().Msgf("could not login, %s", err)
	} else {
		log.Info().Interface("user", user.Username).Interface("resp", resp).Msg("")
		log.Info().Msg("logged in to mattermost")

		c.BotID = user.Username
	}

	post := &model.Post{}
	post.ChannelId = message.ChannelID
	post.Message = message.Output

	if _, _, err := api.CreatePost(ctx, post); err != nil {
		log.Error().Err(err).Msg("failed to create post")
	}
}
