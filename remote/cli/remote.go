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
	if len(user) == 0 {
		user = "Flottbot-CLI-User"
	}
	fmt.Println(`MMMMMMMMMMMMMMMMMMMMMMMWNNWMMMMMMMMMMMMMMMMMMMMMMM
MMMMMMMMMMMMMMMMMMMMNkl;;;;lONMMMMMMMMMMMMMMMMMMMM
MMMMMMMMMMMMMMMMMMMNo.   .  .dNMMMMMMMMMMMMMMMMMMM
MMMMMMMMMMMMMMMMMMMK:       .cXMMMMMMMMMMMMMMMMMMM
MMMMMMMMMMMMMMMMMMMWk,.    .;OWMMMMMMMMMMMMMMMMMMM
MMMMMMMMMMMMMMMMMMMMWKc.  .lXMMMMMMMMMMMMMMMMMMMMM
MMMMXkdooooooooooooooo;.  .;ooooooooooooooodkXMMMM      ______      __  __  __          __
MMMK:.                                      .cXMMM     / __/ /___  / /_/ /_/ /_  ____  / /_
MMMO'                                        ,0MMM    / /_/ / __ \/ __/ __/ __ \/ __ \/ __/
MMMO'      .;lodl;.           ..;ldol,.      ,0MMM   / __/ / /_/ / /_/ /_/ /_/ / /_/ / /_
MMMO'    .,kNMMMMNk;.        .;ONMMMMNx,.    ,0MMM  /_/ /_/\____/\__/\__/_.___/\____/\__/
MMMO'    .xWMMMMMMWx.        'kMMMMMMMWd.    ,0MMM
MMMO'    .oNMMMMMMWd.        .xWMMMMMMNl.    ,0MMM
MMMO'     .l0NWWN0l.          .o0NWWNOc.     ,0MMM            __             __           __
MMMO'      ..,;;,..            ..,;;,.       ,0MMM      _____/ /_____ ______/ /____  ____/ /
MMMO'                                        ,0MMM    / ___/ __/ __  / ___/ __/ _ \/ __  /
MMMXl................        ................oXMMM   (__  ) /_/ /_/ / /  / /_/  __/ /_/ /
MMMMWKkkkdc:::::cdkkxl..  ..cxkkdc:::::cdkkOKWMMMM  /____/\__/\__,_/_/   \__/\___/\__,_/
MMMMMMMMMK:......c0WMW0occo0WMW0c......cXMMMMMMMMM
MMMMMMMMMW0:.,'. .'cx0XNNNNX0xc.. .',':0MMMMMMMMMM
MMMMMMMMMMMNXNKd'.  ..',,,,'..  .,dXNXNMMMMMMMMMMM
MMMMMMMMMMMMMMMNc.    ......    .lNMMMMMMMMMMMMMMM
MMMMMMMMMMMMMMMNkc,...lkkkkl...,ckNMMMMMMMMMMMMMMM
MMMMMMMMMMMMMMMMMWN0kONMMMMNOOKNWMMMMMMMMMMMMMMMMM
MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM`)
	fmt.Println(version.String())
	fmt.Println("Enter CLI mode: hit <Enter>. <Ctrl-C> to exit.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Print("\n", bot.Name, "> ")
		req := scanner.Text()
		if len(strings.TrimSpace(req)) > 0 {
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
		bot.Log.Debugln(os.Stderr, "reading standard input:", err)
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
