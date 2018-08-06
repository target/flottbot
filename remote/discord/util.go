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

// removeBotMention - parse out the preppended bot mention in a message
func removeBotMention(contents, botID string) (string, bool) {
	mention := fmt.Sprintf("<@%s>", botID)
	wasMentioned := false
	if strings.HasPrefix(contents, mention) {
		contents = strings.Replace(contents, mention, "", -1)
		contents = strings.TrimSpace(contents)
		wasMentioned = true
	}
	return contents, wasMentioned
}
