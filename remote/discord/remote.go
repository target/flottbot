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
// Note: Discord expects the actual unicode emoji, so you need to have that in the rule setup, ie.
// .
// reaction: 🔥
// .
func (c *Client) Reaction(message models.Message, rule models.Rule, bot *models.Bot) {
	if rule.RemoveReaction != "" {
		// Init api client
		dg := c.new()
		// Remove bot reaction from message
		if err := dg.MessageReactionRemove(message.ChannelID, message.ID, rule.RemoveReaction, "@me"); err != nil {
			bot.Log.Error().Msgf("could not add reaction '%s' - make sure to use actual emoji unicode characters", err)
			return
		}

		bot.Log.Info().Msgf("removed reaction '%s' for rule '%s'", rule.RemoveReaction, rule.Name)
	}

	if rule.Reaction != "" {
		// Init api client
		dg := c.new()
		// React with desired reaction
		if err := dg.MessageReactionAdd(message.ChannelID, message.ID, rule.Reaction); err != nil {
			bot.Log.Error().Msgf("could not add reaction: %v", err)
			return
		}

		bot.Log.Info().Msgf("added reaction '%s' for rule '%s'", rule.Reaction, rule.Name)
	}
}

// Read implementation to satisfy remote interface
func (c *Client) Read(inputMsgs chan<- models.Message, rules map[string]models.Rule, bot *models.Bot) {
	dg := c.new()
	if dg == nil {
		bot.Log.Error().Msg("failed to initialize discord client")
		return
	}

	err := dg.Open()
	if err != nil {
		bot.Log.Error().Msgf("failed to open connection to discord server - error: %v", err)
		return
	}
	// Wait here until CTRL-C or other term signal is received
	bot.Log.Info().Msgf("discord is now running '%s' - press ctrl-c to exit", bot.Name)

	// get informatiom about ourself
	botuser, err := dg.User("@me")
	if err != nil {
		bot.Log.Error().Msgf("failed to get bot name from discord - error: %v", err)
		return
	}

	bot.Name = botuser.Username
	bot.ID = botuser.ID

	foundGuild := false

	guilds := dg.State.Guilds
	for _, g := range guilds {
		if g.ID == bot.DiscordServerID {
			foundGuild = true
			break
		}
	}

	if !foundGuild {
		bot.Log.Error().Msg("unable to find server defined in 'discord_server_id' - has the bot been added to the server?")
		return
	}

	rooms := make(map[string]string)
	users := make(map[string]string)
	groups := make(map[string]string)

	// populate rooms
	gchans, err := dg.GuildChannels(bot.DiscordServerID)
	if err != nil {
		bot.Log.Error().Msgf("unable to get channels - error: %v", err)
	}

	for _, gchan := range gchans {
		rooms[gchan.Name] = gchan.ID
	}

	// populate users - 1000 is API limit
	// TODO: paginate to *really* get all - would have to find highest ID
	// from prev results and pass in as second param to .GuildMembers
	gmembers, err := dg.GuildMembers(bot.DiscordServerID, "", 1000)
	if err != nil {
		bot.Log.Error().Msg("unable to get users")
	}

	for _, gmember := range gmembers {
		users[gmember.User.Username] = gmember.User.ID
	}

	// populate user groups
	groles, err := dg.GuildRoles(bot.DiscordServerID)
	if err != nil {
		bot.Log.Error().Msg("unable to get roles")
	}

	for _, grole := range groles {
		groups[grole.Name] = grole.ID
	}

	bot.Rooms = rooms
	bot.Users = users
	bot.UserGroups = groups

	// Register a callback for MessageCreate events
	dg.AddHandler(handleDiscordMessage(bot, inputMsgs))
}

// Send implementation to satisfy remote interface
func (c *Client) Send(message models.Message, bot *models.Bot) {
	dg := c.new()

	// Timestamp message
	message.EndTime = models.MessageTimestamp()

	switch message.Type {
	case models.MsgTypeDirect, models.MsgTypeChannel:
		send(dg, message, bot)
	default:
		bot.Log.Error().Msgf("unable to send message of type: %d", message.Type)
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
		// check if we should respond to bot messages
		if m.Author.Bot && !bot.RespondToBots {
			return
		}

		// ignore messages from self
		if m.Author.ID == bot.ID {
			return
		}

		// Process message
		message := models.NewMessage()
		switch m.Type {
		case discordgo.MessageTypeDefault:
			var msgType models.MessageType

			ch, err := s.Channel(m.ChannelID)
			if err != nil {
				bot.Log.Error().Msg("discord: failed to retrieve channel")
			}

			t, err := m.Timestamp.Parse()
			if err != nil {
				bot.Log.Error().Msgf("discord remote: failed to parse message timestamp")
			}

			timestamp := strconv.FormatInt(t.Unix(), 10)

			switch ch.Type {
			case discordgo.ChannelTypeDM:
				msgType = models.MsgTypeDirect
			case discordgo.ChannelTypeGuildText:
				msgType = models.MsgTypeChannel
			default:
				msgType = models.MsgTypeChannel
				bot.Log.Warn().Msgf("discord: read message from unsupported channel type '%d' - defaulting to use channel type 0 ('GUILD_TEXT')", ch.Type)
			}

			contents, mentioned := removeBotMention(m.Content, s.State.User.ID)
			message = populateMessage(message, msgType, m.ChannelID, m.Message.ID, contents, timestamp, mentioned, m.Author, bot)
		default:
			bot.Log.Error().Msgf("discord: read message of unsupported type '%d' - unable to populate message attributes", m.Type)
		}
		inputMsgs <- message
	}
}
