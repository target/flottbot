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
	bot.Log.Debug().Msg("looking for rules directory...")
	searchDir, err := utils.PathExists(path.Join("config", "rules"))
	if err != nil {
		bot.Log.Fatal().Msgf("could not parse rules: %v", err)
	}

	// Loop through the rules directory and create a list of rules
	bot.Log.Info().Msg("fetching all rule files...")
	fileList := []string{}
	err = filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		bot.Log.Fatal().Msgf("could not parse rules: %v", err)
	}

	// If the rules directory is empty, log a warning and exit the function
	if len(fileList) == 0 {
		bot.Log.Warn().Msg("looks like there aren't any rules")
		return
	}

	// Loop through the list of rules, creating a Rule object
	// for each rule, then populate the map of Rule objects
	bot.Log.Debug().Msg("reading and parsing rule files...")
	for _, ruleFile := range fileList {
		ruleConf := viper.New()
		ruleConf.SetConfigFile(ruleFile)
		err := ruleConf.ReadInConfig()
		if err != nil {
			bot.Log.Error().Msgf("error while reading rule file '%s': %v", ruleFile, err)
		}

		rule := models.Rule{}
		err = ruleConf.Unmarshal(&rule)
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		err = validateRule(bot, &rule)
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		(*rules)[ruleFile] = rule
	}

	bot.Log.Info().Msgf("configured '%s' rules!", bot.Name)
}

// Validate applies any environmental changes
func validateRule(bot *models.Bot, r *models.Rule) error {

	for i := range r.OutputToRooms {
		token, err := utils.Substitute(r.OutputToRooms[i], map[string]string{})
		if err != nil {
			return fmt.Errorf("could not configure output room: %s", err.Error())
		}

		r.OutputToRooms[i] = token
	}
	return nil
}
