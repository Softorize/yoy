package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Softorize/yoy/internal/yahoo"
)

// MailCmd groups all mail subcommands.
type MailCmd struct {
	List       MailListCmd       `cmd:"" help:"List messages in a folder."`
	Search     MailSearchCmd     `cmd:"" help:"Search messages."`
	Read       MailReadCmd       `cmd:"" help:"Read a message."`
	Send       SendCmd           `cmd:"" help:"Send a new email."`
	Reply      MailReplyCmd      `cmd:"" help:"Reply to a message."`
	Forward    MailForwardCmd    `cmd:"" help:"Forward a message."`
	Delete     MailDeleteCmd     `cmd:"" help:"Delete a message."`
	Move       MailMoveCmd       `cmd:"" help:"Move a message to another folder."`
	Star       MailStarCmd       `cmd:"" help:"Star a message."`
	Unstar     MailUnstarCmd     `cmd:"" help:"Unstar a message."`
	MarkRead   MailMarkReadCmd   `cmd:"" help:"Mark a message as read."`
	MarkUnread MailMarkUnreadCmd `cmd:"" help:"Mark a message as unread."`
}

// MailListCmd lists messages in a folder.
type MailListCmd struct {
	Limit uint32 `help:"Number of messages to show." short:"n" default:"25"`
}

// Run lists messages.
func (c *MailListCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	limit := c.Limit
	if limit == 0 {
		limit = uint32(ctx.Config.MailLimit)
	}

	messages, err := client.ListMessages(ctx.Folder, limit)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		fmt.Println("No messages found.")
		return nil
	}

	return ctx.Formatter().FormatMessages(os.Stdout, messages)
}

// MailSearchCmd searches for messages.
type MailSearchCmd struct {
	Query string `arg:"" help:"Search query."`
}

// Run searches messages.
func (c *MailSearchCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	messages, err := client.SearchMessages(ctx.Folder, c.Query)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		fmt.Println("No messages found.")
		return nil
	}

	return ctx.Formatter().FormatMessages(os.Stdout, messages)
}

// MailReadCmd reads a message by UID.
type MailReadCmd struct {
	UID uint32 `arg:"" help:"Message UID."`
}

// Run reads a message.
func (c *MailReadCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	message, err := client.ReadMessage(ctx.Folder, c.UID)
	if err != nil {
		return err
	}

	return ctx.Formatter().FormatMessage(os.Stdout, message)
}

// MailDeleteCmd deletes one or more messages.
type MailDeleteCmd struct {
	UIDs []uint32 `arg:"" name:"uid" help:"Message UID(s) to delete."`
}

// Run deletes messages.
func (c *MailDeleteCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	if err := client.DeleteMessages(ctx.Folder, c.UIDs); err != nil {
		return err
	}

	fmt.Printf("Deleted %d message(s).\n", len(c.UIDs))
	return nil
}

// MailMoveCmd moves one or more messages to another folder.
type MailMoveCmd struct {
	UIDs []uint32 `arg:"" name:"uid" help:"Message UID(s) to move."`
	To   string   `help:"Destination folder." required:""`
}

// Run moves messages.
func (c *MailMoveCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	if err := client.MoveMessages(ctx.Folder, c.UIDs, c.To); err != nil {
		return err
	}

	fmt.Printf("Moved %d message(s) to %s.\n", len(c.UIDs), c.To)
	return nil
}

// MailStarCmd stars a message.
type MailStarCmd struct {
	UIDs []uint32 `arg:"" name:"uid" help:"Message UID(s)."`
}

// Run stars messages.
func (c *MailStarCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	if err := client.StarMessages(ctx.Folder, c.UIDs); err != nil {
		return err
	}

	fmt.Printf("Starred %d message(s).\n", len(c.UIDs))
	return nil
}

// MailUnstarCmd unstars a message.
type MailUnstarCmd struct {
	UIDs []uint32 `arg:"" name:"uid" help:"Message UID(s)."`
}

