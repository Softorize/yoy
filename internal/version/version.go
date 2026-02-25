package version

import "fmt"

// These variables are set at build time via ldflags.
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// Full returns a formatted version string.
func Full() string {
	return fmt.Sprintf("yoy %s (commit: %s, built: %s)", Version, Commit, BuildDate)
}

// Short returns just the version number.
func Short() string {
	return Version
}
