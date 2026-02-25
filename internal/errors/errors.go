package errors

import (
	"errors"
	"fmt"
	"strings"
)

// Exit codes for the CLI.
const (
	ExitOK            = 0
	ExitGeneral       = 1
	ExitAuth          = 2
	ExitNotFound      = 3
	ExitPermission    = 4
	ExitIMAPError     = 5
	ExitInvalidInput  = 6
	ExitNetwork       = 7
	ExitConfig        = 8
	ExitNotConfigured = 9
	ExitSMTPError     = 10
)

// YoyError is the standard error type for yoy.
type YoyError struct {
	Message  string
	Err      error
	ExitCode int
	Hint     string
}

func (e *YoyError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *YoyError) Unwrap() error {
	return e.Err
}

// New creates a new YoyError.
func New(msg string, exitCode int) *YoyError {
	return &YoyError{Message: msg, ExitCode: exitCode}
}

// Wrap wraps an existing error into a YoyError.
func Wrap(msg string, err error, exitCode int) *YoyError {
	return &YoyError{Message: msg, Err: err, ExitCode: exitCode}
}

// WithHint adds a user-facing hint to the error.
func (e *YoyError) WithHint(hint string) *YoyError {
	e.Hint = hint
	return e
}

// FromIMAPError converts an IMAP error into a YoyError with appropriate hints.
func FromIMAPError(err error) *YoyError {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	switch {
	case strings.Contains(lower, "authenticate") || strings.Contains(lower, "login"):
		return &YoyError{
			Message:  "Authentication failed",
			Err:      err,
			ExitCode: ExitAuth,
			Hint:     "Run 'yoy auth login' to re-authenticate.",
		}
	case strings.Contains(lower, "no such mailbox") || strings.Contains(lower, "doesn't exist"):
		return &YoyError{
			Message:  "Folder not found",
			Err:      err,
			ExitCode: ExitNotFound,
			Hint:     "Use 'yoy folders list' to see available folders.",
		}
	case strings.Contains(lower, "permission") || strings.Contains(lower, "denied"):
		return &YoyError{
			Message:  "Permission denied",
			Err:      err,
			ExitCode: ExitPermission,
		}
	case strings.Contains(lower, "connection") || strings.Contains(lower, "timeout") || strings.Contains(lower, "eof"):
		return &YoyError{
			Message:  "Connection failed",
			Err:      err,
			ExitCode: ExitNetwork,
			Hint:     "Check your internet connection and try again.",
		}
	default:
		return &YoyError{
			Message:  "IMAP error",
			Err:      err,
			ExitCode: ExitIMAPError,
		}
	}
}

// FromSMTPError converts an SMTP error into a YoyError.
func FromSMTPError(err error) *YoyError {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	switch {
	case strings.Contains(lower, "auth") || strings.Contains(lower, "credential"):
		return &YoyError{
			Message:  "SMTP authentication failed",
			Err:      err,
			ExitCode: ExitAuth,
			Hint:     "Run 'yoy auth login' to re-authenticate.",
		}
	case strings.Contains(lower, "connection") || strings.Contains(lower, "timeout"):
		return &YoyError{
			Message:  "SMTP connection failed",
			Err:      err,
			ExitCode: ExitNetwork,
			Hint:     "Check your internet connection and try again.",
		}
	default:
		return &YoyError{
			Message:  "SMTP error",
			Err:      err,
			ExitCode: ExitSMTPError,
		}
	}
}

// ExitCodeFrom extracts the exit code from an error.
func ExitCodeFrom(err error) int {
	var yoyErr *YoyError
	if errors.As(err, &yoyErr) {
		return yoyErr.ExitCode
	}
	return ExitGeneral
}

// HintFrom extracts the hint from an error, if any.
func HintFrom(err error) string {
	var yoyErr *YoyError
	if errors.As(err, &yoyErr) {
		return yoyErr.Hint
	}
	return ""
}
