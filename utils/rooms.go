// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/target/flottbot/models"
)

// GetRoomIDs helps find a room by name, if we have 'cached' it.
func GetRoomIDs(wantRooms []string, bot *models.Bot) []string {
	rooms := []string{}

	for _, room := range wantRooms {
		roomMatch := bot.Rooms[strings.ToLower(room)]
		if roomMatch != "" {
			rooms = append(rooms, roomMatch)
		} else {
			log.Error().Msgf("room %#q does not exist", room)
		}
	}

	return rooms
}
