// SPDX-License-Identifier: Apache-2.0

package core

import (
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/remote/cli"
	"github.com/target/flottbot/remote/discord"
	"github.com/target/flottbot/remote/gchat"
	"github.com/target/flottbot/remote/mattermost"
	"github.com/target/flottbot/remote/slack"
	"github.com/target/flottbot/remote/telegram"
)

// Outputs determines where messages are output based on fields set in the bot.yml
// TODO: Refactor to keep remote specifics in remote/.
func Outputs(outputMsgs <-chan models.Message, hitRule <-chan models.Rule, bot *models.Bot) {
	for {
		message := <-outputMsgs
		rule := <-hitRule
		service := message.Service

		switch service {
		case models.MsgServiceChat, models.MsgServiceScheduler:
			chatApp := strings.ToLower(bot.ChatApplication)

			switch chatApp {
			case models.ChatAppDiscord:
				if service == models.MsgServiceScheduler {
					log.Warn().Msg("scheduler does not currently support discord")
					break
				}

				remoteDiscord := &discord.Client{Token: bot.DiscordToken}
				remoteDiscord.Reaction(message, rule, bot)
				remoteDiscord.Send(message, bot)
			case models.ChatAppMattermost:
				remoteMM := &mattermost.Client{
					Server: bot.MatterMostServer,
					Token:  bot.MatterMostToken,
				}
				if strings.ToLower(bot.MatterMostInsecureProtocol) == "1" {
					remoteMM.Insecure = true
				}

				remoteMM.Send(message, bot)
			case models.ChatAppSlack:
				// Create Slack client
				remoteSlack := &slack.Client{
					ListenerPort:  bot.SlackListenerPort,
					Token:         bot.SlackToken,
					AppToken:      bot.SlackAppToken,
					SigningSecret: bot.SlackSigningSecret,
				}

				if service == models.MsgServiceChat {
					remoteSlack.Reaction(message, rule, bot)
				}

				remoteSlack.Send(message, bot)
			case models.ChatAppTelegram:
				remoteTelegram := &telegram.Client{
					Token: bot.TelegramToken,
				}
				remoteTelegram.Send(message, bot)
			case models.ChatAppGoogleChat:
				gchat.HandleRemoteOutput(message, bot)
			default:
				log.Error().Msgf("chat application %#q is not supported", chatApp)
			}
		case models.MsgServiceCLI:
			remoteCLI := &cli.Client{}
			remoteCLI.Send(message, bot)
		case models.MsgServiceUnknown:
			log.Error().Msg("found unknown service")
		default:
			log.Error().Msg("no service found")
		}
	}
}
