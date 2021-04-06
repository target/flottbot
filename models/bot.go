package models

import "github.com/sirupsen/logrus"

// Bot is a struct representation of bot.yml
type Bot struct {
	// Bot fields
	ID                            string            `mapstructure:"id"`
	Name                          string            `mapstructure:"name" binding:"required"`
	SlackToken                    string            `mapstructure:"slack_token"`
	SlackSigningSecret            string            `mapstructure:"slack_signing_secret"`
	SlackEventsCallbackPath       string            `mapstructure:"slack_events_callback_path"`
	SlackInteractionsCallbackPath string            `mapstructure:"slack_interactions_callback_path"`
	SlackListenerPort             string            `mapstructure:"slack_listener_port"`
	DiscordToken                  string            `mapstructure:"discord_token"`
	DiscordServerID               string            `mapstructure:"discord_server_id"`
	TelegramToken                 string            `mapstructure:"telegram_token"`
	Users                         map[string]string `mapstructure:"slack_users"`
	UserGroups                    map[string]string `mapstructure:"slack_usergroups"`
	Rooms                         map[string]string `mapstructure:"slack_channels"`
	CLI                           bool              `mapstructure:"cli,omitempty"`
	CLIUser                       string            `mapstructure:"cli_user,omitempty"`
	Scheduler                     bool              `mapstructure:"scheduler,omitempty"`
	ChatApplication               string            `mapstructure:"chat_application" binding:"required"`
	Debug                         bool              `mapstructure:"debug,omitempty"`
	LogJSON                       bool              `mapstructure:"log_json,omitempty"`
	InteractiveComponents         bool              `mapstructure:"interactive_components,omitempty"`
	Metrics                       bool              `mapstructure:"metrics,omitempty"`
	CustomHelpText                string            `mapstructure:"custom_help_text,omitempty"`
	DisableNoMatchHelp            bool              `mapstructure:"disable_no_match_help,omitempty"`
	RespondToBots                 bool              `mapstructure:"respond_to_bots,omitempty"`
	// System
	Log          logrus.Logger
	RunChat      bool
	RunCLI       bool
	RunScheduler bool
}
