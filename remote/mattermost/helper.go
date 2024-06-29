// SPDX-License-Identifier: Apache-2.0

package mattermost

import (
	"fmt"
	"strings"
)

func removeBotMention(message, bot string) (string, bool) {
	mentioned := false
	needle := fmt.Sprintf("@%s", bot)

	if strings.HasPrefix(message, needle) {
		message = strings.ReplaceAll(message, needle, " ")
		message = strings.TrimSpace(message)
		mentioned = true
	}

	return message, mentioned
}
