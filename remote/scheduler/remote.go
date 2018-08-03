package scheduler

import (
	"fmt"
	"strings"
	"sync"

	"github.com/robfig/cron"
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
			bot.Log.Debugf("Scheduler connected to %s channels%s", strings.Title(bot.ChatApplication), _nil)
			break
		}
	}
	// Create a list of cron jobs to execute
	jobs := []*cron.Cron{}

	// Find and create schedules
	for _, rule := range rules {
		if rule.Active && len(rule.Schedule) > 0 {
			// Pre-checks before executing rule as a cron job
			if len(rule.OutputToRooms) == 0 && len(rule.OutputToUsers) == 0 {
				bot.Log.Debug("Scheduling rules requires the 'output_to_rooms' and/or 'output_to_users' fields to be set")
				continue
			} else if len(rule.OutputToRooms) > 0 && len(bot.Rooms) == 0 {
				bot.Log.Debugf("Could not connect Scheduler to rooms: %s", rule.OutputToRooms)
				continue
			} else if len(rule.Respond) > 0 || len(rule.Hear) > 0 {
				bot.Log.Debug("Scheduling rules does not allow the 'respond' and 'hear' fields")
				continue
			}

			// TODO - Regex check for correct cron syntax

			bot.Log.Debugf("Scheduler is running rule '%s'", rule.Name)
			cron := cron.New()
			scheduleName := rule.Name
			input := fmt.Sprintf("<@%s> ", bot.ID) // send message as self
			outputRooms := rule.OutputToRooms
			outputUsers := rule.OutputToUsers
			cron.AddFunc(rule.Schedule, func() {
				// Build message
				message := models.NewMessage()
				message.Service = models.MsgServiceScheduler
				message.Input = input // send message as self
				message.Attributes["from_schedule"] = scheduleName
				message.Type = models.MsgTypeChannel
				message.OutputToRooms = outputRooms
				message.OutputToUsers = outputUsers
				inputMsgs <- message
			})
			jobs = append(jobs, cron)
		}
	}

	if len(jobs) == 0 {
		bot.Log.Warn("Found no schedule-type rules. Scheduler is closing")
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
	bot.Log.Warn("Scheduler is closing")
}
