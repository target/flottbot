package version

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

var (
	// Version supplies the semantic version
	Version string
	// GitHash supplies the git sha hash used
	GitHash string
)

// String prints the build information for the bot
func String() string {
	_, err := semver.NewVersion(Version)
	if err != nil {
		Version = "Dev build (no valid version)"
	}
	if GitHash == "" {
		GitHash = "N/A"
	}
	return fmt.Sprintf("Version : %s\nGit Hash: %s\n", Version, GitHash)
}
