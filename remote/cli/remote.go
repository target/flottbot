package cli

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/remote"
	"github.com/target/flottbot/version"
)

// Client struct
type Client struct {
}

// validate that Client adheres to remote interface
var _ remote.Remote = (*Client)(nil)

// Reaction implementation to satisfy remote interface
func (c *Client) Reaction(message models.Message, rule models.Rule, bot *models.Bot) {
	// not implemented for CLI
}

// Read implementation to satisfy remote interface
func (c *Client) Read(inputMsgs chan<- models.Message, rules map[string]models.Rule, bot *models.Bot) {
	user := bot.CLIUser
	if user == "" {
		user = "Flottbot-CLI-User"
	}
	fmt.Print(`
     ( )
.-----'-----.
| ( )   ( ) |  -( flottbot started )
'-----.-----' 
 / '+---+' \
 \/--|_|--\/` + "\n\n")
	fmt.Println(version.String())
	fmt.Print("Entering CLI mode. <Ctrl-C> to exit.\n\n")
	fmt.Print(user + "> ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		req := scanner.Text()
		if strings.TrimSpace(req) != "" {
			message := models.NewMessage()

			message.Type = models.MsgTypeDirect
			message.Service = models.MsgServiceCLI
			message.Input = req

			message.Vars["_user.id"] = user
			message.Vars["_user.firstname"] = user
			message.Vars["_user.name"] = user
			inputMsgs <- message
		} else {
			// nothing was entered. prevent blank line.
			fmt.Print(user + "> ")
		}
	}
	if err := scanner.Err(); err != nil {
		bot.Log.Debugf("Error reading standard input: %v", err)
	}
}

// Send implementation to satisfy remote interface
func (c *Client) Send(message models.Message, bot *models.Bot) {
	var re = regexp.MustCompile(`(?m)^(.*)`)
	var substitution = fmt.Sprintf(`%s> $1`, bot.Name)

	user := bot.CLIUser
	if user == "" {
		user = "Flottbot-CLI-User"
	}

	w := bufio.NewWriter(os.Stdout)
	fmt.Fprintln(w, re.ReplaceAllString(message.Output, substitution))

	// after sending the main message, also present a new prompt
	fmt.Fprint(w, user+"> ")
	w.Flush()
}

// InteractiveComponents implementation to satisfy remote interface
func (c *Client) InteractiveComponents(inputMsgs chan<- models.Message, message *models.Message, rule models.Rule, bot *models.Bot) {
	// not implemented for CLI
}
