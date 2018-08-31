package models

import "github.com/nlopes/slack"

// Remotes is a struct that holds data for various remotes
type Remotes struct {
	Slack   SlackConfig   `yaml:"slack" binding:"omitempty"`
	Discord DiscordConfig `yaml:"discord" binding:"omitempty"`
}

// SlackConfig is a support struct that holds Slack specific data
type SlackConfig struct {
	Attachments []slack.Attachment `yaml:"attachments"`
}

// DiscordConfig is a support struct that holds DiscordConfig specific data
type DiscordConfig struct {
	// Discord things
}
