package errors

import (
	"errors"
	"testing"
)

func TestFromIMAPError(t *testing.T) {
	cases := []struct {
		name     string
		in       error
		wantCode int
	}{
		{"auth", errors.New("AUTHENTICATE failed"), ExitAuth},
		{"login", errors.New("LOGIN rejected"), ExitAuth},
		{"mailbox", errors.New("no such mailbox Foo"), ExitNotFound},
		{"permission", errors.New("permission denied"), ExitPermission},
		{"connection", errors.New("connection reset"), ExitNetwork},
		{"eof", errors.New("unexpected EOF"), ExitNetwork},
		{"other", errors.New("something weird"), ExitIMAPError},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := FromIMAPError(c.in)
			if got.ExitCode != c.wantCode {
				t.Errorf("ExitCode = %d, want %d", got.ExitCode, c.wantCode)
			}
		})
	}
	if FromIMAPError(nil) != nil {
		t.Error("FromIMAPError(nil) should be nil")
	}
}

func TestFromSMTPError(t *testing.T) {
	cases := []struct {
		name     string
		in       error
		wantCode int
	}{
		{"auth", errors.New("auth credentials invalid"), ExitAuth},
		{"connection", errors.New("connection timeout"), ExitNetwork},
		{"other", errors.New("550 rejected"), ExitSMTPError},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := FromSMTPError(c.in)
			if got.ExitCode != c.wantCode {
				t.Errorf("ExitCode = %d, want %d", got.ExitCode, c.wantCode)
			}
		})
	}
	if FromSMTPError(nil) != nil {
		t.Error("FromSMTPError(nil) should be nil")
	}
}

func TestExitCodeFrom(t *testing.T) {
	if code := ExitCodeFrom(New("boom", ExitNotFound)); code != ExitNotFound {
		t.Errorf("ExitCodeFrom(YoyError) = %d, want %d", code, ExitNotFound)
	}
	if code := ExitCodeFrom(errors.New("plain")); code != ExitGeneral {
		t.Errorf("ExitCodeFrom(plain) = %d, want %d", code, ExitGeneral)
	}
}

func TestHintFrom(t *testing.T) {
	err := New("boom", ExitAuth).WithHint("do the thing")
	if h := HintFrom(err); h != "do the thing" {
		t.Errorf("HintFrom = %q, want %q", h, "do the thing")
	}
	if h := HintFrom(errors.New("plain")); h != "" {
		t.Errorf("HintFrom(plain) = %q, want empty", h)
	}
}

func TestYoyErrorUnwrap(t *testing.T) {
	inner := errors.New("inner")
	err := Wrap("outer", inner, ExitGeneral)
	if !errors.Is(err, inner) {
		t.Error("errors.Is should find the wrapped error")
	}
}
