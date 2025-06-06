// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"

	"github.com/target/flottbot/models"
)

func TestCanTrigger(t *testing.T) {
	type args struct {
		currentUserName string
		currentUserID   string
		rule            models.Rule
		bot             *models.Bot
	}

	testBot := new(models.Bot)
	testBot.ChatApplication = models.ChatAppSlack

	discordBot := new(models.Bot)
	discordBot.ChatApplication = models.ChatAppDiscord

	strangeBot := new(models.Bot)
	strangeBot.ChatApplication = "strange"

	tests := []struct {
		name string
		args args
		want bool
	}{
		{"No restrictions", args{"jane.doe", "F123456", models.Rule{}, testBot}, true},
		{"User is allowed", args{"jane.doe", "F123456", models.Rule{AllowUsers: []string{"john.doe", "jane.doe"}}, testBot}, true},
		{"User is allowed by Id", args{"jane.doe", "F123456", models.Rule{AllowUserIDs: []string{"F123456", "G123456"}}, testBot}, true},
		{"User not allowed", args{"jane.doe", "F123456", models.Rule{AllowUsers: []string{"john.doe", "jack.jill"}}, testBot}, false},
		{"User not allowed by Id", args{"jane.doe", "F123456", models.Rule{AllowUserIDs: []string{"H123456", "I123456"}}, testBot}, false},
		{"User is ignored", args{"jane.doe", "F123456", models.Rule{IgnoreUsers: []string{"jane.doe", "jack.jill"}}, testBot}, false},
		{"User not in ignore list", args{"jane.doe", "F123456", models.Rule{IgnoreUsers: []string{"john.doe", "jack.jill"}}, testBot}, true},
		{"User is allowed but ignored", args{"jane.doe", "F123456", models.Rule{AllowUsers: []string{"jane.doe"}, IgnoreUsers: []string{"jane.doe", "jack.jill"}}, testBot}, false},
		{"User is not allowed and ignored", args{"john.doe", "F123456", models.Rule{AllowUsers: []string{"jane.doe"}, IgnoreUsers: []string{"john.doe", "jack.jill"}}, testBot}, false},
		{"Group - Discord - Not supported", args{"jane.doe", "F123456", models.Rule{AllowUserGroups: []string{"admins"}}, discordBot}, false},
		{"Group - Chat network not supported", args{"jane.doe", "F123456", models.Rule{AllowUserGroups: []string{"admins"}}, strangeBot}, false},
		// TODO: figure out how to test this below:
		// {"User in allow group but ignored", args{"jane.doe", "F123456", models.Rule{}, testBot}, false},
		// {"User in ignore group but allowed", args{"jane.doe", "F123456", models.Rule{}, testBot}, false},
		// {"User in ignore group and allow group", args{"jane.doe", "F123456", models.Rule{}, testBot}, false},
		// {"User in allow group and not ignored", args{"jane.doe", "F123456", models.Rule{}, testBot}, true},
		// {"User is not in allow group and not ignored", args{"jane.doe", "F123456", models.Rule{}, testBot}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanTrigger(tt.args.currentUserName, tt.args.currentUserID, tt.args.rule, tt.args.bot); got != tt.want {
				t.Errorf("CanTrigger() = %v, want %v", got, tt.want)
			}
		})
	}
}
