package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Softorize/yoy/internal/auth"
	"github.com/Softorize/yoy/internal/config"
	"github.com/Softorize/yoy/internal/yahoo"
)

// AuthCmd groups all authentication subcommands.
type AuthCmd struct {
	Login  AuthLoginCmd  `cmd:"" help:"Authenticate with Yahoo (OAuth2 or app password)."`
	Logout AuthLogoutCmd `cmd:"" help:"Remove stored credentials."`
	Status AuthStatusCmd `cmd:"" help:"Show current authentication status."`
}

// AuthLoginCmd performs authentication via OAuth2 or app password.
type AuthLoginCmd struct {
	Port        int    `help:"Local port for OAuth callback." default:"8086"`
	Email       string `help:"Yahoo email address." required:""`
	AppPassword string `help:"Yahoo app password (skip OAuth, use IMAP/SMTP login)." name:"app-password"`
}

// Run executes the login flow.
func (c *AuthLoginCmd) Run(ctx *Context) error {
	if c.AppPassword != "" {
		return c.runAppPassword(ctx)
	}
	return c.runOAuth(ctx)
}

func (c *AuthLoginCmd) runAppPassword(ctx *Context) error {
	if err := auth.StoreAppPassword(c.AppPassword); err != nil {
		return fmt.Errorf("storing app password: %w", err)
	}

	if err := storeEmail(c.Email); err != nil {
		return fmt.Errorf("storing email: %w", err)
	}

	fmt.Println("App password stored. Verifying IMAP connection...")

	// Quick verification
	_, err := yahoo.NewIMAPClient(ctx.ctx, c.Email)
	if err != nil {
		// Remove stored credentials on failure
		auth.RemoveAppPassword()
		return fmt.Errorf("IMAP verification failed: %w\nCheck your app password and try again", err)
	}

	fmt.Println("Authentication successful.")
	return nil
}

func (c *AuthLoginCmd) runOAuth(ctx *Context) error {
	fmt.Println("Opening browser for Yahoo authentication...")

	token, err := auth.BrowserLogin(c.Port)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if err := auth.StoreToken(token); err != nil {
		return fmt.Errorf("storing token: %w", err)
	}

	if err := storeEmail(c.Email); err != nil {
		return fmt.Errorf("storing email: %w", err)
	}

	fmt.Println("Authentication successful.")
	return nil
}

// AuthLogoutCmd removes stored credentials.
type AuthLogoutCmd struct{}

// Run removes stored tokens.
func (c *AuthLogoutCmd) Run(ctx *Context) error {
	auth.RemoveAppPassword()
	if err := auth.RemoveToken(); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}
	_ = removeEmail()
	fmt.Println("Logged out successfully.")
	return nil
}

// AuthStatusCmd displays the current authentication status.
type AuthStatusCmd struct{}

// Run checks and displays the auth status.
func (c *AuthStatusCmd) Run(ctx *Context) error {
	creds, err := auth.LoadCredentials()
	if err != nil {
		fmt.Println("Status: not authenticated")
		fmt.Println("Run 'yoy auth login --email your@yahoo.com' to authenticate.")
		return nil
	}

	email, _ := loadEmail()

	fmt.Println("Status: authenticated")
	if email != "" {
		fmt.Printf("Email:  %s\n", email)
	}

	if creds.Method == auth.AuthMethodAppPassword {
		fmt.Println("Method: app password")
	} else {
		fmt.Println("Method: OAuth2")
		token, err := auth.LoadToken()
		if err == nil {
			if token.Expiry.IsZero() {
				fmt.Println("Expiry: none (non-expiring token)")
			} else if token.Expiry.Before(time.Now()) {
				fmt.Printf("Expiry: %s (expired, will auto-refresh)\n", token.Expiry.Format(time.RFC3339))
			} else {
				fmt.Printf("Expiry: %s\n", token.Expiry.Format(time.RFC3339))
			}
			fmt.Printf("Type:   %s\n", token.TokenType)
		}
	}

	return nil
}

// emailFilePath returns the path to the stored email file.
func emailFilePath() string {
	return filepath.Join(config.Dir(), "email.json")
}

// storeEmail saves the user's email address.
func storeEmail(email string) error {
	dir := config.Dir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, _ := json.Marshal(map[string]string{"email": email})
	return os.WriteFile(emailFilePath(), data, 0600)
}

// loadEmail loads the stored email address.
func loadEmail() (string, error) {
	data, err := os.ReadFile(emailFilePath())
	if err != nil {
		return "", err
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return "", err
	}
	return m["email"], nil
}

// removeEmail removes the stored email.
func removeEmail() error {
	return os.Remove(emailFilePath())
}
