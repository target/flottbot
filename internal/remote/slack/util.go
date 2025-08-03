// SPDX-License-Identifier: Apache-2.0

package slack

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/target/flottbot/internal/models"
)

/*
=======================================================
Utility functions (does not use 'slack-go/slack' package)
=======================================================
*/

// findKey - find the key value in the map based on its value pair.
func findKey(m map[string]string, value string) (key string, ok bool) {
	for k, v := range m {
		if v == value {
			key = k
			ok = true

			return
		}
	}

	return
}

// getMessageType - gets the type of message based on where it came from.
func getMessageType(channel string) (models.MessageType, error) {
	re := regexp.MustCompile(`^(C|D|G)[A-Z0-9]{4,}$`) // match known channel ID types
	match := re.FindStringSubmatch(channel)

	if len(match) > 0 {
		switch match[1] { // [1] grabs the first letter, [0] will grab the entire channel ID
		case "D":
			return models.MsgTypeDirect, nil
		case "C":
			return models.MsgTypeChannel, nil
		case "G":
			return models.MsgTypePrivateChannel, nil
		default:
			return models.MsgTypeUnknown, fmt.Errorf("unable to handle channel: UNKNOWN_%s", channel)
		}
	}

	return models.MsgTypeUnknown, fmt.Errorf("unable to handle channel: UNKNOWN_%s", channel)
}

// removeBotMention - parse out the prepended bot mention in a message.
func removeBotMention(contents, botID string) (string, bool) {
	mention := fmt.Sprintf("<@%s> ", botID)
	wasMentioned := false

	if strings.HasPrefix(contents, mention) {
		contents = strings.ReplaceAll(contents, mention, "")
		contents = strings.TrimSpace(contents)
		wasMentioned = true
	}

	return contents, wasMentioned
}
