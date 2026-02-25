package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/99designs/keyring"

	"github.com/Softorize/yoy/internal/config"
)

const appPasswordKey = "app_password"

// StoredCredentials holds the current auth configuration.
type StoredCredentials struct {
	AppPassword string `json:"app_password,omitempty"`
}

// StoreAppPassword saves an app password for IMAP/SMTP plain auth.
func StoreAppPassword(password string) error {
	data, _ := json.Marshal(StoredCredentials{
		AppPassword: password,
	})

	kr, err := openKeyring()
	if err == nil {
		err = kr.Set(keyring.Item{Key: appPasswordKey, Data: data})
		if err == nil {
			return nil
		}
	}

	// File fallback
	dir := config.TokenDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating credential dir: %w", err)
	}
	return os.WriteFile(credFilePath(), data, 0600)
}

// LoadCredentials loads the stored credentials.
func LoadCredentials() (*StoredCredentials, error) {
	kr, err := openKeyring()
	if err == nil {
		item, err := kr.Get(appPasswordKey)
		if err == nil {
			var creds StoredCredentials
			if err := json.Unmarshal(item.Data, &creds); err == nil {
				return &creds, nil
			}
		}
	}

	// Try file fallback
	data, err := os.ReadFile(credFilePath())
	if err == nil {
		var creds StoredCredentials
		if err := json.Unmarshal(data, &creds); err == nil {
			return &creds, nil
		}
	}

	return nil, fmt.Errorf("no credentials stored")
}

// RemoveAppPassword removes stored app password credentials.
func RemoveAppPassword() error {
	kr, err := openKeyring()
	if err == nil {
		_ = kr.Remove(appPasswordKey)
	}
	_ = os.Remove(credFilePath())
	return nil
}

func credFilePath() string {
	return filepath.Join(config.TokenDir(), "credentials.json")
}
