package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/spf13/viper"

	"github.com/target/flottbot/core"
	"github.com/target/flottbot/models"
	"github.com/target/flottbot/version"
)

func newBot() *models.Bot {
	bot := viper.New()
	bot.AddConfigPath("./config")
	bot.AddConfigPath(".")
	bot.SetConfigName("bot")
	err := bot.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %s \n", err)
	}

	var botC models.Bot
	err = bot.Unmarshal(&botC)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return &botC
}

func main() {
	var rules = make(map[string]models.Rule)
	var hitRule = make(chan models.Rule, 1)
	var inputMsgs = make(chan models.Message, 1)
	var outputMsgs = make(chan models.Message, 1)

	ver := flag.Bool("version", false, "print version information")
	v := flag.Bool("v", false, "print version information")

	flag.Parse()
	if *v || *ver {
		fmt.Println(version.String())
		os.Exit(0)
	}

	// Configure the bot to the core framework
	bot := newBot()
	core.Configure(bot)

	// Populate the global rules map
	core.Rules(&rules, bot)

	// Initialize and run Prometheus metrics logging
	go core.Prommetric("init", bot)

	// Create the wait group for handling concurrent runs (see further down)
	// Add 3 to the wait group so the three separate processes run concurrently
	// - process 1: core.Remotes - reads messages
	// - process 2: core.Matcher - processes messages
	// - process 3: core.Outpus - sends out messages
	var wg sync.WaitGroup
	wg.Add(3)

	go core.Remotes(inputMsgs, rules, bot)
	go core.Matcher(inputMsgs, outputMsgs, rules, hitRule, bot)
	go core.Outputs(outputMsgs, hitRule, bot)

	defer wg.Done()

	// This will run the bot indefinitely because the wait group will
	// attempt to wait for the above never-ending go routines.
	// Since said go routines run forever, they will never finish
	// and so this program will wait, or essentially run, forever until
	// terminated
	wg.Wait()
}
