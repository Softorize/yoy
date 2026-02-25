package cmd

import (
	"fmt"

	"github.com/Softorize/yoy/internal/yahoo"
)

// SendCmd sends a new email.
type SendCmd struct {
	To      []string `help:"Recipient email addresses." required:"" sep:","`
	Subject string   `help:"Email subject." required:""`
	Body    string   `help:"Email body text." required:""`
	Cc      []string `help:"CC recipients." sep:","`
	Bcc     []string `help:"BCC recipients." sep:","`
}

// Run sends the email.
func (c *SendCmd) Run(ctx *Context) error {
	email, err := ctx.Email()
	if err != nil {
		return err
	}

	opts := &yahoo.SendOptions{
		From:    email,
		To:      c.To,
		Cc:      c.Cc,
		Bcc:     c.Bcc,
		Subject: c.Subject,
		Body:    c.Body,
	}

	if err := yahoo.SendMail(ctx.ctx, email, opts); err != nil {
		return err
	}

	fmt.Println("Email sent successfully.")
	return nil
}
