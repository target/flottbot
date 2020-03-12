package core

import (
	"os"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/target/flottbot/models"
)

func TestInitLogger(t *testing.T) {
	type args struct {
		bot *models.Bot
	}

	testBot := new(models.Bot)

	// Test setting the error and debug level flags
	levelTests := []struct {
		name string
		args args
		want string
	}{
		{"error level set", args{testBot}, "error"},
		{"debug level set", args{testBot}, "debug"},
	}
	for _, tt := range levelTests {
		if tt.want == "debug" {
			testBot.Debug = true
		}
		t.Run(tt.name, func(t *testing.T) {
			initLogger(tt.args.bot)
			if tt.want != tt.args.bot.Log.Level.String() {
				t.Errorf("initLogger() wanted level set at %s, but got %s", tt.want, tt.args.bot.Log.Level.String())
			}
		})
	}

	// Test setting the JSON formatter
	jsonTests := []struct {
		name string
		args args
		want bool
	}{
		{"JSON logging set", args{testBot}, true},
		{"JSON logging not set", args{testBot}, false},
	}
	for _, tt := range jsonTests {
		testBot.LogJSON = tt.want
		t.Run(tt.name, func(t *testing.T) {
			initLogger(tt.args.bot)
			equals := reflect.DeepEqual(tt.args.bot.Log.Formatter, logrus.JSONFormatter{})
			if equals {
				t.Errorf("initLogger() wanted to set JSON logging formatter to %t, but got %t", tt.want, equals)
			}

		})
	}
}

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
	testBotSlackNoToken.ChatApplication = "slack"
	validateRemoteSetup(testBotSlackNoToken)

	testBotBadName := new(models.Bot)
	testBotBadName.CLI = true
	testBotBadName.ChatApplication = "slack"
	testBotBadName.Name = "${BOT_NAME}"
	validateRemoteSetup(testBotBadName)

	testBotSlackBadToken := new(models.Bot)
	testBotSlackBadToken.CLI = true
	testBotSlackBadToken.ChatApplication = "slack"
	testBotSlackBadToken.SlackToken = "${TOKEN}"
	validateRemoteSetup(testBotSlackBadToken)

	testBotSlackBadVerificationToken := new(models.Bot)
	testBotSlackBadVerificationToken.CLI = true
	testBotSlackBadVerificationToken.ChatApplication = "slack"
	testBotSlackBadVerificationToken.SlackToken = "${TOKEN}"
	testBotSlackBadVerificationToken.SlackVerificationToken = "${TEST_BAD_VERIFICATION_TOKEN}"
	validateRemoteSetup(testBotSlackBadVerificationToken)

	testBotSlackBadWorkspaceToken := new(models.Bot)
	testBotSlackBadWorkspaceToken.CLI = true
	testBotSlackBadWorkspaceToken.ChatApplication = "slack"
	testBotSlackBadWorkspaceToken.SlackToken = "${TOKEN}"
	testBotSlackBadWorkspaceToken.SlackWorkspaceToken = "${TEST_BAD_WORKSPACE_TOKEN}"
	validateRemoteSetup(testBotSlackBadWorkspaceToken)

	testBotSlack := new(models.Bot)
	testBotSlack.CLI = true
	testBotSlack.ChatApplication = "slack"
	testBotSlack.SlackToken = "${TEST_SLACK_TOKEN}"
	os.Setenv("TEST_SLACK_TOKEN", "TESTTOKEN")
	validateRemoteSetup(testBotSlack)

	testBotSlackInteraction := new(models.Bot)
	testBotSlackInteraction.CLI = true
	testBotSlackInteraction.InteractiveComponents = true
	testBotSlackInteraction.ChatApplication = "slack"
	testBotSlackInteraction.SlackToken = "${TEST_SLACK_TOKEN}"
	testBotSlackInteraction.SlackInteractionsCallbackPath = "${TEST_SLACK_INTERACTIONS_CALLBACK_PATH}"
	os.Setenv("TEST_SLACK_TOKEN", "TESTTOKEN")
	os.Setenv("TEST_SLACK_INTERACTIONS_CALLBACK_PATH", "TESTPATH")
	validateRemoteSetup(testBotSlackInteraction)

	testBotSlackInteractionFail := new(models.Bot)
	testBotSlackInteractionFail.CLI = true
	testBotSlackInteractionFail.InteractiveComponents = true
	testBotSlackInteractionFail.ChatApplication = "slack"
	testBotSlackInteractionFail.SlackToken = "${TEST_SLACK_TOKEN}"
	testBotSlackInteractionFail.SlackInteractionsCallbackPath = "${TEST_SLACK_INTERACTIONS_CALLBACK_PATH_FAIL}"
	os.Setenv("TEST_SLACK_TOKEN", "TESTTOKEN")
	os.Setenv("TEST_SLACK_INTERACTIONS_CALLBACK_PATH_FAIL", "")
	validateRemoteSetup(testBotSlackInteractionFail)

	testBotSlackEventsCallbackFail := new(models.Bot)
	testBotSlackEventsCallbackFail.CLI = true
	testBotSlackEventsCallbackFail.InteractiveComponents = true
	testBotSlackEventsCallbackFail.ChatApplication = "slack"
	testBotSlackEventsCallbackFail.SlackToken = "${TEST_SLACK_TOKEN}"
	testBotSlackEventsCallbackFail.SlackInteractionsCallbackPath = "${TEST_SLACK_INTERACTIONS_CALLBACK_PATH_FAIL}"
	testBotSlackEventsCallbackFail.SlackEventsCallbackPath = "${TEST_SLACK_EVENTS_CALLBACK_PATH_FAIL}"
	validateRemoteSetup(testBotSlackEventsCallbackFail)

	testBotDiscordNoToken := new(models.Bot)
	testBotDiscordNoToken.CLI = true
	testBotDiscordNoToken.ChatApplication = "discord"
	validateRemoteSetup(testBotDiscordNoToken)

	testBotDiscordBadToken := new(models.Bot)
	testBotDiscordBadToken.CLI = true
	testBotDiscordBadToken.ChatApplication = "discord"
	testBotDiscordBadToken.DiscordToken = "${TOKEN}"
	validateRemoteSetup(testBotDiscordBadToken)

	testBotDiscordServerID := new(models.Bot)
	testBotDiscordServerID.CLI = true
	testBotDiscordServerID.ChatApplication = "discord"
	testBotDiscordServerID.DiscordToken = "${TEST_DISCORD_TOKEN}"
	testBotDiscordServerID.DiscordServerID = "${TEST_DISCORD_SERVER_ID}"
	os.Setenv("TEST_DISCORD_TOKEN", "TESTTOKEN")
	os.Setenv("TEST_DISCORD_SERVER_ID", "TESTSERVERID")
	validateRemoteSetup(testBotDiscordServerID)

	testBotDiscordBadServerID := new(models.Bot)
	testBotDiscordBadServerID.CLI = true
	testBotDiscordBadServerID.ChatApplication = "discord"
	testBotDiscordBadServerID.DiscordToken = "${TEST_DISCORD_TOKEN}"
	testBotDiscordBadServerID.DiscordServerID = "${TOKEN}"
	validateRemoteSetup(testBotDiscordServerID)

	tests := []struct {
		name                           string
		args                           args
		shouldRunChat                  bool
		shouldRunInteractiveComponents bool
	}{
		{"Fail", args{bot: testBot}, false, false},
		{"Fail - no chat_application not set", args{bot: testBotNoChat}, false, false},
		{"Fail - Invalid value for chat_application", args{bot: testBotInvalidChat}, false, false},
		{"Bad Name", args{bot: testBotBadName}, false, false},
		{"Slack - no token", args{bot: testBotSlackNoToken}, false, false},
		{"Slack - bad token", args{bot: testBotSlackBadToken}, false, false},
		{"Slack - bad verification token", args{bot: testBotSlackBadVerificationToken}, false, false},
		{"Slack - bad workspace token", args{bot: testBotSlackBadWorkspaceToken}, false, false},
		{"Slack", args{bot: testBotSlack}, true, false},
		{"Slack w/ interaction", args{bot: testBotSlackInteraction}, true, true},
		{"Slack w/ interaction - empty path", args{bot: testBotSlackInteractionFail}, true, false},
		{"Slack w/ bad events callback", args{bot: testBotSlackEventsCallbackFail}, true, false},
		{"Discord - no token", args{bot: testBotDiscordNoToken}, false, false},
		{"Discord - bad token", args{bot: testBotDiscordBadToken}, false, false},
		{"Discord w/ server id", args{bot: testBotDiscordServerID}, true, false},
		{"Discord w/ bad server id", args{bot: testBotDiscordBadServerID}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configureChatApplication(tt.args.bot)
			if tt.shouldRunChat != tt.args.bot.RunChat {
				t.Errorf("configureChatApplication() wanted RunChat set to %v, but got %v", tt.shouldRunChat, tt.args.bot.RunChat)
			}

			if tt.shouldRunInteractiveComponents != tt.args.bot.InteractiveComponents {
				t.Errorf("configureChatApplication() wanted InteractiveComponents set to %v, but got %v", tt.shouldRunInteractiveComponents, tt.args.bot.InteractiveComponents)
			}
		})
	}

	os.Unsetenv("TEST_SLACK_TOKEN")
	os.Unsetenv("TEST_DISCORD_TOKEN")
	os.Unsetenv("TEST_DISCORD_SERVER_ID")
	os.Unsetenv("TEST_SLACK_INTERACTIONS_CALLBACK_PATH")
	os.Unsetenv("TEST_SLACK_INTERACTIONS_CALLBACK_PATH_FAIL")
}

