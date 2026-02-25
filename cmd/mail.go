package cmd

import (
	"fmt"
	"os"

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

// MailDeleteCmd deletes a message.
type MailDeleteCmd struct {
	UID uint32 `arg:"" help:"Message UID."`
}

// Run deletes a message.
func (c *MailDeleteCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	if err := client.DeleteMessage(ctx.Folder, c.UID); err != nil {
		return err
	}

	fmt.Printf("Message %d deleted.\n", c.UID)
	return nil
}

// MailMoveCmd moves a message to another folder.
type MailMoveCmd struct {
	UID        uint32 `arg:"" help:"Message UID."`
	DestFolder string `arg:"" help:"Destination folder."`
}

// Run moves a message.
func (c *MailMoveCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	if err := client.MoveMessage(ctx.Folder, c.UID, c.DestFolder); err != nil {
		return err
	}

	fmt.Printf("Message %d moved to %s.\n", c.UID, c.DestFolder)
	return nil
}

// MailStarCmd stars a message.
type MailStarCmd struct {
	UID uint32 `arg:"" help:"Message UID."`
}

// Run stars a message.
func (c *MailStarCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	if err := client.StarMessage(ctx.Folder, c.UID); err != nil {
		return err
	}

	fmt.Printf("Message %d starred.\n", c.UID)
	return nil
}

// MailUnstarCmd unstars a message.
type MailUnstarCmd struct {
	UID uint32 `arg:"" help:"Message UID."`
}

// Run unstars a message.
func (c *MailUnstarCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	if err := client.UnstarMessage(ctx.Folder, c.UID); err != nil {
		return err
	}

	fmt.Printf("Message %d unstarred.\n", c.UID)
	return nil
}

// MailMarkReadCmd marks a message as read.
type MailMarkReadCmd struct {
	UID uint32 `arg:"" help:"Message UID."`
}

// Run marks a message as read.
func (c *MailMarkReadCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	if err := client.MarkRead(ctx.Folder, c.UID); err != nil {
		return err
	}

	fmt.Printf("Message %d marked as read.\n", c.UID)
	return nil
}

// MailMarkUnreadCmd marks a message as unread.
type MailMarkUnreadCmd struct {
	UID uint32 `arg:"" help:"Message UID."`
}

// Run marks a message as unread.
func (c *MailMarkUnreadCmd) Run(ctx *Context) error {
	client, err := ctx.IMAPClient()
	if err != nil {
		return err
	}

	if err := client.MarkUnread(ctx.Folder, c.UID); err != nil {
		return err
	}

	fmt.Printf("Message %d marked as unread.\n", c.UID)
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
	if len(subject) < 4 || subject[:4] != "Re: " {
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

	if err := yahoo.SendMail(ctx.ctx, email, opts); err != nil {
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

	if err := yahoo.SendMail(ctx.ctx, email, opts); err != nil {
		return err
	}

	fmt.Println("Message forwarded.")
	return nil
}
