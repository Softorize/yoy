package auth

import (
	"path/filepath"

	"github.com/99designs/keyring"

	"github.com/Softorize/yoy/internal/config"
)

const keyringService = "yoy"

func openKeyring() (keyring.Keyring, error) {
	return keyring.Open(keyring.Config{
		ServiceName: keyringService,
		FileDir:     filepath.Join(config.Dir(), "keyring"),
		FilePasswordFunc: func(_ string) (string, error) {
			return "", nil
		},
	})
}
