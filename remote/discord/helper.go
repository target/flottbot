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

	// if msgType != models.MsgTypeDirect {
	// 	name, ok := findKey(bot.Rooms, channel)
	// 	if !ok {
	// 		bot.Log.Warnf("Could not find name of channel '%s'.", channel)
	// 	}
	// 	message.ChannelName = name
	// }

	message.Vars["_channel.id"] = channel
	message.Vars["_channel.name"] = channel

	// Populate message user sender
	// These will be accessible on rules via ${_user.email}, etc
	if user != nil { // nil user implies a message from an api/bot (i.e. not an actual user)
		message.Vars["_user.email"] = user.Email
		// message.Vars["_user.firstname"] = ""
		// message.Vars["_user.lastname"] = ""
		message.Vars["_user.name"] = user.Username
		message.Vars["_user.id"] = user.ID
	}

	message.Debug = true
	return message
}
