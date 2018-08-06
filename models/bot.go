package models

import "github.com/sirupsen/logrus"

// Bot is a struct representation of bot.yml
type Bot struct {
	// Bot YAML fields
	ID                            string            `yaml:"id"`
	Name                          string            `yaml:"name" binding:"required"`
	SlackToken                    string            `yaml:"slack_token"`
	SlackVerificationToken        string            `yaml:"slack_verification_token"`
	SlackWorkspaceToken           string            `yaml:"slack_workspace_token"`
	SlackEventsCallbackPath       string            `yaml:"slack_events_callback_path"`
	SlackInteractionsCallbackPath string            `yaml:"slack_interactions_callback_path"`
	DiscordToken                  string            `yaml:"discord_token"`
	Users                         map[string]string `yaml:"slack_users"`
	UserGroups                    map[string]string `yaml:"slack_usergroups"`
	Rooms                         map[string]string `yaml:"slack_channels"`
	CLI                           bool              `yaml:"cli,omitempty"`
	CLIUser                       string            `yaml:"cli_user,omitempty"`
	Scheduler                     bool              `yaml:"scheduler,omitempty"`
	ChatApplication               string            `yaml:"chat_application" binding:"required"`
	Debug                         bool              `yaml:"debug,omitempty"`
	LogJSON                       bool              `yaml:"log_json,omitempty"`
	InteractiveComponents         bool              `yaml:"interactive_components,omitempty"`
	Metrics                       bool              `yaml:"metrics,omitempty"`
	CustomHelpText                string            `yaml:"custom_help_text,omitempty"`
	// System
	Log          logrus.Logger
	RunChat      bool
	RunCLI       bool
	RunScheduler bool
}
