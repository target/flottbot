package core

import (
	"io"
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
	log.Info().Msg("configuring bot...")

	initLogger(bot)

	validateRemoteSetup(bot)

	configureChatApplication(bot)

	bot.Log.Info().Msgf("configured bot '%s'!", bot.Name)
}

// initLogger sets log configuration for the bot
func initLogger(b *models.Bot) {
	var out io.Writer
	var level zerolog.Level

	// defaults
	out = os.Stdout
	level = zerolog.InfoLevel

	// for CLI, use zerolog's ConsoleWriter
	if b.CLI {
		out = zerolog.ConsoleWriter{Out: os.Stderr}
	}

	if b.Debug {
		level = zerolog.DebugLevel
	}

	b.Log = zerolog.New(out).With().Timestamp().Logger().Level(level)
}

// configureChatApplication configures a user's specified chat application
// TODO: Refactor to keep remote specifics in remote/
func configureChatApplication(bot *models.Bot) {
	// emptyMap for substitute function
	// (it will only replace from env vars)
	emptyMap := map[string]string{}

	// update the bot name
	token, err := utils.Substitute(bot.Name, emptyMap)
	if err != nil {
		bot.Log.Warn().Msgf("could not configure bot 'name' field: %s", err.Error())
	}

	bot.Name = token

	if bot.ChatApplication != "" {
		switch strings.ToLower(bot.ChatApplication) {
		case "discord":
			// Discord bot token
			token, err := utils.Substitute(bot.DiscordToken, emptyMap)
			if err != nil {
				bot.Log.Error().Msgf("could not set 'discord_token': %s", err.Error())
				bot.RunChat = false
			}

			bot.DiscordToken = token

			// Discord Server ID
			// See https://support.discordapp.com/hc/en-us/articles/206346498-Where-can-I-find-my-User-Server-Message-ID-
			serverID, err := utils.Substitute(bot.DiscordServerID, emptyMap)
			if err != nil {
				bot.Log.Error().Msgf("could not set 'discord_server_id': %s", err.Error())
				bot.RunChat = false
			}

			bot.DiscordServerID = serverID

			if !isSet(token, serverID) {
				bot.Log.Error().Msg("bot is not configured correctly for discord - check that 'discord_token' and 'discord_server_id' are set")
				bot.RunChat = false
			}

		case "slack":
			configureSlackBot(bot)

		case "telegram":
			token, err := utils.Substitute(bot.TelegramToken, emptyMap)
			if err != nil {
				bot.Log.Error().Msgf("could not set 'telegram_token': %s", err.Error())
				bot.RunChat = false
			}

			if !isSet(token) {
				bot.Log.Error().Msg("bot is not configured correctly for telegram - check that 'telegram_token' is set")
				bot.RunChat = false
			}

			bot.TelegramToken = token

		default:
			bot.Log.Error().Msgf("chat application '%s' is not supported", bot.ChatApplication)
			bot.RunChat = false
		}
	}
}

func configureSlackBot(bot *models.Bot) {
	// emptyMap for substitute function
	// (it will only replace from env vars)
	emptyMap := map[string]string{}

	// slack_token
	token, err := utils.Substitute(bot.SlackToken, emptyMap)
	if err != nil {
		bot.Log.Error().Msgf("could not set 'slack_token': %s", err.Error())
	}

	bot.SlackToken = token

	// slack_app_token
	appToken, err := utils.Substitute(bot.SlackAppToken, emptyMap)
	if err != nil {
		bot.Log.Warn().Msgf("could not set 'slack_app_token': %s", err.Error())
	}

	bot.SlackAppToken = appToken

	// slack_signing_secret
	signingSecret, err := utils.Substitute(bot.SlackSigningSecret, emptyMap)
	if err != nil {
		bot.Log.Warn().Msgf("could not set 'slack_signing_secret': %s", err.Error())
	}

	bot.SlackSigningSecret = signingSecret

	// slack_events_callback_path
	eCallbackPath, err := utils.Substitute(bot.SlackEventsCallbackPath, emptyMap)
	if err != nil {
		bot.Log.Warn().Msgf("could not set 'slack_events_callback_path': %s", err.Error())
	}

	bot.SlackEventsCallbackPath = eCallbackPath

	// slack_interactions_callback_path
	iCallbackPath, err := utils.Substitute(bot.SlackInteractionsCallbackPath, emptyMap)
	if err != nil {
		bot.Log.Warn().Msgf("could not set 'slack_interactions_callback_path': %s", err.Error())
	}

	bot.SlackInteractionsCallbackPath = iCallbackPath

	// slack_listener_port
	lPort, err := utils.Substitute(bot.SlackListenerPort, emptyMap)
	if err != nil {
		bot.Log.Warn().Msgf("could not set 'slack_listener_port': %s", err.Error())
	}

	// set slack http listener port from config file or default
	if !isSet(lPort) {
		bot.Log.Warn().Msgf("'slack_listener_port' not set: %s", lPort)
		bot.Log.Info().Str("defaultSlackListenerPort", defaultSlackListenerPort).Msg("using default slack listener port.")
		lPort = defaultSlackListenerPort
	}

	bot.SlackListenerPort = lPort

	// check for valid setup
	// needs one of the following to be valid
	// 1. SLACK_TOKEN + SLACK_APP_TOKEN (socket mode)
	// 2. SLACK_TOKEN + SLACK_SIGNING_SECRET + SLACK_EVENTS_CALLBACK_PATH (events api)
	isSocketMode := isSet(token, appToken)
	isEventsAPI := isSet(token, signingSecret, eCallbackPath)
	if !isSocketMode && !isEventsAPI {
		bot.Log.Fatal().Msg("must have either 'slack_token', 'slack_app_token' or 'slack_token', 'slack_signing_secret', and 'slack_events_callback_path' set")
		bot.RunChat = false
	}
}

func validateRemoteSetup(bot *models.Bot) {
	if bot.ChatApplication != "" {
		bot.RunChat = true
	}

	if bot.CLI {
		bot.RunCLI = true
	}

	if !bot.CLI && bot.ChatApplication == "" {
		bot.Log.Fatal().Msgf("no i'chat_application specified' and cli mode is not enabled. exiting...")
	}

	if bot.Scheduler {
		bot.RunScheduler = true
		if bot.CLI && bot.ChatApplication == "" {
			bot.Log.Warn().Msg("scheduler does not support scheduled outputs to cli mode")
			bot.RunScheduler = false
		}

		if bot.ChatApplication == "" {
			bot.Log.Warn().Msg("scheduler did not find any configured chat applications - scheduler is closing")
			bot.RunScheduler = false
		}
	}
}

// isSet is a helper function to check whether any of the supplied
// strings are empty or unsubstituted (ie. still in ${<string>} format)
func isSet(s ...string) bool {
	for _, v := range s {
		if v == "" || strings.HasPrefix(v, "${") {
			return false
		}
	}

	return true
}
