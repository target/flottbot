package models

// Action defines the structure for Actions used within Rules
type Action struct {
	Name             string                 `yaml:"name" binding:"required"`
	Type             string                 `yaml:"type" binding:"required"`
	URL              string                 `yaml:"url"`
	Cmd              string                 `yaml:"cmd"`
	Timeout          int                    `yaml:"timeout"`
	QueryData        map[string]interface{} `yaml:"query_data"`
	CustomHeaders    map[string]string      `yaml:"custom_headers"`
	Auth             []Auth                 `yaml:"auth"`
	ExposeJSONFields map[string]string      `yaml:"expose_json_fields"`
	Response         string                 `yaml:"response"`
	LimitToRooms     []string               `yaml:"limit_to_rooms"`
	Message          string                 `yaml:"message"`
	Reaction         string                 `yaml:"update_reaction" binding:"omitempty"`
}

// Auth is a basic Auth data structure
type Auth struct {
	Type string `yaml:"type"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}
