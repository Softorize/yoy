package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/Softorize/yoy/internal/config"
)

// ConfigCmd groups configuration management subcommands.
type ConfigCmd struct {
	Get  ConfigGetCmd  `cmd:"" help:"Get a config value."`
	Set  ConfigSetCmd  `cmd:"" help:"Set a config value."`
	List ConfigListCmd `cmd:"" help:"List all config values."`
	Path ConfigPathCmd `cmd:"" help:"Print config file path."`
}

// ConfigGetCmd prints a single config value.
type ConfigGetCmd struct {
	Key string `arg:"" help:"Config key (e.g., default_folder)."`
}

// Run prints the value for the given key.
func (c *ConfigGetCmd) Run(ctx *Context) error {
	value, err := ctx.Config.Get(c.Key)
	if err != nil {
		return err
	}
	fmt.Println(value)
	return nil
}

// ConfigSetCmd updates a config value.
type ConfigSetCmd struct {
	Key   string `arg:"" help:"Config key."`
	Value string `arg:"" help:"Config value."`
}

// Run sets the config value and saves to disk.
func (c *ConfigSetCmd) Run(ctx *Context) error {
	if err := ctx.Config.Set(c.Key, c.Value); err != nil {
		return err
	}
	if err := ctx.Config.Save(); err != nil {
		return err
	}
	fmt.Printf("Set %s = %s\n", c.Key, c.Value)
	return nil
}

// ConfigListCmd lists all config values.
type ConfigListCmd struct{}

// Run prints all config key-value pairs.
func (c *ConfigListCmd) Run(ctx *Context) error {
	values := ctx.Config.List()

	if ctx.JSON {
		return ctx.Formatter().FormatKeyValue(os.Stdout, values)
	}

	// Sort keys for consistent output.
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	data := make(map[string]string, len(keys))
	for _, k := range keys {
		data[k] = values[k]
	}
	return ctx.Formatter().FormatKeyValue(os.Stdout, data)
}

// ConfigPathCmd prints the config file path.
type ConfigPathCmd struct{}

// Run prints the config file location.
func (c *ConfigPathCmd) Run(ctx *Context) error {
	fmt.Println(config.FilePath())
	return nil
}
