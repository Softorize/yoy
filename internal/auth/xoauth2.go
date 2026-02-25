package auth

import (
	"fmt"

	"github.com/emersion/go-sasl"
)

// XOAUTH2 mechanism name.
const XOAUTH2 = "XOAUTH2"

// xoauth2Client implements the XOAUTH2 SASL mechanism.
type xoauth2Client struct {
	email       string
	accessToken string
}

func (c *xoauth2Client) Start() (string, []byte, error) {
	resp := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", c.email, c.accessToken)
	return XOAUTH2, []byte(resp), nil
}

func (c *xoauth2Client) Next(challenge []byte) ([]byte, error) {
	return nil, fmt.Errorf("XOAUTH2 unexpected challenge: %s", string(challenge))
}

// NewXOAuth2Client creates a SASL XOAUTH2 client for IMAP/SMTP authentication.
func NewXOAuth2Client(email, accessToken string) sasl.Client {
	return &xoauth2Client{email: email, accessToken: accessToken}
}
