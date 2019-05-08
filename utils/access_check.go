package utils

import (
	"fmt"
	"strings"

	"github.com/nlopes/slack"
	"github.com/target/flottbot/models"
)

// CanTrigger ensures the user is allowed to use the respective rule
func CanTrigger(currentUserName string, currentUserID string, rule models.Rule, bot *models.Bot) bool {
	var canRunRule bool

	// no restriction were given for this rule, allow to proceed
	if len(rule.AllowUsers)+len(rule.AllowUserGroups)+len(rule.AllowUserIds)+len(rule.IgnoreUsers)+len(rule.IgnoreUserGroups) == 0 {
		return true
	}

	// are they ignored directly? deny
	for _, name := range rule.IgnoreUsers {
		if name == currentUserName {
			bot.Log.Debugf("'%s' is on the ignore_users list for rule: '%s'", currentUserName, rule.Name)
			return false
		}
	}

	// are they part of a usergroup to be ignored? deny
	isIgnored, err := isMemberOfGroup(currentUserID, rule.IgnoreUserGroups, bot)
	// deny access if unable to check group membership due to error
	if err != nil {
		return false
	}
	if isIgnored {
		bot.Log.Debugf("'%s' is part of any group in ignore_usergroups: %s", currentUserName, strings.Join(rule.IgnoreUserGroups, ", "))
		return false
	}

	// if they didn't get denied at this point and no 'allow' rules are set, let them through
	if len(rule.AllowUsers)+len(rule.AllowUserGroups)+len(rule.AllowUserIds) == 0 {
		return true
	}

	// check if they are part of the allow users list
	for _, name := range rule.AllowUsers {
		if name == currentUserName {
			canRunRule = true
			break
		}
	}

	// check if they are part of the allow users ids list
	for _, userId := range rule.AllowUserIds {
		if userId == currentUserID {
			canRunRule = true
			break
		}
	}

	// if they still can't run the rule,
	// check if they are a member of any of the supplied allowed user groups
	if !canRunRule && len(rule.AllowUserGroups) > 0 {
		isAllowed, err := isMemberOfGroup(currentUserID, rule.AllowUserGroups, bot)
		// deny access if unable to check group membership due to error
		if err != nil {
			return false
		}
		canRunRule = isAllowed
	}

	if !canRunRule {
		if len(rule.AllowUsers) > 0 {
			bot.Log.Debugf("'%s' is not part of allow_users: %s", currentUserName, strings.Join(rule.AllowUsers, ", "))
		}

		if len(rule.AllowUserIds) > 0 {
			bot.Log.Debugf("'%s' is not part of allow_userids: %s", currentUserID, strings.Join(rule.AllowUserIds, ", "))
		}

		if  len(rule.AllowUserGroups) > 0 {
			bot.Log.Debugf("'%s' is not part of any groups in allow_usergroups: %s", currentUserName, strings.Join(rule.AllowUserGroups, ", "))
		}
	}

	return canRunRule
}

// utility function to check if a user is part of the specified user groups,
// if it's unable to check groupmembership, it will return an error
// TODO: Refactor to keep remote specific stuff in remote, also to allow increase testability
func isMemberOfGroup(currentUserID string, userGroups []string, bot *models.Bot) (bool, error) {
	if len(userGroups) == 0 {
		return false, nil
	}

	capp := strings.ToLower(bot.ChatApplication)
	switch capp {
	case "discord":
		bot.Log.Error("Discord is currently not supported for validating user permissions on rules")
		return false, nil
	case "slack":
		if bot.SlackWorkspaceToken == "" {
			bot.Log.Debugf("Limiting to usergroups only works if you register " +
				"your bot as an app with Slack and set the 'slack_workspace_token' property. " +
				"Restricting access to rule. Unset 'allow_usergroups' and/or 'ignore_usergroups', or set 'slack_workspace_token'.")
			return false, fmt.Errorf("slack_workspace_token not supplied - restricting access")
		}
		// Check if we are restricting by usergroup
		if bot.SlackWorkspaceToken != "" {
			wsAPI := slack.New(bot.SlackWorkspaceToken)
			for _, usergroupName := range userGroups {
				// Get the ID of the group from the usergroups the bot is aware of
				for knownUserGroupName, knownUserGroupID := range bot.UserGroups {
					if knownUserGroupName == usergroupName {
						// Get the members of the group
						userGroupMembers, err := wsAPI.GetUserGroupMembers(knownUserGroupID)
						if err != nil {
							bot.Log.Debugf("Unable to retrieve user group members, %s", err.Error())
						}
						// Check if any of the members are the current user
						for _, userGroupMemberID := range userGroupMembers {
							if userGroupMemberID == currentUserID {
								return true, nil
							}
						}
						break
					}
				}
			}
			wsAPI = nil
		}
		return false, nil
	default:
		bot.Log.Errorf("Chat application %s is not supported", capp)
		return false, nil
	}
}
