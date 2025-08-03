// SPDX-License-Identifier: Apache-2.0

package version

import (
	"fmt"
	"runtime/debug"

	"github.com/Masterminds/semver/v3"
)

// Version supplies the semantic version.
var Version string

// String prints the build information for the bot.
func String() string {
	hash := "unknown"

	_, err := semver.NewVersion(Version)
	if err != nil {
		Version = "dev"
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				hash = s.Value
			}
		}
	}

	return fmt.Sprintf("Version : %s\nGit Hash: %s\n", Version, hash)
}
