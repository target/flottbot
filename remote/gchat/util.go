package gchat

import (
	"fmt"

	"github.com/target/flottbot/models"
	"google.golang.org/api/chat/v1"
)

// getMessageType converts the type returned for a
// Google Chat message to an internal, equivalent type
func getMessageType(event chat.DeprecatedEvent) (models.MessageType, error) {
	msgType := models.MsgTypeUnknown

	switch event.Type {
	case "MESSAGE":
		msgType = models.MsgTypeChannel

		if event.Message.Space.SingleUserBotDm {
			msgType = models.MsgTypeDirect
		}

	case "ADDED_TO_SPACE":
		return msgType, fmt.Errorf("event %s not supported", event.Type)

	default:
		return msgType, fmt.Errorf("unable to handle unkown event type: %s", event.Type)
	}

	return msgType, nil

}
