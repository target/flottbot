// Copyright (c) 2023 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package gchat

import (
	"fmt"

	"google.golang.org/api/chat/v1"

	"github.com/target/flottbot/models"
)

// getMessageType converts the type returned for a
// Google Chat message to an internal, equivalent type.
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
		return msgType, fmt.Errorf("unable to handle unknown event type: %s", event.Type)
	}

	return msgType, nil
}
