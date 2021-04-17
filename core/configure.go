package core

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/utils"
)

var defaultSlackListenerPort = "3000"

// Configure searches the config directory for the bot.yml to create a Bot object.
// The Bot object will be passed around to make accessible system-specific information.
func Configure(bot *models.Bot) {
	log.Info("Configuring bot...")

	initLogger(bot)

	validateRemoteSetup(bot)

	configureChatApplication(bot)

	bot.Log.Infof("Configured bot '%s'!", bot.Name)
}

// initLogger sets log configuration for the bot
func initLogger(b *models.Bot) {
	b.Log = *log.New()

	b.Log.SetLevel(log.ErrorLevel)

	if b.Debug {
		b.Log.SetLevel(log.DebugLevel)
	}

	if b.LogJSON {
		b.Log.Formatter = &log.JSONFormatter{}
	}
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
		bot.Log.Warnf("could not configure bot 'name' field: %s", err.Error())
	}

	bot.Name = token

	if bot.ChatApplication != "" {
		switch strings.ToLower(bot.ChatApplication) {
		case "discord":
			// Discord bot token
			token, err := utils.Substitute(bot.DiscordToken, emptyMap)
			if err != nil {
				bot.Log.Errorf("could not set discord_token: %s", err.Error())
				bot.RunChat = false
			}

			bot.DiscordToken = token

			// Discord Server ID
			// See https://support.discordapp.com/hc/en-us/articles/206346498-Where-can-I-find-my-User-Server-Message-ID-
			serverID, err := utils.Substitute(bot.DiscordServerID, emptyMap)
			if err != nil {
				bot.Log.Errorf("could not set discord_server_id: %s", err.Error())
				bot.RunChat = false
			}

			bot.DiscordServerID = serverID

			if !isSet(token, serverID) {
				bot.Log.Error("bot is not configured correctly for discord - check that discord_token and discord_server_id are set")
				bot.RunChat = false
			}

		case "slack":
			configureSlackBot(bot)

		case "telegram":
			token, err := utils.Substitute(bot.TelegramToken, emptyMap)
			if err != nil {
				bot.Log.Errorf("could not set telegram_token: %s", err.Error())
				bot.RunChat = false
			}

			if !isSet(token) {
				bot.Log.Error("bot is not configured correctly for telegram - check that telegram_token is set")
				bot.RunChat = false
			}

			bot.TelegramToken = token

		default:
			bot.Log.Errorf("chat application '%s' is not supported", bot.ChatApplication)
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
		bot.Log.Errorf("could not set slack_token: %s", err.Error())
	}

	bot.SlackToken = token

	// slack_app_token
	appToken, err := utils.Substitute(bot.SlackAppToken, emptyMap)
	if err != nil {
		bot.Log.Warnf("could not set slack_app_token: %s", err.Error())
	}

	bot.SlackAppToken = appToken

	// slack_signing_secret
	signingSecret, err := utils.Substitute(bot.SlackSigningSecret, emptyMap)
	if err != nil {
		bot.Log.Warnf("could not set slack_signing_secret: %s", err.Error())
	}

	bot.SlackSigningSecret = signingSecret

	// slack_events_callback_path
	eCallbackPath, err := utils.Substitute(bot.SlackEventsCallbackPath, emptyMap)
	if err != nil {
		bot.Log.Warnf("could not set slack_events_callback_path: %s", err.Error())
	}

	bot.SlackEventsCallbackPath = eCallbackPath

	// slack_interactions_callback_path
	iCallbackPath, err := utils.Substitute(bot.SlackInteractionsCallbackPath, emptyMap)
	if err != nil {
		bot.Log.Warnf("could not set slack_interactions_callback_path: %s", err.Error())
	}

	bot.SlackInteractionsCallbackPath = iCallbackPath

	// slack_listener_port
	lPort, err := utils.Substitute(bot.SlackListenerPort, emptyMap)
	if err != nil {
		bot.Log.Warnf("could not set slack_listener_port: %s", err.Error())
	}

	// set slack http listener port from config file or default
	if !isSet(lPort) {
		bot.Log.Warnf("slack_listener_port not set: %s", lPort)
		bot.Log.WithField("defaultSlackListenerPort", defaultSlackListenerPort).Info("using default slack listener port.")
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
		bot.Log.Error("bot is not configured correctly for slack - check that either slack_token and slack_app_token OR slack_token, slack_signing_secret, and slack_events_callback_path are set")
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
		bot.Log.Fatalf("No chat_application specified and cli mode is not enabled. Exiting...")
	}

	if bot.Scheduler {
		bot.RunScheduler = true
		if bot.CLI && bot.ChatApplication == "" {
			bot.Log.Warn("Scheduler does not support scheduled outputs to CLI mode")
			bot.RunScheduler = false
		}

		if bot.ChatApplication == "" {
			bot.Log.Warn("Scheduler did not find any configured chat applications. Scheduler is closing")
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