func Test_setSlackListenerPort(t *testing.T) {
	t.Run("slack listener port reads from env var config", func(t *testing.T) {
		testBotSlackListenerPort := new(models.Bot)
		os.Setenv("TEST_SLACK_LISTENER_PORT", "TESTPORT")
		validateRemoteSetup(testBotSlackListenerPort)

		configureChatApplication(testBotSlackListenerPort)
		expected := "TESTPORT"
		actual := testBotSlackListenerPort.SlackListenerPort
		if expected != actual {
			t.Errorf("configureChatApplication() wanted SlackListenerPort set to %v, but got %v", expected, actual)
		}
	})

	t.Run("slack listener port defaults if config is not supplied", func(t *testing.T) {
		os.Unsetenv("TEST_SLACK_LISTENER_PORT")
		testBotSlackListenerNoPort := new(models.Bot)
		testBotSlackListenerNoPort.SlackListenerPort = "${TEST_SLACK_LISTENER_PORT}"
		validateRemoteSetup(testBotSlackListenerNoPort)

		configureChatApplication(testBotSlackListenerNoPort)
		expected := defaultSlackListenerPort
		actual := testBotSlackListenerNoPort.SlackListenerPort
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
	testBotCLIChat.ChatApplication = "slack"

	testBotCLIChatScheduler := new(models.Bot)
	testBotCLIChatScheduler.CLI = true
	testBotCLIChatScheduler.ChatApplication = "slack"
	testBotCLIChatScheduler.Scheduler = true

	testBotChatScheduler := new(models.Bot)
	testBotChatScheduler.ChatApplication = "slack"
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

	os.Setenv("FB_ENV", "dev")
	defer os.Unsetenv("FB_ENV")

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
