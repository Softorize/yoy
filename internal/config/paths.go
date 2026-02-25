package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const appName = "yoy"

// Dir returns the platform-specific configuration directory.
func Dir() string {
	if v := os.Getenv("YOY_CONFIG_DIR"); v != "" {
		return v
	}

	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", appName)
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), appName)
	default: // linux and others
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, appName)
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", appName)
	}
}

// FilePath returns the full path to the config file.
func FilePath() string {
	return filepath.Join(Dir(), "config.yaml")
}

// TokenDir returns the directory for token storage (file fallback).
func TokenDir() string {
	return filepath.Join(Dir(), "tokens")
}
