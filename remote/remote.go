// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package remote

import (
	"context"

	"github.com/target/flottbot/models"
)

// Remote - this interface allows us to keep the bot "remote agnostic" meaning
// that the bot should not care about what specific remote (e.g. Slack or Discord)
// it is reading/sending messages from. It is up to the developer to implement
// the details of how messages should be read/sent in the specific package for
// the remote (e.g. see '/remote/slack/remote.go')
// Each remote will share generic functions as seen below that will be evoked
// by the bot (e.g. see '/core/remotes.go' or '/core/outputs.go'), and each function
// will be implemented in detail within its specific remote package.
type Remote interface {
	Reaction(message models.Message, rule models.Rule, bot *models.Bot)

	Read(inputMsgs chan<- models.Message, rules map[string]models.Rule, bot *models.Bot)

	Send(message models.Message, bot *models.Bot)

	Name() string
}

// Reaction enables the bot to add emoji reactions to messages.
func Reaction(c context.Context, message models.Message, rule models.Rule, bot *models.Bot) {
	FromContext(c).Reaction(message, rule, bot)
}

// Read enables the bot to read messages from a remote.
func Read(c context.Context, inputMsgs chan<- models.Message, rules map[string]models.Rule, bot *models.Bot) {
	FromContext(c).Read(inputMsgs, rules, bot)
}

// Send enables the bot to send messages to a remote.
func Send(c context.Context, message models.Message, bot *models.Bot) {
	FromContext(c).Send(message, bot)
}

// Name returns the name of the remote.
func Name(c context.Context) string {
	return FromContext(c).Name()
}
