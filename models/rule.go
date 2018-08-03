package models

// Rule is a struct representation of the .yml rules
type Rule struct {
	Name               string   `yaml:"name" binding:"required"`
	Respond            string   `yaml:"respond" binding:"omitempty"`
	Hear               string   `yaml:"hear" binding:"omitempty"`
	Schedule           string   `json:"schedule" yaml:"schedule"`
	Args               []string `yaml:"args" binding:"required"`
	DirectMessageOnly  bool     `yaml:"direct_message_only" binding:"required"`
	OutputToRooms      []string `yaml:"output_to_rooms" binding:"omitempty"`
	OutputToUsers      []string `yaml:"output_to_users" binding:"omitempty"`
	AllowUsers         []string `yaml:"allow_users" binding:"omitempty"`
	AllowUserGroups    []string `yaml:"allow_usergroups" binding:"omitempty"`
	IgnoreUsers        []string `yaml:"ignore_users" binding:"omitempty"`
	IgnoreUserGroups   []string `yaml:"ignore_usergroups" binding:"omitempty"`
	StartMessageThread bool     `yaml:"start_message_thread" binding:"omitempty"`
	FormatOutput       string   `yaml:"format_output"`
	HelpText           string   `yaml:"help_text"`
	IncludeInHelp      bool     `yaml:"include_in_help" binding:"required"`
	Active             bool     `yaml:"active" binding:"required"`
	Debug              bool     `yaml:"debug" binding:"required"`
	Actions            []Action `yaml:"actions" binding:"required"`
	Remotes            Remotes  `yaml:"remotes" binding:"omitempty"`
	Reaction           string   `yaml:"reaction" binding:"omitempty"`
	// The following fields are not included in rule yaml
	RemoveReaction string
}
