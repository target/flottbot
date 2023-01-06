// Copyright (c) 2023 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package utils

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"

	"github.com/target/flottbot/models"
)

// CanTrigger ensures the user is allowed to use the respective rule.
func CanTrigger(currentUserName string, currentUserID string, rule models.Rule, bot *models.Bot) bool {
	var canRunRule bool

	// no restriction were given for this rule, allow to proceed
	if len(rule.AllowUsers)+len(rule.AllowUserGroups)+len(rule.AllowUserIds)+len(rule.IgnoreUsers)+len(rule.IgnoreUserGroups) == 0 {
		return true
	}

	// are they ignored directly? deny
	for _, name := range rule.IgnoreUsers {
		if name == currentUserName {
			log.Info().Msgf("%#q is on the 'ignore_users' list for rule: %#q", currentUserName, rule.Name)
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
		log.Info().
			Msgf("%#q is part of a group in ignore_usergroups: %#q", currentUserName, strings.Join(rule.IgnoreUserGroups, ", "))
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
	for _, userID := range rule.AllowUserIds {
		if userID == currentUserID {
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
			log.Info().
				Msgf("%#q is not part of allow_users: %#q", currentUserName, strings.Join(rule.AllowUsers, ", "))
		}

		if len(rule.AllowUserIds) > 0 {
			log.Info().
				Msgf("%#q is not part of allow_userids: %#q", currentUserID, strings.Join(rule.AllowUserIds, ", "))
		}

		if len(rule.AllowUserGroups) > 0 {
			log.Info().
				Msgf("%#q is not part of any groups in allow_usergroups: %#q", currentUserName, strings.Join(rule.AllowUserGroups, ", "))
		}
	}

	return canRunRule
}

// utility function to check if a user is part of the specified user groups,
// if it's unable to check groupmembership, it will return an error
// TODO: Refactor to keep remote specific stuff in remote, also to allow increase testability.
func isMemberOfGroup(currentUserID string, userGroups []string, bot *models.Bot) (bool, error) {
	if len(userGroups) == 0 {
		return false, nil
	}

	capp := strings.ToLower(bot.ChatApplication)
	switch capp {
	case "discord":
		var usr *discordgo.Member

		dg, err := discordgo.New("Bot " + bot.DiscordToken)
		if err != nil {
			return false, err
		}

		usr, err = dg.GuildMember(bot.DiscordServerID, currentUserID)
		if err != nil {
			log.Error().Msgf("error while searching for user - error: %v", err)
			return false, nil
		}

		for _, group := range userGroups {
			for _, uGroup := range usr.Roles {
				if strings.EqualFold(bot.UserGroups[group], uGroup) {
					return true, nil
				}
			}
		}

		return false, nil
	case "slack":
		// Check if we are restricting by usergroup
		api := slack.New(bot.SlackToken)

		for _, usergroupName := range userGroups {
			// Get the ID of the group from the usergroups the bot is aware of
			for knownUserGroupName, knownUserGroupID := range bot.UserGroups {
				if knownUserGroupName == usergroupName {
					// Get the members of the group
					userGroupMembers, err := api.GetUserGroupMembers(knownUserGroupID)
					if err != nil {
						log.Error().Msgf("unable to retrieve user group members, %v", err)
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

		return false, nil
	default:
		log.Error().Msgf("chat application %#q is not supported", capp)
		return false, nil
	}
}
