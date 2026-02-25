package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/99designs/keyring"
	"golang.org/x/oauth2"

	"github.com/Softorize/yoy/internal/config"
)

const keyringService = "yoy"
const tokenKey = "oauth_token"

func openKeyring() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		ServiceName: keyringService,
		FileDir:     filepath.Join(config.Dir(), "keyring"),
		FilePasswordFunc: func(_ string) (string, error) {
			return "", nil
		},
	})
}

// StoreToken saves the OAuth2 token to the system keyring, falling back to a file.
func StoreToken(token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("marshaling token: %w", err)
	}

	kr, err := openKeyring()
	if err == nil {
		err = kr.Set(keyring.Item{
			Key:  tokenKey,
			Data: data,
		})
		if err == nil {
			return nil
		}
	}

	return storeTokenFile(data)
}

// LoadToken loads the OAuth2 token from the system keyring, falling back to a file.
func LoadToken() (*oauth2.Token, error) {
	kr, err := openKeyring()
	if err == nil {
		item, err := kr.Get(tokenKey)
		if err == nil {
			var token oauth2.Token
			if err := json.Unmarshal(item.Data, &token); err != nil {
				return nil, fmt.Errorf("unmarshaling token from keyring: %w", err)
			}
			return &token, nil
		}
	}

	return loadTokenFile()
}

// RemoveToken removes the OAuth2 token from the keyring and file storage.
func RemoveToken() error {
	kr, err := openKeyring()
	if err == nil {
		_ = kr.Remove(tokenKey)
	}

	path := tokenFilePath()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing token file: %w", err)
	}
	return nil
}

// GetTokenSource creates a reusing token source from the stored token that auto-refreshes.
func GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	token, err := LoadToken()
	if err != nil {
		return nil, fmt.Errorf("loading token: %w", err)
	}

	cfg, err := OAuthConfig(config.DefaultOAuthPort)
	if err != nil {
		return nil, err
	}
	ts := cfg.TokenSource(ctx, token)

	return oauth2.ReuseTokenSource(token, &savingTokenSource{
		base: ts,
	}), nil
}

// savingTokenSource wraps a token source and saves new tokens when refreshed.
type savingTokenSource struct {
	base oauth2.TokenSource
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.base.Token()
	if err != nil {
		return nil, err
	}
	// Save refreshed token.
	_ = StoreToken(token)
	return token, nil
}

// GetAccessToken returns the current access token string.
func GetAccessToken(ctx context.Context) (string, error) {
	ts, err := GetTokenSource(ctx)
	if err != nil {
		return "", err
	}
	token, err := ts.Token()
	if err != nil {
		return "", fmt.Errorf("getting access token: %w", err)
	}
	return token.AccessToken, nil
}

func tokenFilePath() string {
	return filepath.Join(config.TokenDir(), "default.json")
}

func storeTokenFile(data []byte) error {
	dir := config.TokenDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating token dir: %w", err)
	}
	if err := os.WriteFile(tokenFilePath(), data, 0600); err != nil {
		return fmt.Errorf("writing token file: %w", err)
	}
	return nil
}

func loadTokenFile() (*oauth2.Token, error) {
	data, err := os.ReadFile(tokenFilePath())
	if err != nil {
		return nil, fmt.Errorf("reading token file: %w", err)
	}
	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("unmarshaling token from file: %w", err)
	}
	return &token, nil
}
