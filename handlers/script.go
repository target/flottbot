// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package handlers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/target/flottbot/models"
	"github.com/target/flottbot/utils"
)

// ScriptExec handles 'exec' actions; script executions for rules.
func ScriptExec(args models.Action, msg *models.Message) (*models.ScriptResponse, error) {
	log.Info().Msgf("executing process for action %#q", args.Name)
	// Default timeout of 20 seconds for any script execution, modifyable in rule file
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
	log.Debug().Msgf("command is: [%s]", args.Cmd)
	cmdProcessed, err := utils.Substitute(args.Cmd, msg.Vars)
	log.Debug().Msgf("substituted: [%s]", cmdProcessed)

	if err != nil {
		return result, err
	}

	// Parse out all the arguments from the supplied command
	bin := utils.ExecArgTokenizer(cmdProcessed)

	// prep the command to be executed with context
	//nolint:gosec // ignore "potential tainted input or cmd arguments" because bot owner controls usage
	cmd := exec.CommandContext(ctx, bin[0], bin[1:]...)

	// run command and capture stdout/stderr
	out, err := cmd.CombinedOutput()
	if err != nil {
		// handle timeouts
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			result.Output = "Hmm, the command timed out. Please try again."

			return result, fmt.Errorf("timeout reached, exec process for action %#q canceled", args.Name)
		}

		// check if file couldn't be found
		errFileNotFoundMsg := "file not found: %s"
		if os.IsNotExist(err) {
			result.Output = fmt.Sprintf(errFileNotFoundMsg, args.Cmd)
		}

		// check other variations that might
		// indicate that a file couldn't be found
		// whether called directly,
		// or indirectly
		outAsLower := strings.ToLower(string(out))
		if strings.Contains(outAsLower, "no such file") ||
			strings.Contains(outAsLower, "can't open") ||
			strings.Contains(err.Error(), "file not found") {
			result.Output = fmt.Sprintf("file not found: %s", args.Cmd)
		}

		// grab the statuscode of the process
		exitCode := cmd.ProcessState.ExitCode()

		// this might be -1 if process didn't exit
		// or was terminated otherwise
		//
		// only pass on "real" error codes
		if exitCode > 0 {
			result.Status = exitCode
		}

		// if we had an error not covered above,
		// attempt to grab the output
		if result.Output == "" {
			result.Output = strings.Trim(string(out), " \n")
		}

		return result, err
	}

	// should be exit code 0 here
	log.Info().Msgf("process finished for action %#q", args.Name)

	// ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
	result.Status = cmd.ProcessState.ExitCode()
	result.Output = strings.Trim(string(out), " \n")

	return result, nil
}
