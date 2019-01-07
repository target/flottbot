package discord

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
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

// creates a new Discord session
func (c *Client) new() *discordgo.Session {
	// Create a new Discord session using the provided bot token
	dg, err := discordgo.New("Bot " + c.Token)
	if err != nil {
		return nil
	}
	return dg
}

// Reaction implementation to satisfy remote interface
func (c *Client) Reaction(message models.Message, rule models.Rule, bot *models.Bot) {
	if rule.RemoveReaction != "" {
		// Init api client
		dg := c.new()
		// Remove bot reaction from message
		if err := dg.MessageReactionRemove(message.ChannelID, message.ID, rule.RemoveReaction, "@me"); err != nil {
			bot.Log.Errorf("Could not add reaction '%s'", err)
			return
		}
		bot.Log.Debugf("Removed reaction '%s' for rule %s", rule.RemoveReaction, rule.Name)
	}
	if rule.Reaction != "" {
		// Init api client
		dg := c.new()
		// React with desired reaction
		if err := dg.MessageReactionAdd(message.ChannelID, message.ID, rule.Reaction); err != nil {
			bot.Log.Errorf("Could not add reaction '%s'", err)
			return
		}
		bot.Log.Debugf("Added reaction '%s' for rule %s", rule.Reaction, rule.Name)
	}
}

// Read implementation to satisfy remote interface
func (c *Client) Read(inputMsgs chan<- models.Message, rules map[string]models.Rule, bot *models.Bot) {
	dg := c.new()
	if dg == nil {
		bot.Log.Error("Failed to initialize Discord client")
		return
	}
	err := dg.Open()
	if err != nil {
		bot.Log.Errorf("Failed to open connection to Discord server. Error: %s", err.Error())
		return
	}
	// Wait here until CTRL-C or other term signal is received
	bot.Log.Infof("Discord is now running '%s'. Press CTRL-C to exit", bot.Name)

	// get informatiom about ourself
	botuser, err := dg.User("@me")
	if err != nil {
		bot.Log.Errorf("Failed to get bot name from Discord. Error: %s", err.Error())
		return
	}
	bot.Name = botuser.Username

	// Register a callback for MessageCreate events
	dg.AddHandler(handleDiscordMessage(bot, inputMsgs))
}

// Send implementation to satisfy remote interface
func (c *Client) Send(message models.Message, bot *models.Bot) {
	dg := c.new()
	switch message.Type {
	case models.MsgTypeDirect, models.MsgTypeChannel:
		dg.ChannelMessageSend(message.ChannelID, message.Output)
	default:
		bot.Log.Errorf("Unable to send message of type %d", message.Type)
	}
}

// InteractiveComponents implementation to satisfy remote interface
func (c *Client) InteractiveComponents(inputMsgs chan<- models.Message, message *models.Message, rule models.Rule, bot *models.Bot) {
	// not implemented for Discord
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to
func handleDiscordMessage(bot *models.Bot, inputMsgs chan<- models.Message) interface{} {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		// This isn't required in this specific example but it's a good practice
		if m.Author.Bot {
			return
		}
		// Ignore messages in public channels that don't mention the bot
		ch, _ := s.Channel(m.ChannelID)
		if ch.Type == discordgo.ChannelTypeGuildText {
			botmention := false
			for _, mention := range m.Mentions {
				if mention.Username == bot.Name {
					botmention = true
				}
			}
			if !botmention {
				return
			}
		}
		// Process message
		message := models.NewMessage()
		switch m.Type {
		case discordgo.MessageTypeDefault:
			t, err := m.Timestamp.Parse()
			if err != nil {
				bot.Log.Errorf("Discord Remote: Failed to parse message timestamp.")
			}
			timestamp := strconv.FormatInt(t.Unix(), 10)
			msgType := models.MsgTypeChannel
			switch ch.Type {
			case discordgo.ChannelTypeDM:
				msgType = models.MsgTypeDirect
			case discordgo.ChannelTypeGuildText:
				break
			default:
				bot.Log.Debugf("Discord Remote: read message from unsupported channel type '%d'. Defaulting to use channel type 0 ('GUILD_TEXT')", ch.Type)
			}
			contents, mentioned := removeBotMention(m.Content, s.State.User.ID)
			message = populateMessage(message, msgType, m.ChannelID, m.Message.ID, contents, timestamp, mentioned, m.Author, bot)
		default:
			bot.Log.Errorf("Discord Remote: read message of unsupported type '%d'. Unable to populate message attributes", m.Type)
		}
		inputMsgs <- message
	}
}
