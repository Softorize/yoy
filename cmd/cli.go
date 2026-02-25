package cmd

import (
	"context"
	"fmt"

	"github.com/Softorize/yoy/internal/config"
	"github.com/Softorize/yoy/internal/output"
	"github.com/Softorize/yoy/internal/yahoo"
)

// CLI is the root command structure for yoy.
type CLI struct {
	// Global flags
	Folder  string `help:"Mail folder." env:"YOY_FOLDER" short:"f" default:"INBOX"`
	JSON    bool   `help:"Output as JSON." env:"YOY_JSON"`
	Plain   bool   `help:"Output as plain TSV (no colors/borders)." env:"YOY_PLAIN"`
	Color   string `help:"Color mode: auto, always, never." default:"auto" enum:"auto,always,never"`
	Verbose bool   `help:"Enable verbose output." short:"v"`

	// Subcommands
	Auth       AuthCmd       `cmd:"" help:"Manage authentication."`
	Mail       MailCmd       `cmd:"" help:"Mail operations."`
	Folders    FoldersCmd    `cmd:"" help:"Manage mail folders."`
	Config     ConfigCmd     `cmd:"" help:"Manage configuration."`
	Version    VersionCmd    `cmd:"" help:"Print version information."`
	Completion CompletionCmd `cmd:"" help:"Generate shell completions."`

	// Aliases
	Send   SendCmd       `cmd:"" help:"Send an email (alias for mail send)." hidden:""`
	Ls     MailListCmd   `cmd:"" help:"List messages (alias for mail list)." hidden:""`
	Search MailSearchCmd `cmd:"" help:"Search messages (alias for mail search)." hidden:""`
}

// Context holds shared state for all commands.
type Context struct {
	Config  *config.Config
	JSON    bool
	Plain   bool
	Verbose bool
	Color   bool
	Folder  string

	ctx        context.Context
	imapClient *yahoo.IMAPClient
	formatter  output.Formatter
	email      string
}

// IMAPClient returns the IMAP client, creating it lazily.
func (c *Context) IMAPClient() (*yahoo.IMAPClient, error) {
	if c.imapClient != nil {
		return c.imapClient, nil
	}

	email, err := c.Email()
	if err != nil {
		return nil, err
	}

	client, err := yahoo.NewIMAPClient(c.ctx, email)
	if err != nil {
		return nil, err
	}
	c.imapClient = client
	return c.imapClient, nil
}

// Email returns the authenticated user's email address.
func (c *Context) Email() (string, error) {
	if c.email != "" {
		return c.email, nil
	}

	email, err := loadEmail()
	if err != nil {
		return "", fmt.Errorf("not authenticated: %w\nRun 'yoy auth login' to authenticate", err)
	}
	c.email = email
	return email, nil
}

// Formatter returns the output formatter.
func (c *Context) Formatter() output.Formatter {
	if c.formatter != nil {
		return c.formatter
	}
	format := "table"
	if c.JSON {
		format = "json"
	} else if c.Plain {
		format = "plain"
	}
	c.formatter = output.NewFormatter(format, c.Color)
	return c.formatter
}

// Close cleans up resources.
func (c *Context) Close() {
	if c.imapClient != nil {
		c.imapClient.Close()
	}
}

// NewContext creates a new command context.
func NewContext(cli *CLI, cfg *config.Config) *Context {
	colorEnabled := output.ShouldColorize(output.ParseColorMode(cli.Color))
	return &Context{
		Config:  cfg,
		JSON:    cli.JSON,
		Plain:   cli.Plain,
		Verbose: cli.Verbose,
		Color:   colorEnabled,
		Folder:  cli.Folder,
		ctx:     context.Background(),
	}
}
