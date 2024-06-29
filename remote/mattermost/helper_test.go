// SPDX-License-Identifier: Apache-2.0
package mattermost

import "testing"

func TestRemoveBotMentions(t *testing.T) {

	cases := []struct {
		Description string
		BotID       string
		Message     string
		Mentioned   bool
		WantMsg     string
	}{
		{"mention at beginning of message", "foobar", "@foobar hello", true, "hello"},
		{"mention at end of message", "foobar", "hi @foobar", false, "hi @foobar"},
		{"no mention in message", "foobar", "supercalifragilistic expialidocious", false, "supercalifragilistic expialidocious"},
		{"mention in the middle of the message", "foobar", "hello @foobar barfoo", false, "hello @foobar barfoo"},
		{"contains name but no mention", "foobar", "hello foobar barfoo", false, "hello foobar barfoo"},
	}

	for _, test := range cases {
		t.Run(test.Description, func(t *testing.T) {
			message, mentioned := removeBotMention(test.Message, test.BotID)
			if message != test.WantMsg {
				t.Errorf("Message mismatch: got %q, want %q", message, test.WantMsg)
			}

			if mentioned != test.Mentioned {
				t.Errorf("Mentioned mismatch: got %v, want %v", mentioned, test.Mentioned)
			}
		})
	}

}
