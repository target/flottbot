// SPDX-License-Identifier: Apache-2.0

package models

import "github.com/slack-go/slack"

const (
	ChatAppSlack      = "slack"
	ChatAppDiscord    = "discord"
	ChatAppMattermost = "mattermost"
	ChatAppTelegram   = "telegram"
	ChatAppGoogleChat = "google_chat"
)

// Remotes is a struct that holds data for various remotes.
type Remotes struct {
	Slack   SlackConfig   `mapstructure:"slack" binding:"omitempty"`
	Discord DiscordConfig `mapstructure:"discord" binding:"omitempty"`
}

// SlackConfig is a support struct that holds Slack specific data.
type SlackConfig struct {
	Attachments []slack.Attachment `mapstructure:"attachments"`
}

// DiscordConfig is a support struct that holds DiscordConfig specific data.
type DiscordConfig struct {
	// Discord things
}
