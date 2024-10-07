// SPDX-License-Identifier: Apache-2.0

package mattermost

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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

func (c *Client) Read(inputMsgs chan<- models.Message, _ map[string]models.Rule, bot *models.Bot) {
	api := c.new()

	ctx := context.Background()

	user, resp, err := api.GetUser(ctx, "me", "")
	if err != nil {
		log.Fatal().Msgf("could not login, %s", err)
	}

	log.Info().Interface("user", user.Username).Interface("resp", resp).Msg("")
	log.Info().Msg("logged in to mattermost")

	c.BotID = user.Username

	go func(b *models.Bot) {
		rooms := make(map[string]string)

		teams, _, err := api.GetTeamsForUser(ctx, user.Id, "")
		if err != nil {
			log.Fatal().Err(err)
		}

		for _, team := range teams {
			r, _, err := api.GetChannelsForTeamForUser(ctx, team.Id, user.Id, false, "")
			if err != nil {
				log.Fatal().Err(err)
			}
			for _, i := range r {
				teamRoom := fmt.Sprintf("%s/%s", team.Name, i.Name)
				rooms[teamRoom] = i.Id
			}
		}
		bot.Rooms = rooms
	}(bot)

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

func (c *Client) Send(message models.Message, bot *models.Bot) {
	api := c.new()
	ctx := context.Background()

	user, _, err := api.GetUser(ctx, "me", "")
	if err != nil {
		log.Fatal().Msgf("could not login, %s", err)
	}
	log.Info().Msg("logged in to mattermost")

	c.BotID = user.Username

	post := &model.Post{}
	post.Message = message.Output

	if message.DirectMessageOnly {
		target := message.Vars["_user.id"]
		log.Info().Msgf("Creating direct message between %s, and %s",
			target, c.BotID)
		directChannel, _, err := api.CreateDirectChannel(ctx,
			user.Id, target)
		if err != nil {
			log.Error().Msgf("%v", err)
			return
		}

		post.ChannelId = directChannel.Id
		if _, resp, err := api.CreatePost(ctx, post); err != nil {
			log.Error().Err(err)
		} else {
			log.Debug().Interface("responce", resp).Msg("")
		}

		return
	}

	if len(message.OutputToRooms) > 0 {
		for _, roomID := range message.OutputToRooms {
			post.ChannelId = roomID
			log.Debug().Msgf("Posting message: %v, to room %v",
				post.Message, post.ChannelId)
			if _, _, err := api.CreatePost(ctx, post); err != nil {
				log.Error().Err(err)
			}

		}
	}

	if len(message.OutputToUsers) > 0 {
		for _, u := range message.OutputToUsers {
			log.Info().Msgf("Getting user id for %s", u)
			target, err := getUserID(api, u)
			log.Info().Msgf("Creating direct message between %s, and %s",
				target, c.BotID)
			directChannel, _, err := api.CreateDirectChannel(ctx,
				user.Id, target)
			if err != nil {
				log.Error().Err(err)
				break
			}

			post.ChannelId = directChannel.Id
			if _, resp, err := api.CreatePost(ctx, post); err != nil {
				log.Error().Err(err)
			} else {
				log.Debug().Interface("responce", resp).Msg("")
			}

		}
	}

	if len(message.OutputToRooms) == 0 && len(message.OutputToUsers) == 0 {
		post := &model.Post{}
		post.ChannelId = message.ChannelID
		post.Message = message.Output

		if _, _, err := api.CreatePost(ctx, post); err != nil {
			log.Error().Err(err).Msg("failed to create post")
		}
	}

}

func getUserID(api *model.Client4, username string) (string, error) {
	log.Info().Msgf("Getting user id for %s", username)

	ctx := context.Background()

	// trim any leading '@' from the provided username
	username = strings.TrimPrefix(username, "@")

	user, _, err := api.GetUserByUsername(ctx, username, "")
	if err != nil {
		return "", nil

	}
	return user.Id, nil
}

func (c Client) sendDirectMessage(api *model.Client4, message models.Message) error {
	userID := message.Vars["_user.id"]
	ctx := context.Background()
	log.Info().Msgf("Creating direct message between %s, and %s",
		userID, c.BotID)
	directChannel, _, err := api.CreateDirectChannel(ctx,
		userID, c.BotID)
	if err != nil {
		log.Error().Msgf("%v", err)
		return err
	}

	post := &model.Post{}
	post.Message = message.Output
	post.ChannelId = directChannel.Id
	return nil
}

func (c Client) sendMessage(api *model.Client4, ctx context.Context, post *model.Post) error {
	if _, resp, err := api.CreatePost(ctx, post); err != nil {
		log.Error().Err(err)
	} else {
		log.Debug().Interface("responce", resp).Msg("")
	}
	return nil
}
