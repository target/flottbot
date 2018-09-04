package handlers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/utils"
)

// ScriptExec handles 'exec' actions; script executions for rules
func ScriptExec(args models.Action, msg *models.Message, bot *models.Bot) (*models.ScriptResponse, error) {
	bot.Log.Debugf("Executing process for action '%s'", args.Name)
	// Default timeout of 20 seconds for any script execution, modifyable in rule yaml file
	if args.Timeout == 0 {
		args.Timeout = 20
	}

	// Prep default response
	result := &models.ScriptResponse{
		Status: 1, // Default is exit code 1 (error)
	}

	// Create context for executing command; will deal with timeouts
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(args.Timeout)*time.Second)
	defer cancel()

	// Deal with variable substitution in command
	cmdProcessed, err := utils.Substitute(args.Cmd, msg.Vars)
	if err != nil {
		return result, err
	}

	// Parse out all the arguments from the supplied command
	bin := utils.FindArgs(cmdProcessed)
	// Execute the command + arguments with the context
	cmd := exec.CommandContext(ctx, bin[0], bin[1:]...)

	// Capture stdout/stderr
	out, err := cmd.Output()

	// Handle timeouts
	if ctx.Err() == context.DeadlineExceeded {
		result.Output = "Hmm, something timed out. Please try again."
		return result, fmt.Errorf("Timeout reached, exec process for action '%s' cancelled", args.Name)
	}

	// Deal with non-zero exit codes
	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			ws := err.(*exec.ExitError).Sys().(syscall.WaitStatus)
			stderr := strings.Trim(string(err.(*exec.ExitError).Stderr), " \n")
			bot.Log.Debugf("Process for action '%s' exited with status %d: %s", args.Name, ws.ExitStatus(), stderr)
			result.Status = ws.ExitStatus()
			result.Output = stderr
		default:
			// this should rarely/never get hit
			bot.Log.Debugf("Couldn't get exit status for action '%s'", args.Name)
			result.Output = strings.Trim(err.Error(), " \n")
		}
		// if something was printed to stdout before the error, use that as output
		if string(out) != "" {
			result.Output = string(out)
		}
		return result, err
	}

	// should be exit code 0 here
	bot.Log.Debugf("Process finished for action '%s'", args.Name)
	ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
	result.Status = ws.ExitStatus()
	result.Output = strings.Trim(string(out), " \n")

	return result, nil
}
