package models

// Rule is a struct representation of the .yml rules
type Rule struct {
	Name               string   `mapstructure:"name" binding:"required"`
	Respond            string   `mapstructure:"respond" binding:"omitempty"`
	Hear               string   `mapstructure:"hear" binding:"omitempty"`
	Schedule           string   `mapstructure:"schedule"`
	Args               []string `mapstructure:"args" binding:"required"`
	DirectMessageOnly  bool     `mapstructure:"direct_message_only" binding:"required"`
	OutputToRooms      []string `mapstructure:"output_to_rooms" binding:"omitempty"`
	OutputToUsers      []string `mapstructure:"output_to_users" binding:"omitempty"`
	AllowUsers         []string `mapstructure:"allow_users" binding:"omitempty"`
	AllowUserGroups    []string `mapstructure:"allow_usergroups" binding:"omitempty"`
	IgnoreUsers        []string `mapstructure:"ignore_users" binding:"omitempty"`
	IgnoreUserGroups   []string `mapstructure:"ignore_usergroups" binding:"omitempty"`
	StartMessageThread bool     `mapstructure:"start_message_thread" binding:"omitempty"`
	FormatOutput       string   `mapstructure:"format_output"`
	HelpText           string   `mapstructure:"help_text"`
	IncludeInHelp      bool     `mapstructure:"include_in_help" binding:"required"`
	Active             bool     `mapstructure:"active" binding:"required"`
	Debug              bool     `mapstructure:"debug" binding:"required"`
	Actions            []Action `mapstructure:"actions" binding:"required"`
	Remotes            Remotes  `mapstructure:"remotes" binding:"omitempty"`
	Reaction           string   `mapstructure:"reaction" binding:"omitempty"`
	// The following fields are not included in rule yaml
	RemoveReaction string
}
