package core

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/utils"
)

var defaultSlackListenerPort = "3000"

// Configure searches the config directory for the bot.yml to create a Bot object.
// The Bot object will be passed around to make accessible system-specific information.
func Configure(bot *models.Bot) {
	log.Info().Msg("Configuring bot...")

	initLogger(bot)

	validateRemoteSetup(bot)

	configureChatApplication(bot)

	bot.Log.Info().Msgf("Configured bot '%s'!", bot.Name)
}

// initLogger sets log configuration for the bot
func initLogger(b *models.Bot) {
	b.Log = zerolog.New(os.Stdout).Level(zerolog.ErrorLevel)

	if b.Debug {
		b.Log.Level(zerolog.DebugLevel)
	}
}

// configureChatApplication configures a user's specified chat application
// TODO: Refactor to keep remote specifics in remote/
func configureChatApplication(bot *models.Bot) {

	// update the bot name
	token, err := utils.Substitute(bot.Name, map[string]string{})
	if err != nil {
		bot.Log.Warn().Msgf("Could not configure bot Name: %s", err)
	}

	bot.Name = token

	if bot.ChatApplication != "" {
		switch strings.ToLower(bot.ChatApplication) {
		case "discord":
			// Discord bot token
			token, err := utils.Substitute(bot.DiscordToken, map[string]string{})
			if err != nil {
				bot.Log.Warn().Msgf("Could not set Discord Token: %s", err)
				bot.RunChat = false
			}

			if token == "" {
				bot.Log.Warn().Msgf("Discord Token is empty: '%s'", token)
				bot.RunChat = false
			}

			bot.DiscordToken = token

			// Discord Server ID
			// See https://support.discordapp.com/hc/en-us/articles/206346498-Where-can-I-find-my-User-Server-Message-ID-
			serverID, err := utils.Substitute(bot.DiscordServerID, map[string]string{})
			if err != nil {
				bot.Log.Warn().Msgf("Could not set Discord Server ID: %s", err)
				bot.RunChat = false
			}

			if serverID == "" {
				bot.Log.Warn().Msgf("Discord Server ID is empty: '%s'", serverID)
				bot.RunChat = false
			}

			bot.DiscordServerID = serverID

		case "slack":
			configureSlackBot(bot)

		case "telegram":
			token, err := utils.Substitute(bot.TelegramToken, map[string]string{})
			if err != nil {
				bot.Log.Warn().Msgf("Could not set telegram Token: %s", err)
				bot.RunChat = false
			}

			if token == "" {
				bot.Log.Warn().Msgf("telegram Token is empty: '%s'", token)
				bot.RunChat = false
			}

			bot.TelegramToken = token

		default:
			bot.Log.Error().Msgf("Chat application '%s' is not supported", bot.ChatApplication)
			bot.RunChat = false
		}
	}
}

func configureSlackBot(bot *models.Bot) {
	// Slack bot token
	token, err := utils.Substitute(bot.SlackToken, map[string]string{})
	if err != nil {
		bot.Log.Warn().Msgf("Could not set Slack Token: %s", err)
		bot.RunChat = false
	}

	if token == "" {
		bot.Log.Warn().Msgf("Slack Token is empty: %s", token)
		bot.RunChat = false
	}

	bot.SlackToken = token

	// Slack signing secret
	signingSecret, err := utils.Substitute(bot.SlackSigningSecret, map[string]string{})
	if err != nil {
		bot.Log.Warn().Msgf("Could not set Slack Signing Secret: %s", err)
		bot.Log.Warn().Msg("Defaulting to use Slack RTM")

		signingSecret = ""
	}

	bot.SlackSigningSecret = signingSecret

	// Get Slack Events path
	eCallbackPath, err := utils.Substitute(bot.SlackEventsCallbackPath, map[string]string{})
	if err != nil {
		bot.Log.Error().Msgf("Could not set Slack Events API callback path: %s", err)
		bot.Log.Warn().Msg("Defaulting to use Slack RTM")
		bot.SlackSigningSecret = ""
	}

	bot.SlackEventsCallbackPath = eCallbackPath

	// Get Slack Interactive Components path
	iCallbackPath, err := utils.Substitute(bot.SlackInteractionsCallbackPath, map[string]string{})
	if err != nil {
		bot.Log.Error().Msgf("Could not set Slack Interactive Components callback path: %s", err)
		bot.InteractiveComponents = false
	}

	if iCallbackPath == "" {
		bot.Log.Warn().Msgf("Slack Interactive Components callback path is empty: %s", iCallbackPath)
		bot.InteractiveComponents = false
	}

	bot.SlackInteractionsCallbackPath = iCallbackPath

	// Get Slack HTTP listener port
	lPort, err := utils.Substitute(bot.SlackListenerPort, map[string]string{})
	if err != nil {
		bot.Log.Error().Msgf("Could not set Slack listener port: %s", err)
		bot.SlackListenerPort = ""
	}

	// set slack http listener port from config file or default
	lPortEnvWasUnset := strings.Contains(lPort, "${") // e.g. slack_listener_port: ${PORT}
	if lPort == "" || lPortEnvWasUnset {
		bot.Log.Warn().Msgf("Slack listener port is empty: %s", lPort)
		bot.Log.Info().Str("defaultSlackListenerPort", defaultSlackListenerPort).Msg("Using default slack listener port.")
		lPort = defaultSlackListenerPort
	}

	bot.SlackListenerPort = lPort
}

func validateRemoteSetup(bot *models.Bot) {
	if bot.ChatApplication != "" {
		bot.RunChat = true
	}

	if bot.CLI {
		bot.RunCLI = true
	}

	if !bot.CLI && bot.ChatApplication == "" {
		bot.Log.Fatal().Msgf("No chat_application specified and cli mode is not enabled. Exiting...")
	}

	if bot.Scheduler {
		bot.RunScheduler = true
		if bot.CLI && bot.ChatApplication == "" {
			bot.Log.Warn().Msg("Scheduler does not support scheduled outputs to CLI mode")
			bot.RunScheduler = false
		}

		if bot.ChatApplication == "" {
			bot.Log.Warn().Msg("Scheduler did not find any configured chat applications. Scheduler is closing")
			bot.RunScheduler = false
		}
	}
}
