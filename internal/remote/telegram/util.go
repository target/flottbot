// SPDX-License-Identifier: Apache-2.0

package telegram

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/target/flottbot/internal/models"
)

// processMessageText removes existing bot mention
// and returns message and whether bot was mentioned.
func processMessageText(m, b string) (string, bool) {
	botMention := fmt.Sprintf("@%s", b)
	isMentioned := strings.HasPrefix(m, botMention)
	message := strings.TrimPrefix(m, botMention)

	return strings.TrimSpace(message), isMentioned
}

// mapMessageType converts the type returned for a
// telegram message to an internal, equivalent type.
func mapMessageType(t tgbotapi.Message) models.MessageType {
	msgType := models.MsgTypeUnknown

	switch t.Chat.Type {
	case "private":
		msgType = models.MsgTypeDirect
	case "privatechannel":
		msgType = models.MsgTypePrivateChannel
	case "channel":
	case "group":
	case "supergroup":
		msgType = models.MsgTypeChannel
	}

	return msgType
}
