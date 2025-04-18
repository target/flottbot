// SPDX-License-Identifier: Apache-2.0

package core

import (
	"testing"

	"github.com/target/flottbot/models"
)

func Test_configureChatApplication(t *testing.T) {
	type args struct {
		bot *models.Bot
	}

	testBot := new(models.Bot)
	testBot.CLI = true
	validateRemoteSetup(testBot)

	testBotNoChat := new(models.Bot)
	testBotNoChat.CLI = true
	testBotNoChat.ChatApplication = ""
	validateRemoteSetup(testBotNoChat)

	testBotInvalidChat := new(models.Bot)
	testBotInvalidChat.CLI = true
	testBotInvalidChat.ChatApplication = "fart"
	validateRemoteSetup(testBotInvalidChat)

	testBotSlackNoToken := new(models.Bot)
	testBotSlackNoToken.CLI = true
	testBotSlackNoToken.ChatApplication = models.ChatAppSlack
	validateRemoteSetup(testBotSlackNoToken)

	testBotBadName := new(models.Bot)
	testBotBadName.CLI = true
	testBotBadName.ChatApplication = models.ChatAppSlack
	testBotBadName.Name = "${BOT_NAME}"
	validateRemoteSetup(testBotBadName)

	testBotSlackBadToken := new(models.Bot)
	testBotSlackBadToken.CLI = true
	testBotSlackBadToken.ChatApplication = models.ChatAppSlack
	testBotSlackBadToken.SlackToken = "${TOKEN}"
	validateRemoteSetup(testBotSlackBadToken)

	testBotSlackBadSigningSecret := new(models.Bot)
	testBotSlackBadSigningSecret.CLI = true
	testBotSlackBadSigningSecret.ChatApplication = models.ChatAppSlack
	testBotSlackBadSigningSecret.SlackToken = "${TOKEN}"
	testBotSlackBadSigningSecret.SlackSigningSecret = "${TEST_BAD_SIGNING_SECRET}"
	validateRemoteSetup(testBotSlackBadSigningSecret)

	testBotSlack := new(models.Bot)
	testBotSlack.CLI = true
	testBotSlack.ChatApplication = models.ChatAppSlack
	testBotSlack.SlackToken = "${TEST_SLACK_TOKEN}"
	testBotSlack.SlackAppToken = "${TEST_SLACK_APP_TOKEN}"

	t.Setenv("TEST_SLACK_TOKEN", "TESTTOKEN")
	t.Setenv("TEST_SLACK_APP_TOKEN", "TESTAPPTOKEN")

	validateRemoteSetup(testBotSlack)

	testBotDiscordNoToken := new(models.Bot)
	testBotDiscordNoToken.CLI = true
	testBotDiscordNoToken.ChatApplication = models.ChatAppDiscord
	validateRemoteSetup(testBotDiscordNoToken)

	testBotDiscordBadToken := new(models.Bot)
	testBotDiscordBadToken.CLI = true
	testBotDiscordBadToken.ChatApplication = models.ChatAppDiscord
	testBotDiscordBadToken.DiscordToken = "${TOKEN}"
	validateRemoteSetup(testBotDiscordBadToken)

	testBotDiscordServerID := new(models.Bot)
	testBotDiscordServerID.CLI = true
	testBotDiscordServerID.ChatApplication = models.ChatAppDiscord
	testBotDiscordServerID.DiscordToken = "${TEST_DISCORD_TOKEN}"
	testBotDiscordServerID.DiscordServerID = "${TEST_DISCORD_SERVER_ID}"

	t.Setenv("TEST_DISCORD_TOKEN", "TESTTOKEN")
	t.Setenv("TEST_DISCORD_SERVER_ID", "TESTSERVERID")

	validateRemoteSetup(testBotDiscordServerID)

	testBotDiscordBadServerID := new(models.Bot)
	testBotDiscordBadServerID.CLI = true
	testBotDiscordBadServerID.ChatApplication = models.ChatAppDiscord
	testBotDiscordBadServerID.DiscordToken = "${TEST_DISCORD_TOKEN}"
	testBotDiscordBadServerID.DiscordServerID = "${TOKEN}"

	validateRemoteSetup(testBotDiscordServerID)

	testBotTelegram := new(models.Bot)
	testBotTelegram.CLI = true
	testBotTelegram.ChatApplication = models.ChatAppTelegram
	testBotTelegram.TelegramToken = "${TEST_TELEGRAM_TOKEN}"

	t.Setenv("TEST_TELEGRAM_TOKEN", "TESTTOKEN")

	validateRemoteSetup(testBotTelegram)

	testBotTelegramNoToken := new(models.Bot)
	testBotTelegramNoToken.CLI = true
	testBotTelegramNoToken.ChatApplication = models.ChatAppTelegram
	validateRemoteSetup(testBotTelegramNoToken)

	testBotTelegramBadToken := new(models.Bot)
	testBotTelegramBadToken.CLI = true
	testBotTelegramBadToken.ChatApplication = models.ChatAppTelegram
	testBotTelegramBadToken.TelegramToken = "${TOKEN}"
	validateRemoteSetup(testBotTelegramBadToken)

	tests := []struct {
		name          string
		args          args
		shouldRunChat bool
	}{
		{"Fail", args{bot: testBot}, false},
		{"Fail - no chat_application not set", args{bot: testBotNoChat}, false},
		{"Fail - Invalid value for chat_application", args{bot: testBotInvalidChat}, false},
		{"Bad Name", args{bot: testBotBadName}, false},
		{"Slack - no token", args{bot: testBotSlackNoToken}, false},
		{"Slack - bad token", args{bot: testBotSlackBadToken}, false},
		{"Slack - bad signing secret", args{bot: testBotSlackBadSigningSecret}, false},
		{models.ChatAppSlack, args{bot: testBotSlack}, true},
		{"Discord - no token", args{bot: testBotDiscordNoToken}, false},
		{"Discord - bad token", args{bot: testBotDiscordBadToken}, false},
		{"Discord w/ server id", args{bot: testBotDiscordServerID}, true},
		{"Discord w/ bad server id", args{bot: testBotDiscordBadServerID}, false},
		{models.ChatAppTelegram, args{bot: testBotTelegram}, true},
		{"Telegram - no token", args{bot: testBotTelegramNoToken}, false},
		{"Telegram - bad token", args{bot: testBotTelegramBadToken}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configureChatApplication(tt.args.bot)

			if tt.shouldRunChat != tt.args.bot.RunChat {
				t.Errorf("configureChatApplication() wanted RunChat set to %v, but got %v", tt.shouldRunChat, tt.args.bot.RunChat)
			}
		})
	}
}

