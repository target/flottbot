// SPDX-License-Identifier: Apache-2.0

package models

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Bot is a struct representation of bot.yml.
type Bot struct {
	// Bot fields
	ID                            string            `mapstructure:"id"`
	Name                          string            `mapstructure:"name" binding:"required"`
	SlackToken                    string            `mapstructure:"slack_token"`
	SlackAppToken                 string            `mapstructure:"slack_app_token"`
	SlackSigningSecret            string            `mapstructure:"slack_signing_secret"`
	SlackEventsCallbackPath       string            `mapstructure:"slack_events_callback_path"`
	SlackInteractionsCallbackPath string            `mapstructure:"slack_interactions_callback_path"`
	SlackListenerPort             string            `mapstructure:"slack_listener_port"`
	MatterMostToken               string            `mapstructure:"mattermost_token"`
	MatterMostServer              string            `mapstructure:"mattermost_server"`
	MatterMostInsecureProtocol    string            `mapstructure:"mattermost_insecure"`
	DiscordToken                  string            `mapstructure:"discord_token"`
	DiscordServerID               string            `mapstructure:"discord_server_id"`
	GoogleChatProjectID           string            `mapstructure:"google_chat_project_id"`
	GoogleChatSubscriptionID      string            `mapstructure:"google_chat_subscription_id"`
	GoogleChatCredentials         string            `mapstructure:"google_chat_credentials"`
	GoogleChatForceReplyToThread  bool              `mapstructure:"google_chat_force_reply_to_thread"`
	TelegramToken                 string            `mapstructure:"telegram_token"`
	Users                         map[string]string `mapstructure:"slack_users"`
	UserGroups                    map[string]string `mapstructure:"slack_usergroups"`
	Rooms                         map[string]string `mapstructure:"slack_channels"`
	CLI                           bool              `mapstructure:"cli,omitempty"`
	CLIUser                       string            `mapstructure:"cli_user,omitempty"`
	Scheduler                     bool              `mapstructure:"scheduler,omitempty"`
	ChatApplication               string            `mapstructure:"chat_application" binding:"required"`
	Debug                         bool              `mapstructure:"debug,omitempty"`
	Metrics                       bool              `mapstructure:"metrics,omitempty"`
	CustomHelpText                string            `mapstructure:"custom_help_text,omitempty"`
	CustomHelpTextPrefix          string            `mapstructure:"custom_help_text_prefix,omitempty"`
	DisableNoMatchHelp            bool              `mapstructure:"disable_no_match_help,omitempty"`
	RespondToBots                 bool              `mapstructure:"respond_to_bots,omitempty"`
	// System
	RunChat      bool
	RunCLI       bool
	RunScheduler bool
}

// NewBot creates a new Bot instance.
func NewBot() *Bot {
	v := viper.New()
	bot := new(Bot)

	// set default search locations
	v.AddConfigPath("./config")
	v.AddConfigPath(".")
	v.SetConfigName("bot")

	// read the config
	err := v.ReadInConfig()
	if err != nil {
		log.Fatal().Msgf("could not read bot config: %s", err)
	}

	// unmarshal the config
	err = v.Unmarshal(bot)
	if err != nil {
		log.Fatal().Msgf("could not unmarshal bot config: %s", err)
	}

	// set debug logging
	if bot.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// prettify log for CLI mode
	if bot.CLI {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return bot
}
