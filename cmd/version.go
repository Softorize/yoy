package cmd

import (
	"fmt"

	"github.com/Softorize/yoy/internal/version"
)

// VersionCmd prints version information.
type VersionCmd struct{}

// Run prints the version string.
func (c *VersionCmd) Run(ctx *Context) error {
	fmt.Println(version.Full())
	return nil
}
