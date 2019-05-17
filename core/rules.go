package core

import (
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
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
	bot.Log.Debug("Looking for rules directory...")
	searchDir, err := utils.PathExists(path.Join("config", "rules"))
	if err != nil {
		bot.Log.Fatalf("Could not parse rules: %v", err)
	}

	// Loop through the rules directory and create a list of rules
	bot.Log.Debug("Fetching all rule files...")
	fileList := []string{}
	err = filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		bot.Log.Fatalf("Could not parse rules: %v", err)
	}

	// If the rules directory is empty, log a warning and exit the function
	if len(fileList) == 0 {
		bot.Log.Warn("Looks like there aren't any rules")
		return
	}

	// Loop through the list of rules, creating a Rule object
	// for each rule, then populate the map of Rule objects
	bot.Log.Debug("Reading and parsing rule files...")
	for _, ruleFile := range fileList {
		ruleConf := viper.New()
		ruleConf.SetConfigFile(ruleFile)
		err := ruleConf.ReadInConfig()
		if err != nil {
			bot.Log.Errorf("Error while reading rule file '%s': %s \n", ruleFile, err)
		}

		rule := models.Rule{}
		err = ruleConf.Unmarshal(&rule)
		if err != nil {
			log.Fatalf(err.Error())
		}
		err = rule.Validate()
		if err != nil {
			log.Fatalf(err.Error())
		}
		(*rules)[ruleFile] = rule
	}

	bot.Log.Infof("Configured '%s' rules!", bot.Name)
}