func Test_setSlackListenerPort(t *testing.T) {
	t.Setenv("TEST_SLACK_TOKEN", "TESTTOKEN")
	t.Setenv("TEST_SLACK_INTERACTIONS_CALLBACK_PATH", "TESTPATH")

	baseBot := func() *models.Bot {
		bot := new(models.Bot)
		bot.CLI = true
		bot.ChatApplication = models.ChatAppSlack
		bot.SlackToken = "${TEST_SLACK_TOKEN}"
		bot.SlackInteractionsCallbackPath = "${TEST_SLACK_INTERACTIONS_CALLBACK_PATH}"

		return bot
	}

	t.Run("slack listener port reads from env var config", func(t *testing.T) {
		bot := baseBot()
		bot.SlackListenerPort = "${TEST_SLACK_LISTENER_PORT}"

		t.Setenv("TEST_SLACK_LISTENER_PORT", "TESTPORT")

		validateRemoteSetup(bot)
		configureChatApplication(bot)

		expected := "TESTPORT"
		actual := bot.SlackListenerPort

		if expected != actual {
			t.Errorf("configureChatApplication() wanted SlackListenerPort set to %v, but got %v", expected, actual)
		}
	})

	t.Run("slack listener port defaults if config is not supplied", func(t *testing.T) {
		bot := baseBot()
		validateRemoteSetup(bot)
		configureChatApplication(bot)

		expected := defaultSlackListenerPort
		actual := bot.SlackListenerPort

		if expected != actual {
			t.Errorf("configureChatApplication() wanted SlackListenerUnsetPortVar set to %v, but got %v", expected, actual)
		}
	})

	t.Run("slack listener port defaults if expected env var is empty", func(t *testing.T) {
		bot := baseBot()
		bot.SlackListenerPort = "${TEST_SLACK_LISTENER_PORT}"
		validateRemoteSetup(bot)
		configureChatApplication(bot)

		expected := defaultSlackListenerPort
		actual := bot.SlackListenerPort

		if expected != actual {
			t.Errorf("configureChatApplication() wanted SlackListenerNoPort set to %v, but got %v", expected, actual)
		}
	})
}

