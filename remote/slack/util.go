// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package slack

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/target/flottbot/models"
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

// isValidPath - regex matches a URL's path string to check if it is a correct path.
func isValidPath(path string) bool {
	pathPattern := regexp.MustCompile(`^([a-z][a-z0-9+\-.]*:(//[^/?#]+)?)?([a-z0-9\-._~%!$&'()*+,;=:@/]*)`)
	matches := pathPattern.FindAllString(path, -1)

	if matches != nil {
		if matches[0] == path {
			return true
		}
	}

	return false
}

// removeBotMention - parse out the prepended bot mention in a message.
func removeBotMention(contents, botID string) (string, bool) {
	mention := fmt.Sprintf("<@%s> ", botID)
	wasMentioned := false

	if strings.HasPrefix(contents, mention) {
		contents = strings.Replace(contents, mention, "", -1)
		contents = strings.TrimSpace(contents)
		wasMentioned = true
	}

	return contents, wasMentioned
}

// sanitizeContents - sanitizes a buffer's contents from incoming http payloads.
func sanitizeContents(b []byte) (string, error) {
	contents := string(b)
	contents = strings.Replace(contents, "payload=", "", 1)

	contents, err := url.QueryUnescape(contents)
	if err != nil {
		return "", err
	}

	contents = strings.Replace(contents, `\/`, `/`, -1)

	return contents, nil
}
