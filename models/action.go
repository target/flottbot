// Copyright (c) 2023 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package models

// Action defines the structure for Actions used within Rules.
type Action struct {
	Name             string            `mapstructure:"name" binding:"required"`
	Type             string            `mapstructure:"type" binding:"required"`
	URL              string            `mapstructure:"url"`
	Cmd              string            `mapstructure:"cmd"`
	Timeout          int               `mapstructure:"timeout"`
	TimeoutMessage   string            `mapstructure:"timeout_message" binding:"omitempty"`
	QueryData        map[string]any    `mapstructure:"query_data"`
	CustomHeaders    map[string]string `mapstructure:"custom_headers"`
	Auth             []Auth            `mapstructure:"auth"`
	ExposeJSONFields map[string]string `mapstructure:"expose_json_fields"`
	Response         string            `mapstructure:"response"`
	LimitToRooms     []string          `mapstructure:"limit_to_rooms"` // deprecated
	OutputToRooms    []string          `mapstructure:"output_to_rooms"`
	Message          string            `mapstructure:"message"`
	Reaction         string            `mapstructure:"update_reaction" binding:"omitempty"`
}

// Auth is a basic Auth data structure.
type Auth struct {
	Type string `mapstructure:"type"`
	User string `mapstructure:"user"`
	Pass string `mapstructure:"pass"`
}