func Test_validateRemoteSetup(t *testing.T) {
	type args struct {
		bot *models.Bot
	}

	// testBot := new(models.Bot)

	testBotCLI := new(models.Bot)
	testBotCLI.CLI = true

	testBotCLIChat := new(models.Bot)
	testBotCLIChat.CLI = true
	testBotCLIChat.ChatApplication = models.ChatAppSlack

	testBotCLIChatScheduler := new(models.Bot)
	testBotCLIChatScheduler.CLI = true
	testBotCLIChatScheduler.ChatApplication = models.ChatAppSlack
	testBotCLIChatScheduler.Scheduler = true

	testBotChatScheduler := new(models.Bot)
	testBotChatScheduler.ChatApplication = models.ChatAppSlack
	testBotChatScheduler.Scheduler = true

	testBotCLIChatSchedulerFail := new(models.Bot)
	testBotCLIChatSchedulerFail.CLI = true
	testBotCLIChatSchedulerFail.ChatApplication = ""
	testBotCLIChatSchedulerFail.Scheduler = true

	testBotCLIScheduler := new(models.Bot)
	testBotCLIScheduler.CLI = true
	testBotCLIScheduler.Scheduler = true

	testNoChatNoCLI := new(models.Bot)
	testNoChatNoCLI.CLI = false
	testNoChatNoCLI.ChatApplication = ""

	tests := []struct {
		name               string
		args               args
		shouldRunChat      bool
		shouldRunCLI       bool
		shouldRunScheduler bool
	}{
		// {"Nothing should run", args{bot: testBot}, false, false, false}, // this should cause fatal exit
		{"CLI Only", args{bot: testBotCLI}, false, true, false},
		{"CLI + Chat", args{bot: testBotCLIChat}, true, true, false},
		// {"No CLI + No Chat", args{bot: testNoChatNoCLI}, false, false, false}, // this will Fatal out
		{"CLI + Chat + Scheduler", args{bot: testBotCLIChatScheduler}, true, true, true},
		{"CLI + Scheduler is not valid without Chat", args{bot: testBotCLIScheduler}, false, true, false},
		{"Chat + Scheduler", args{bot: testBotChatScheduler}, true, false, true},
		{"Invalid Chat + Scheduler", args{bot: testBotCLIChatSchedulerFail}, false, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validateRemoteSetup(tt.args.bot)

			if tt.shouldRunChat != tt.args.bot.RunChat {
				t.Errorf("validateRemoteSetup() wanted RunChat set to %v, but got %v", tt.shouldRunChat, tt.args.bot.RunChat)
			}

			if tt.shouldRunCLI != tt.args.bot.RunCLI {
				t.Errorf("validateRemoteSetup() wanted RunCLI set to %v, but got %v", tt.shouldRunCLI, tt.args.bot.RunCLI)
			}

			if tt.shouldRunScheduler != tt.args.bot.RunScheduler {
				t.Errorf("validateRemoteSetup() wanted RunScheduler set to %v, but got %v", tt.shouldRunScheduler, tt.args.bot.RunScheduler)
			}
		})
	}
}

func TestConfigure(t *testing.T) {
	testBot := new(models.Bot)
	testBot.Name = "mybot(${FB_ENV})"
	testBot.CLI = true

	t.Setenv("FB_ENV", "dev")

	type args struct {
		bot *models.Bot
	}

	tests := []struct {
		name   string
		args   args
		expect args
	}{
		{"Basic", args{bot: testBot}, args{bot: &models.Bot{Name: "mybot(dev)"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Configure(tt.args.bot)

			if tt.args.bot.Name != tt.expect.bot.Name {
				t.Errorf("configure() wanted bot.Name set to %v, but got %v", tt.args.bot.Name, tt.expect.bot.Name)
			}
		})
	}
}