// Run unstars messages.
func (c *MailUnstarCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	if err := client.UnstarMessages(ctx.Folder, c.UIDs); err != nil {
		return err
	}

	fmt.Printf("Unstarred %d message(s).\n", len(c.UIDs))
	return nil
}

// MailMarkReadCmd marks a message as read.
type MailMarkReadCmd struct {
	UIDs []uint32 `arg:"" name:"uid" help:"Message UID(s)."`
}

// Run marks messages as read.
func (c *MailMarkReadCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	if err := client.MarkReadMulti(ctx.Folder, c.UIDs); err != nil {
		return err
	}

	fmt.Printf("Marked %d message(s) as read.\n", len(c.UIDs))
	return nil
}

// MailMarkUnreadCmd marks a message as unread.
type MailMarkUnreadCmd struct {
	UIDs []uint32 `arg:"" name:"uid" help:"Message UID(s)."`
}

// Run marks messages as unread.
func (c *MailMarkUnreadCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	if err := client.MarkUnreadMulti(ctx.Folder, c.UIDs); err != nil {
		return err
	}

	fmt.Printf("Marked %d message(s) as unread.\n", len(c.UIDs))
	return nil
}

// MailReplyCmd replies to a message.
type MailReplyCmd struct {
	UID  uint32 `arg:"" help:"Message UID to reply to."`
	Body string `help:"Reply body text." required:""`
	All  bool   `help:"Reply to all recipients." default:"false"`
}

// Run replies to a message.
func (c *MailReplyCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	// Fetch the original message.
	original, err := client.ReadMessage(ctx.Folder, c.UID)
	if err != nil {
		return err
	}

	email, err := ctx.Email()
	if err != nil {
		return err
	}

	// Build reply.
	subject := original.Subject
	if !strings.HasPrefix(strings.ToLower(subject), "re:") {
		subject = "Re: " + subject
	}

	to := []string{original.From.Address}
	if c.All {
		for _, addr := range original.To {
			if addr.Address != email {
				to = append(to, addr.Address)
			}
		}
		for _, addr := range original.Cc {
			if addr.Address != email {
				to = append(to, addr.Address)
			}
		}
	}

	headers := map[string]string{}
	if original.MessageID != "" {
		headers["In-Reply-To"] = original.MessageID
		headers["References"] = original.MessageID
	}

	opts := &yahoo.SendOptions{
		From:    email,
		To:      to,
		Subject: subject,
		Body:    c.Body,
		Headers: headers,
	}

	if err := yahoo.SendMail(email, opts); err != nil {
		return err
	}

	fmt.Println("Reply sent.")

	// Mark original as read.
	_ = client.MarkRead(ctx.Folder, c.UID)

	return nil
}

// MailForwardCmd forwards a message.
type MailForwardCmd struct {
	UID  uint32   `arg:"" help:"Message UID to forward."`
	To   []string `help:"Recipient email addresses." required:"" sep:","`
	Body string   `help:"Additional message body." default:""`
}

// Run forwards a message.
func (c *MailForwardCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	// Fetch the original message.
	original, err := client.ReadMessage(ctx.Folder, c.UID)
	if err != nil {
		return err
	}

	email, err := ctx.Email()
	if err != nil {
		return err
	}

	subject := original.Subject
	if len(subject) < 5 || subject[:5] != "Fwd: " {
		subject = "Fwd: " + subject
	}

	body := c.Body
	if body != "" {
		body += "\n\n"
	}
	body += "---------- Forwarded message ----------\n"
	body += fmt.Sprintf("From: %s <%s>\n", original.From.Name, original.From.Address)
	body += fmt.Sprintf("Date: %s\n", original.Date.Format("2006-01-02 15:04"))
	body += fmt.Sprintf("Subject: %s\n\n", original.Subject)
	body += original.Body

	opts := &yahoo.SendOptions{
		From:    email,
		To:      c.To,
		Subject: subject,
		Body:    body,
	}

	if err := yahoo.SendMail(email, opts); err != nil {
		return err
	}

	fmt.Println("Message forwarded.")
	return nil
}
