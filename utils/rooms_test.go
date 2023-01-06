// Copyright (c) 2023 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package utils

import (
	"reflect"
	"testing"

	"github.com/target/flottbot/models"
)

func TestGetRoomIDs(t *testing.T) {
	type args struct {
		wantRooms []string
		bot       *models.Bot
	}

	// For Room Exists
	RoomExistsIn := []string{"testing", "testing-room"}

	RoomExistsActive := make(map[string]string)
	RoomExistsActive["testing"] = "123"
	RoomExistsActive["testing-room"] = "456"

	RoomExistsWant := []string{"123", "456"}

	// For Room Doesn't Exist
	RoomDoesNotExistIn := []string{"not"}

	RoomDoesNotExistActive := make(map[string]string)
	RoomDoesNotExistActive["testing"] = "123"
	RoomDoesNotExistActive["testing-room"] = "456"

	RoomDoesNotExistWant := []string{}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"Basic", args{}, []string{}},
		{"Room exists", args{wantRooms: RoomExistsIn, bot: &models.Bot{Rooms: RoomExistsActive}}, RoomExistsWant},
		{"Room does not exist", args{wantRooms: RoomDoesNotExistIn, bot: &models.Bot{Rooms: RoomDoesNotExistActive}}, RoomDoesNotExistWant},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRoomIDs(tt.args.wantRooms, tt.args.bot); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRoomIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}
