package discord

import (
	"fmt"
	"strings"
)

/*
================================================
Utility functions (does not use discord package)
================================================
*/

// findKey - find the key value in the map based on its value pair
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

// removeBotMention - parse out the preppended bot mention in a message
func removeBotMention(contents, botID string) (string, bool) {
	mention := fmt.Sprintf("<@!%s>", botID)
	wasMentioned := false

	if strings.HasPrefix(contents, mention) {
		contents = strings.Replace(contents, mention, "", -1)
		contents = strings.TrimSpace(contents)
		wasMentioned = true
	}

	return contents, wasMentioned
}
