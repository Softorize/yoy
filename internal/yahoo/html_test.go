package yahoo

import (
	"strings"
	"testing"
)

func TestHTMLToText(t *testing.T) {
	in := `<html><head><style>.a{color:red}</style><title>ignore</title></head>` +
		`<body><p>Hello &amp; welcome</p><div>Line two</div>` +
		`<script>alert('x')</script><p>Third&nbsp;line</p></body></html>`
	got := HTMLToText(in)

	if strings.Contains(got, "color:red") || strings.Contains(got, "alert") || strings.Contains(got, "ignore") {
		t.Fatalf("style/script/title leaked into output: %q", got)
	}
	for _, want := range []string{"Hello & welcome", "Line two", "Third line"} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q missing %q", got, want)
		}
	}
	if strings.Contains(got, "\n\n\n") {
		t.Errorf("output has 3+ consecutive newlines: %q", got)
	}
}

func TestHTMLToTextBreaks(t *testing.T) {
	got := HTMLToText("a<br>b")
	if got != "a\nb" {
		t.Errorf("got %q, want %q", got, "a\nb")
	}
}

func TestDisplayBodyPrefersPlain(t *testing.T) {
	m := &Message{Body: "plain text", HTMLBody: "<p>html</p>"}
	if got := m.DisplayBody(); got != "plain text" {
		t.Errorf("DisplayBody = %q, want plain text", got)
	}
}

func TestDisplayBodyRendersHTMLFallback(t *testing.T) {
	// Parser sets Body == HTMLBody when there is no text/plain part.
	htmlBody := "<p>Only <b>HTML</b> here</p>"
	m := &Message{Body: htmlBody, HTMLBody: htmlBody}
	got := m.DisplayBody()
	if strings.Contains(got, "<p>") || strings.Contains(got, "<b>") {
		t.Errorf("DisplayBody left HTML tags: %q", got)
	}
	if !strings.Contains(got, "Only HTML here") {
		t.Errorf("DisplayBody = %q, want text content", got)
	}
}
