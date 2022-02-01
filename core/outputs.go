package core

import (
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/target/flottbot/models"
	"github.com/target/flottbot/remote/cli"
	"github.com/target/flottbot/remote/discord"
	"github.com/target/flottbot/remote/gchat"
	"github.com/target/flottbot/remote/slack"
	"github.com/target/flottbot/remote/telegram"
)

// Outputs determines where messages are output based on fields set in the bot.yml
// TODO: Refactor to keep remote specifics in remote/
func Outputs(outputMsgs <-chan models.Message, hitRule <-chan models.Rule, bot *models.Bot) {
	for {
		message := <-outputMsgs
		rule := <-hitRule
		service := message.Service

		switch service {
		case models.MsgServiceChat, models.MsgServiceScheduler:
			chatApp := strings.ToLower(bot.ChatApplication)

			switch chatApp {
			case "discord":
				if service == models.MsgServiceScheduler {
					log.Warn().Msg("scheduler does not currently support discord")
					break
				}
				remoteDiscord := &discord.Client{Token: bot.DiscordToken}
				remoteDiscord.Reaction(message, rule, bot)
				remoteDiscord.Send(message, bot)
			case "slack":
				// Create Slack client
				remoteSlack := &slack.Client{
					ListenerPort:  bot.SlackListenerPort,
					Token:         bot.SlackToken,
					AppToken:      bot.SlackAppToken,
					SigningSecret: bot.SlackSigningSecret,
				}
				if service == models.MsgServiceChat {
					if bot.InteractiveComponents {
						remoteSlack.InteractiveComponents(nil, &message, rule, bot)
					}
					remoteSlack.Reaction(message, rule, bot)
				}
				remoteSlack.Send(message, bot)
			case "telegram":
				remoteTelegram := &telegram.Client{
					Token: bot.TelegramToken,
				}
				remoteTelegram.Send(message, bot)
			case "google_chat":
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
