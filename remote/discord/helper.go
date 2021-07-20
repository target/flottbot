package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/target/flottbot/models"
)

/*
=================================================================
Discord helper functions (anything that uses the discord package)
=================================================================
*/

// populateMessage - populates the 'Message' object to be passed on for processing/sending
func populateMessage(message models.Message, msgType models.MessageType, channel, id string, text, timeStamp string, mentioned bool, user *discordgo.User, bot *models.Bot) models.Message {
	// Populate message attributes
	message.Type = msgType
	message.Service = models.MsgServiceChat
	message.ChannelID = channel
	message.Input = text
	message.Output = ""
	message.Timestamp = timeStamp
	message.BotMentioned = mentioned
	message.ID = id

	if msgType != models.MsgTypeDirect {
		name, ok := findKey(bot.Rooms, channel)
		if !ok {
			bot.Log.Error().Msgf("could not find name of channel '%s'", channel)
		}

		message.ChannelName = name
	}

	message.Vars["_channel.id"] = channel
	message.Vars["_channel.name"] = message.ChannelName

	message.Vars["_source.timestamp"] = timeStamp

	// Populate message user sender
	// These will be accessible on rules via ${_user.email}, etc
	if user != nil { // nil user implies a message from an api/bot (i.e. not an actual user)
		message.Vars["_user.email"] = user.Email
		message.Vars["_user.name"] = user.Username
		message.Vars["_user.id"] = user.ID
	}

	message.Debug = true

	return message
}

// send - handles the sending logic of a message going to Discord
func send(dg *discordgo.Session, message models.Message, bot *models.Bot) {
	if message.DirectMessageOnly {
		err := handleDirectMessage(dg, message, bot)
		if err != nil {
			bot.Log.Error().Msgf("problem sending message: %v", err)
		}
	} else {
		err := handleNonDirectMessage(dg, message, bot)
		if err != nil {
			bot.Log.Error().Msgf("problem sending message: %v", err)
		}
	}
}

// handleDirectMessage - handle sending logic for direct messages
func handleDirectMessage(dg *discordgo.Session, message models.Message, bot *models.Bot) error {
	// Is output to rooms set?
	if len(message.OutputToRooms) > 0 {
		bot.Log.Warn().Msg("you have specified 'direct_message_only' as 'true' and provided 'output_to_rooms' -" +
			" messages will not be sent to listed rooms - if you want to send messages to these rooms," +
			" please set 'direct_message_only' to 'false'")
	}
	// Is output to users set?
	if len(message.OutputToUsers) > 0 {
		bot.Log.Warn().Msg("you have specified 'direct_message_only' as 'true' and provided 'output_to_users' -" +
			" messages will not be sent to the listed users (other than you) - if you want to send messages to other users," +
			" please set 'direct_message_only' to 'false'.")
	}

	userChannel, err := dg.UserChannelCreate(message.Vars["_user.id"])
	if err != nil {
		return err
	}

	_, err = dg.ChannelMessageSend(userChannel.ID, message.Output)
	if err != nil {
		return err
	}

	return nil
}

// handleNonDirectMessage - handle sending logic for non direct messages
func handleNonDirectMessage(dg *discordgo.Session, message models.Message, bot *models.Bot) error {
	if len(message.OutputToUsers) == 0 && len(message.OutputToRooms) == 0 && len(message.Output) > 0 {
		_, err := dg.ChannelMessageSend(message.ChannelID, message.Output)
		if err != nil {
			return err
		}
	}

	// message.OutputToRooms is already processed to be the translated IDs
	// vs. the originally provided names
	if len(message.OutputToRooms) > 0 {
		for _, roomID := range message.OutputToRooms {
			_, err := dg.ChannelMessageSend(roomID, message.Output)
			if err != nil {
				return err
			}
		}
	}

	if len(message.OutputToUsers) > 0 {
		for _, user := range message.OutputToUsers {
			userChannel, err := dg.UserChannelCreate(bot.Users[user])
			if err != nil {
				return err
			}

			_, err = dg.ChannelMessageSend(userChannel.ID, message.Output)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
