package handlers

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/utils"
)

// ScriptExec handles 'exec' actions; script executions for rules
func ScriptExec(args models.Action, msg *models.Message, bot *models.Bot) (*models.ScriptResponse, error) {

	if args.Timeout == 0 {
		// Default timeout of 20 seconds for any script execution, modifyable in rule yaml file
		args.Timeout = 20
	}

	result := &models.ScriptResponse{
		Status: 1, // Default is exit code 1 (error)
	}

	cmdProcessed, err := utils.Substitute(args.Cmd, msg.Vars)
	if err != nil {
		return result, err
	}

	bin := utils.FindArgs(cmdProcessed)
	cmd := exec.Command(bin[0], bin[1:]...)

	var stdout, stderr, buf bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	done := make(chan error)

	go func() {
		if _, err := buf.ReadFrom(&stdout); err != nil {
			bot.Log.Errorf("Exec rule for action '%s' failed with %s: %s", args.Name, err, stderr.String())
		}
		done <- cmd.Run()
	}()

	// Catch timeouts and return error
	select {
	case <-time.After(time.Duration(args.Timeout) * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			return result, fmt.Errorf("Failed to kill exec process for action '%s': %s", args.Name, err.Error())
		}
		result.Output = "Hmm, something timed out. Please try again."
		return result, fmt.Errorf("Timeout reached, exec process for action '%s' killed", args.Name)
	case err := <-done:
		if err != nil {
			close(done)
			result.Output = stdout.String()
			// Assuming that the error contains an exit status code integer that can be parsed out.
			re := regexp.MustCompile("[0-9]+")
			match := re.FindStringSubmatch(err.Error())
			if len(match) > 0 {
				statuscode := re.FindAllString(err.Error(), -1)[0]
				status, err2 := strconv.Atoi(statuscode)
				if err2 != nil {
					return result, fmt.Errorf("Failed to parse error status code from Stderr")
				}
				result.Status = status
			} else {
				return result, fmt.Errorf("Did not find exit status code in Stderr, using default error status code of 1")
			}
			return result, fmt.Errorf("Exec rule for action '%s' failed with %s: %s", args.Name, err, stderr.String())
		}

		result.Status = 0
		result.Output = strings.Trim(stdout.String(), " \n")

		return result, nil
	}
}
