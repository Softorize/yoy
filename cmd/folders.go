package cmd

import (
	"fmt"
	"os"
)

// FoldersCmd groups folder management subcommands.
type FoldersCmd struct {
	List   FoldersListCmd   `cmd:"" help:"List all folders."`
	Create FoldersCreateCmd `cmd:"" help:"Create a new folder."`
	Delete FoldersDeleteCmd `cmd:"" help:"Delete a folder."`
}

// FoldersListCmd lists all folders.
type FoldersListCmd struct{}

// Run lists folders.
func (c *FoldersListCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	folders, err := client.ListFolders()
	if err != nil {
		return err
	}

	return ctx.Formatter().FormatFolders(os.Stdout, folders)
}

// FoldersCreateCmd creates a new folder.
type FoldersCreateCmd struct {
	Name string `arg:"" help:"Folder name to create."`
}

// Run creates a folder.
func (c *FoldersCreateCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	if err := client.CreateFolder(c.Name); err != nil {
		return err
	}

	fmt.Printf("Folder %q created.\n", c.Name)
	return nil
}

// FoldersDeleteCmd deletes a folder.
type FoldersDeleteCmd struct {
	Name string `arg:"" help:"Folder name to delete."`
}

// Run deletes a folder.
func (c *FoldersDeleteCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	if err := client.DeleteFolder(c.Name); err != nil {
		return err
	}

	fmt.Printf("Folder %q deleted.\n", c.Name)
	return nil
}
