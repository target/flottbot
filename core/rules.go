// Copyright (c) 2023 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package core

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/utils"
)

// Rules - searches the rules directory for any existing .yml rules
// and proceeds to create Rule objects for each .yml rule,
// and then finally populates a rules map with said Rule objects.
// The rules map is used to dictate the bots behavior and response patterns.
func Rules(rules *map[string]models.Rule, bot *models.Bot) {
	// Check if the rules directory even exists
	log.Debug().Msg("looking for rules directory...")

	currDir, err := os.Getwd()
	if err != nil {
		log.Error().Msg("can't get current working directory")
	}

	// TODO: make customizable
	rulesDir := path.Join(currDir, "config", "rules")

	_, err = os.Stat(rulesDir)
	if err != nil {
		log.Error().Msg("config/rules directory not found")
	}

	// Loop through the rules directory and create a list of rules
	log.Info().Msg("fetching all rule files...")

	fileList := []string{}

	err = filepath.Walk(rulesDir, func(path string, f os.FileInfo, err error) error {
		if f != nil && !f.IsDir() {
			fileList = append(fileList, path)
		}

		return nil
	})
	if err != nil {
		log.Error().Msgf("could not parse rules: %v", err)
	}

	// If the rules directory is empty, log a warning and exit
	if len(fileList) == 0 {
		log.Warn().Msg("looks like there aren't any rules")

		return
	}

	// Loop through the list of rules, creating a Rule object
	// for each rule, then populate the map of Rule objects
	log.Debug().Msgf("parsing %d rule files...", len(fileList))

	for _, ruleFile := range fileList {
		ruleConf := viper.New()
		ruleConf.SetConfigFile(ruleFile)

		err := ruleConf.ReadInConfig()
		if err != nil {
			log.Error().Msgf("error while reading rule file %#q: %v", ruleFile, err)
		}

		rule := models.Rule{}

		err = ruleConf.Unmarshal(&rule)
		if err != nil {
			log.Error().Msg(err.Error())
		}

		err = validateRule(&rule)
		if err != nil {
			log.Error().Msg(err.Error())
		}

		(*rules)[ruleFile] = rule
	}

	log.Info().Msgf("configured %#q rules!", bot.Name)
}

// Validate applies any environmental changes.
func validateRule(r *models.Rule) error {
	for i := range r.OutputToRooms {
		token, err := utils.Substitute(r.OutputToRooms[i], map[string]string{})
		if err != nil {
			return fmt.Errorf("could not configure output room: %w", err)
		}

		r.OutputToRooms[i] = token
	}

	return nil
}
