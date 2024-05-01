// SPDX-License-Identifier: Apache-2.0

package core

import (
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/remote/gchat"
	"github.com/target/flottbot/utils"
)

var defaultSlackListenerPort = "3000"

// Configure searches the config directory for the bot.yml to create a Bot object.
// The Bot object will be passed around to make accessible system-specific information.
func Configure(bot *models.Bot) {
	log.Info().Msg("configuring bot...")

	validateRemoteSetup(bot)

	configureChatApplication(bot)

	log.Info().Msgf("configured bot %#q!", bot.Name)
}

// configureChatApplication configures a user's specified chat application
// TODO: Refactor to keep remote specifics in remote/.
func configureChatApplication(bot *models.Bot) {
	// emptyMap for substitute function
	// (it will only replace from env vars)
	emptyMap := map[string]string{}

	// update the bot name
	token, err := utils.Substitute(bot.Name, emptyMap)
	if err != nil {
		log.Warn().Msgf("could not configure bot 'name' field: %s", err.Error())
	}

	bot.Name = token

	if bot.ChatApplication != "" {
		log.Info().Msgf("Looking for chat application '%#q'", bot.ChatApplication)

		switch strings.ToLower(bot.ChatApplication) {
		//nolint:goconst // refactor
		case "discord":
			// Discord bot token
			token, err := utils.Substitute(bot.DiscordToken, emptyMap)
			if err != nil {
				log.Error().Msgf("could not set 'discord_token': %s", err.Error())

				bot.RunChat = false
			}

			bot.DiscordToken = token

			// Discord Server ID
			// See https://support.discordapp.com/hc/en-us/articles/206346498-Where-can-I-find-my-User-Server-Message-ID-
			serverID, err := utils.Substitute(bot.DiscordServerID, emptyMap)
			if err != nil {
				log.Error().Msgf("could not set 'discord_server_id': %s", err.Error())

				bot.RunChat = false
			}

			bot.DiscordServerID = serverID

			if !utils.IsSet(token, serverID) {
				log.Error().Msg("bot is not configured correctly for discord - check that 'discord_token' and 'discord_server_id' are set")

				bot.RunChat = false
			}

		//nolint:goconst // refactor
		case "slack":
			configureSlackBot(bot)

		//nolint:goconst // refactor
		case "mattermost":
			log.Info().Msgf("Configuring remote '%#q'", bot.ChatApplication)
			token, err := utils.Substitute(bot.MatterMostToken, emptyMap)

			if err != nil {
				log.Error().Msgf("could not set 'mattermost_token': %s", err.Error())

				bot.RunChat = false
			}

			bot.MatterMostToken = token

			server, err := utils.Substitute(bot.MatterMostServer, emptyMap)

			if err != nil {
				log.Error().Msgf("could not set 'mattermost_server': %s", err.Error())

				bot.RunChat = false
			}

			bot.MatterMostServer = server

		//nolint:goconst // refactor
		case "telegram":
			token, err := utils.Substitute(bot.TelegramToken, emptyMap)
			if err != nil {
				log.Error().Msgf("could not set 'telegram_token': %s", err.Error())

				bot.RunChat = false
			}

			if !utils.IsSet(token) {
				log.Error().Msg("bot is not configured correctly for telegram - check that 'telegram_token' is set")

				bot.RunChat = false
			}

			bot.TelegramToken = token

		//nolint:goconst // refactor
		case "google_chat":
			gchat.Configure(bot)

		default:
			log.Error().Msgf("chat application %#q is not supported", bot.ChatApplication)
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
		log.Error().Msgf("could not set 'slack_token': %s", err.Error())
	}

	bot.SlackToken = token

	// slack_app_token
	appToken, err := utils.Substitute(bot.SlackAppToken, emptyMap)
	if err != nil {
		log.Warn().Msgf("could not set 'slack_app_token': %s", err.Error())
	}

	bot.SlackAppToken = appToken

	// slack_signing_secret
	signingSecret, err := utils.Substitute(bot.SlackSigningSecret, emptyMap)
	if err != nil {
		log.Warn().Msgf("could not set 'slack_signing_secret': %s", err.Error())
	}

	bot.SlackSigningSecret = signingSecret

	// slack_events_callback_path
	eCallbackPath, err := utils.Substitute(bot.SlackEventsCallbackPath, emptyMap)
	if err != nil {
		log.Warn().Msgf("could not set 'slack_events_callback_path': %s", err.Error())
	}

	bot.SlackEventsCallbackPath = eCallbackPath

	// slack_interactions_callback_path
	iCallbackPath, err := utils.Substitute(bot.SlackInteractionsCallbackPath, emptyMap)
	if err != nil {
		log.Warn().Msgf("could not set 'slack_interactions_callback_path': %s", err.Error())
	}

	bot.SlackInteractionsCallbackPath = iCallbackPath

	// slack_listener_port
	lPort, err := utils.Substitute(bot.SlackListenerPort, emptyMap)
	if err != nil {
		log.Warn().Msgf("could not set 'slack_listener_port': %s", err.Error())
	}

	// set slack http listener port from config file or default
	if !utils.IsSet(lPort) {
		log.Warn().Msgf("'slack_listener_port' not set: %#q", lPort)
		log.Info().Str("defaultSlackListenerPort", defaultSlackListenerPort).Msg("using default slack listener port.")
		lPort = defaultSlackListenerPort
	}

	bot.SlackListenerPort = lPort

	// check for valid setup
	// needs one of the following to be valid
	// 1. SLACK_TOKEN + SLACK_APP_TOKEN (socket mode)
	// 2. SLACK_TOKEN + SLACK_SIGNING_SECRET + SLACK_EVENTS_CALLBACK_PATH (events api)
	isSocketMode := utils.IsSet(token, appToken)
	isEventsAPI := utils.IsSet(token, signingSecret, eCallbackPath)

	if !isSocketMode && !isEventsAPI {
		log.Error().Msg("must have either 'slack_token', 'slack_app_token' or 'slack_token', 'slack_signing_secret', and 'slack_events_callback_path' set")

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
		log.Error().Msgf("no 'chat_application' specified and cli mode is not enabled. exiting...")
	}

	if bot.Scheduler {
		bot.RunScheduler = true
		if bot.CLI && bot.ChatApplication == "" {
			log.Warn().Msg("scheduler does not support scheduled outputs to cli mode")

			bot.RunScheduler = false
		}

		if bot.ChatApplication == "" {
			log.Warn().Msg("scheduler did not find any configured chat applications - scheduler is closing")

			bot.RunScheduler = false
		}
	}
}
