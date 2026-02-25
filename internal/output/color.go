package output

import (
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

// ANSI color codes.
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	White     = "\033[37m"
	BoldCyan  = "\033[1;36m"
	BoldGreen = "\033[1;32m"
	BoldRed   = "\033[1;31m"
)

// ColorMode determines when colors are used.
type ColorMode int

const (
	ColorAuto   ColorMode = iota
	ColorAlways
	ColorNever
)

// ParseColorMode converts a string to ColorMode.
func ParseColorMode(s string) ColorMode {
	switch strings.ToLower(s) {
	case "always":
		return ColorAlways
	case "never":
		return ColorNever
	default:
		return ColorAuto
	}
}

// ShouldColorize returns true if output should be colorized.
func ShouldColorize(mode ColorMode) bool {
	switch mode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	default:
		if _, ok := os.LookupEnv("NO_COLOR"); ok {
			return false
		}
		return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	}
}

// Colorize wraps text in ANSI color codes if colors are enabled.
func Colorize(text, color string, enabled bool) string {
	if !enabled {
		return text
	}
	return color + text + Reset
}
