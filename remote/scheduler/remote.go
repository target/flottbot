package scheduler

import (
	"fmt"
	"strings"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/target/flottbot/models"
	"github.com/target/flottbot/remote"
)

// Client struct
type Client struct {
}

// validate that Client adheres to remote interface
var _ remote.Remote = (*Client)(nil)

// Reaction implementation to satisfy remote interface
func (c *Client) Reaction(message models.Message, rule models.Rule, bot *models.Bot) {
	// not implemented for Scheduler
}

// Read implementation to satisfy remote interface
// This will read in schedule type rules from the rules map and create cronjobs that will
// trigger messages to be sent for processing to the Matcher function via 'inputMsgs' channel.
func (c *Client) Read(inputMsgs chan<- models.Message, rules map[string]models.Rule, bot *models.Bot) {
	// Wait for bot.Rooms to populate (find a less hacky way to do this)
	for {
		_nil := bot.Rooms[""]
		if len(bot.Rooms) > 0 {
			bot.Log.Debugf("scheduler connected to %s channels: %s", strings.Title(bot.ChatApplication), _nil)
			break
		}
	}

	var job *cron.Cron
	// Create a list of cron jobs to execute
	jobs := []*cron.Cron{}

	// Find and create schedules
	for _, rule := range rules {
		if rule.Active && rule.Schedule != "" {
			// Pre-checks before executing rule as a cron job
			if len(rule.OutputToRooms) == 0 && len(rule.OutputToUsers) == 0 {
				bot.Log.Debug("scheduling rules require the 'output_to_rooms' and/or 'output_to_users' fields to be set")
				continue
			} else if len(rule.OutputToRooms) > 0 && len(bot.Rooms) == 0 {
				bot.Log.Debugf("unable to connect scheduler to these rooms: %s", rule.OutputToRooms)
				continue
			} else if rule.Respond != "" || rule.Hear != "" {
				bot.Log.Debug("scheduling rules does not allow the 'respond' and 'hear' fields")
				continue
			}

			bot.Log.Debugf("scheduler is adding rule '%s'", rule.Name)

			scheduleName := rule.Name
			input := fmt.Sprintf("<@%s> ", bot.ID) // send message as self
			outputRooms := rule.OutputToRooms
			outputUsers := rule.OutputToUsers

			// prepare the job function
			jobFunc := func() {
				bot.Log.Debugf("executing scheduler for rule '%s'", scheduleName)
				// build the message
				message := models.NewMessage()
				message.Service = models.MsgServiceScheduler
				message.Input = input // send message as self
				message.Attributes["from_schedule"] = scheduleName
				message.Type = models.MsgTypeChannel
				message.OutputToRooms = outputRooms
				message.OutputToUsers = outputUsers
				inputMsgs <- message
			}

			// use our logger for cron
			cronLogger := cron.PrintfLogger(&bot.Log)

			// check if the provided schedule is of standard format, ie. 5 fields
			_, err := cron.ParseStandard(rule.Schedule)
			if err == nil {
				// standard cron
				job = cron.New(cron.WithChain(cron.SkipIfStillRunning(cronLogger)))
			} else {
				// (probably?) quartz cron
				job = cron.New(cron.WithSeconds(), cron.WithChain(cron.SkipIfStillRunning(cronLogger)))
			}

			// try to create new cron job
			_, err = job.AddFunc(rule.Schedule, jobFunc)
			if err != nil {
				// typically the error is due to incorrect cron format
				bot.Log.Errorf("unable to add schedule for rule '%s': verify that the supplied schedule is supported", rule.Name)
				// more verbose log. note: will probably convey that spec
				// needs to be 6 fields, although any supported format will work.
				bot.Log.Debugf("error while adding job: %v", err)
				continue
			}

			jobs = append(jobs, job)
		}
	}

	if len(jobs) == 0 {
		bot.Log.Warn("no schedules were added - please check for errors")
		return
	}

	processJobs(jobs, bot)
}

// Send implementation to satisfy remote interface
func (c *Client) Send(message models.Message, bot *models.Bot) {
	// not implemented for Scheduler
}

// InteractiveComponents implementation to satisfy remote interface
func (c *Client) InteractiveComponents(inputMsgs chan<- models.Message, message *models.Message, rule models.Rule, bot *models.Bot) {
	// not implemented for Scheduler
}

// Process the Cron jobs
func processJobs(jobs []*cron.Cron, bot *models.Bot) {
	// Create wait group for cron jobs and execute them
	wg := &sync.WaitGroup{}
	wg.Add(len(jobs))
	for _, job := range jobs {
		go func(c *cron.Cron) {
			c.Start()
		}(job)
		defer job.Stop()
	}
	wg.Wait()
	bot.Log.Warn("scheduler is closing")
}
