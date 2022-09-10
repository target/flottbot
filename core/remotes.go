// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package core

import (
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/remote/cli"
	"github.com/target/flottbot/remote/discord"
	"github.com/target/flottbot/remote/gchat"
	"github.com/target/flottbot/remote/scheduler"
	"github.com/target/flottbot/remote/slack"
	"github.com/target/flottbot/remote/telegram"
)

// Remotes - the purpose of this function is to READ incoming messages from various places, i.e. remotes.
// Whenever a message is read from a remote, the `inputMsgs` channel will store the read message as a
// 'Message' object and pass it along to the Matcher function (see '/core/matcher.go') for processing.
// Currently, we support 3 types of remotes: chat applications, CLI, and Scheduler.
// Remote 1: Chat applications
//
//	This remote allows us to read messages from various chat application platforms, e.g. Slack, Discord, etc.
//	We typically read the messages from these chat applications using their respective APIs.
//	* Note: right now we only support reading from one chat application at a time.
//
// Remote 2: CLI
//
//	This remote is enabled when 'CLI mode' is set to true in the bot.yml configuration.
//	Messages from this remote are read from the user's input via the terminal.
//
// Remote 3: Scheduler
//
//	This remote allows us to read messages being sent internally by a running cronjob
//	created by a schedule type rule, e.g. see '/config/rules/schedule.yml'.
//
// TODO: Refactor to keep remote specific stuff in remote/.
func Remotes(inputMsgs chan<- models.Message, rules map[string]models.Rule, bot *models.Bot) {
	// Run a chat application
	if bot.ChatApplication != "" {
		chatApp := strings.ToLower(bot.ChatApplication)
		log.Info().Msgf("running %#q on %#q", bot.Name, chatApp)

		switch chatApp {
		// Setup remote to use the Discord client to read from Discord
		case "discord":
			// Create Discord client
			remoteDiscord := &discord.Client{
				Token: bot.DiscordToken,
			}
			// Read messages from Discord
			go remoteDiscord.Read(inputMsgs, rules, bot)
		// Setup remote to use the Slack client to read from Slack
		case "slack":
			// Create Slack client
			remoteSlack := &slack.Client{
				Token:         bot.SlackToken,
				AppToken:      bot.SlackAppToken,
				SigningSecret: bot.SlackSigningSecret,
			}
			// Read messages from Slack
			go remoteSlack.Read(inputMsgs, rules, bot)
		// Setup remote to use the Telegram client to read from Telegram
		case "telegram":
			remoteTelegram := &telegram.Client{
				Token: bot.TelegramToken,
			}
			// Read messages from Telegram
			go remoteTelegram.Read(inputMsgs, rules, bot)
		case "google_chat":
			gchat.HandleRemoteInput(inputMsgs, rules, bot)
		default:
			log.Error().Msgf("chat application %#q is not supported", chatApp)
		}
	}

	// Run CLI mode
	if bot.CLI {
		log.Info().Msgf("running cli mode for %#q", bot.Name)

		remoteCLI := &cli.Client{}

		go remoteCLI.Read(inputMsgs, rules, bot)
	}

	// Run Scheduler
	// CAUTION: Will not work properly when multiple instances of your bot are deployed (i.e. will get duplicated scheduled output)
	if bot.Scheduler {
		log.Info().Msgf("running scheduler for %#q", bot.Name)

		remoteScheduler := &scheduler.Client{}

		go remoteScheduler.Read(inputMsgs, rules, bot)
	}
}
