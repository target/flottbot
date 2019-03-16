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
	fmt.Println(`
     ( )
.-----'-----.
| ( )   ( ) |  -( flottbot started )
'-----.-----' 
 / '+---+' \
 \/--|_|--\/
`)
	fmt.Println(version.String())
	fmt.Println("Enter CLI mode: hit <Enter>. <Ctrl-C> to exit.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Print("\n", bot.Name, "> ")
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
		}
	}
	if err := scanner.Err(); err != nil {
		bot.Log.Debugf("Error reading standard input: %v", err)
	}
}

// Send implementation to satisfy remote interface
func (c *Client) Send(message models.Message, bot *models.Bot) {
	w := bufio.NewWriter(os.Stdout)
	var re = regexp.MustCompile(`(?m)^(.*)`)
	var substitution = fmt.Sprintf(`%s> $1`, bot.Name)
	fmt.Fprintln(w, re.ReplaceAllString(message.Output, substitution))
	w.Flush()
}

// InteractiveComponents implementation to satisfy remote interface
func (c *Client) InteractiveComponents(inputMsgs chan<- models.Message, message *models.Message, rule models.Rule, bot *models.Bot) {
	// not implemented for CLI
}
