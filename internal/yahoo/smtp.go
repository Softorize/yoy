package yahoo

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"

	"github.com/emersion/go-sasl"

	"github.com/Softorize/yoy/internal/auth"
	"github.com/Softorize/yoy/internal/config"
	yoyerrors "github.com/Softorize/yoy/internal/errors"
)

// SendMail sends an email via Yahoo's SMTP server using XOAUTH2 authentication.
func SendMail(ctx context.Context, email string, opts *SendOptions) error {
	accessToken, err := auth.GetAccessToken(ctx)
	if err != nil {
		return yoyerrors.Wrap("getting access token", err, yoyerrors.ExitAuth).
			WithHint("Run 'yoy auth login' to authenticate.")
	}

	// Compose the message.
	if opts.From == "" {
		opts.From = email
	}
	msgBytes, err := ComposeMessage(opts)
	if err != nil {
		return yoyerrors.Wrap("composing message", err, yoyerrors.ExitGeneral)
	}

	// Connect via TLS (port 465).
	addr := fmt.Sprintf("%s:%d", config.DefaultSMTPHost, config.DefaultSMTPPort)
	tlsConfig := &tls.Config{ServerName: config.DefaultSMTPHost}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return yoyerrors.Wrap("connecting to SMTP server", err, yoyerrors.ExitNetwork).
			WithHint("Check your internet connection and try again.")
	}

	host, _, _ := net.SplitHostPort(addr)
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return yoyerrors.Wrap("creating SMTP client", err, yoyerrors.ExitSMTPError)
	}
	defer c.Close()

	// Authenticate with XOAUTH2.
	xoauth2Client := auth.NewXOAuth2Client(email, accessToken)
	if err := c.Auth(saslToSMTPAuth(xoauth2Client)); err != nil {
		return yoyerrors.Wrap("SMTP authentication failed", err, yoyerrors.ExitAuth).
			WithHint("Run 'yoy auth login' to re-authenticate.")
	}

	// Set sender.
	if err := c.Mail(opts.From); err != nil {
		return yoyerrors.FromSMTPError(fmt.Errorf("MAIL FROM: %w", err))
	}

	// Set recipients.
	allRecipients := make([]string, 0, len(opts.To)+len(opts.Cc)+len(opts.Bcc))
	allRecipients = append(allRecipients, opts.To...)
	allRecipients = append(allRecipients, opts.Cc...)
	allRecipients = append(allRecipients, opts.Bcc...)

	for _, rcpt := range allRecipients {
		if err := c.Rcpt(rcpt); err != nil {
			return yoyerrors.FromSMTPError(fmt.Errorf("RCPT TO %s: %w", rcpt, err))
		}
	}

	// Send body.
	wc, err := c.Data()
	if err != nil {
		return yoyerrors.FromSMTPError(fmt.Errorf("DATA: %w", err))
	}

	if _, err := wc.Write(msgBytes); err != nil {
		wc.Close()
		return yoyerrors.FromSMTPError(fmt.Errorf("writing message: %w", err))
	}

	if err := wc.Close(); err != nil {
		return yoyerrors.FromSMTPError(fmt.Errorf("closing data: %w", err))
	}

	return c.Quit()
}

// saslToSMTPAuth adapts a go-sasl Client to net/smtp.Auth.
type saslSMTPAuth struct {
	client sasl.Client
}

func saslToSMTPAuth(client sasl.Client) smtp.Auth {
	return &saslSMTPAuth{client: client}
}

func (a *saslSMTPAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return a.client.Start()
}

func (a *saslSMTPAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	return a.client.Next(fromServer)
}
