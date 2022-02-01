package gchat

import (
	"github.com/target/flottbot/models"
	"google.golang.org/api/chat/v1"
)

// mapMessageType converts the type returned for a
// Google Chat message to an internal, equivalent type
func mapMessageType(event chat.DeprecatedEvent) models.MessageType {
	msgType := models.MsgTypeUnknown

	switch event.Type {
	case "MESSAGE":
		msgType = models.MsgTypeChannel

		if event.Message.Space.SingleUserBotDm {
			msgType = models.MsgTypeDirect
		}
	}

	return msgType
}
