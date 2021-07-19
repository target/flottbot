package core

import (
	"strings"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/remote/cli"
	"github.com/target/flottbot/remote/discord"
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
					bot.Log.Warn().Msg("Scheduler does not currently support Discord")
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
			default:
				bot.Log.Debug().Msgf("Chat application %s is not supported", chatApp)
			}
		case models.MsgServiceCLI:
			remoteCLI := &cli.Client{}
			remoteCLI.Send(message, bot)
		case models.MsgServiceUnknown:
			bot.Log.Error().Msg("Found unknown service")
		default:
			bot.Log.Error().Msg("No service found")
		}
	}
}
