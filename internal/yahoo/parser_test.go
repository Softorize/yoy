package yahoo

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseMessagePlain(t *testing.T) {
	raw := "From: Alice Example <alice@example.com>\r\n" +
		"To: bob@yahoo.com\r\n" +
		"Subject: Hello there\r\n" +
		"Date: Mon, 02 Jan 2006 15:04:05 -0700\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		"This is the body.\r\n"

	msg, err := ParseMessage(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
	if msg.From.Address != "alice@example.com" {
		t.Errorf("From = %q, want alice@example.com", msg.From.Address)
	}
	if msg.From.Name != "Alice Example" {
		t.Errorf("From name = %q, want Alice Example", msg.From.Name)
	}
	if len(msg.To) != 1 || msg.To[0].Address != "bob@yahoo.com" {
		t.Errorf("To = %+v, want [bob@yahoo.com]", msg.To)
	}
	if msg.Subject != "Hello there" {
		t.Errorf("Subject = %q", msg.Subject)
	}
	if !strings.Contains(msg.Body, "This is the body.") {
		t.Errorf("Body = %q", msg.Body)
	}
}

func TestParseMessageMultipartHTML(t *testing.T) {
	raw := "From: a@example.com\r\n" +
		"To: b@yahoo.com\r\n" +
		"Subject: Multipart\r\n" +
		"Content-Type: multipart/alternative; boundary=\"BOUND\"\r\n" +
		"\r\n" +
		"--BOUND\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		"plain version\r\n" +
		"--BOUND\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		"<p>html version</p>\r\n" +
		"--BOUND--\r\n"

	msg, err := ParseMessage(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
	if !strings.Contains(msg.Body, "plain version") {
		t.Errorf("Body = %q, want plain version", msg.Body)
	}
	if !strings.Contains(msg.HTMLBody, "html version") {
		t.Errorf("HTMLBody = %q, want html version", msg.HTMLBody)
	}
}

func TestParseMessageAttachment(t *testing.T) {
	raw := "From: a@example.com\r\n" +
		"To: b@yahoo.com\r\n" +
		"Subject: With attachment\r\n" +
		"Content-Type: multipart/mixed; boundary=\"BOUND\"\r\n" +
		"\r\n" +
		"--BOUND\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		"see attached\r\n" +
		"--BOUND\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"Content-Disposition: attachment; filename=\"notes.txt\"\r\n" +
		"\r\n" +
		"hello\r\n" +
		"--BOUND--\r\n"

	msg, err := ParseMessage(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
	if len(msg.Attachments) != 1 {
		t.Fatalf("Attachments = %d, want 1", len(msg.Attachments))
	}
	if msg.Attachments[0].Filename != "notes.txt" {
		t.Errorf("Filename = %q, want notes.txt", msg.Attachments[0].Filename)
	}
	if msg.Attachments[0].Size != len("hello") {
		t.Errorf("Size = %d, want %d", msg.Attachments[0].Size, len("hello"))
	}
}

func TestParseMessageFallsBackToHTMLBody(t *testing.T) {
	raw := "From: a@example.com\r\n" +
		"To: b@yahoo.com\r\n" +
		"Subject: HTML only\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		"<p>only html</p>\r\n"

	msg, err := ParseMessage(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
	if !strings.Contains(msg.Body, "only html") {
		t.Errorf("Body = %q, want it to fall back to HTML", msg.Body)
	}
}

func TestComposeMessageRoundTrip(t *testing.T) {
	opts := &SendOptions{
		From:    "me@yahoo.com",
		To:      []string{"you@example.com"},
		Cc:      []string{"cc@example.com"},
		Subject: "Round trip",
		Body:    "hello world",
	}
	b, err := ComposeMessage(opts)
	if err != nil {
		t.Fatalf("ComposeMessage: %v", err)
	}

	msg, err := ParseMessage(strings.NewReader(string(b)))
	if err != nil {
		t.Fatalf("ParseMessage(composed): %v", err)
	}
	if msg.From.Address != "me@yahoo.com" {
		t.Errorf("From = %q", msg.From.Address)
	}
	if len(msg.To) != 1 || msg.To[0].Address != "you@example.com" {
		t.Errorf("To = %+v", msg.To)
	}
	if len(msg.Cc) != 1 || msg.Cc[0].Address != "cc@example.com" {
		t.Errorf("Cc = %+v", msg.Cc)
	}
	if msg.Subject != "Round trip" {
		t.Errorf("Subject = %q", msg.Subject)
	}
	if !strings.Contains(msg.Body, "hello world") {
		t.Errorf("Body = %q", msg.Body)
	}
}

func TestDecodeRFC2047(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"=?UTF-8?B?SGVsbG8gV29ybGQ=?=", "Hello World"},
		{"=?UTF-8?Q?Caf=C3=A9?=", "Café"},
		{"plain text passes through", "plain text passes through"},
		{"=?ISO-8859-1?Q?=E9t=E9?=", "été"},
	}
	for _, c := range cases {
		if got := DecodeRFC2047(c.in); got != c.want {
			t.Errorf("DecodeRFC2047(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestComposeMessageWithAttachmentRoundTrips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Claim 123 - Summons (2 pages).pdf")
	payload := bytes.Repeat([]byte{0x25, 0x50, 0x44, 0x46, 0x00, 0xff}, 5000)
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}

	raw, err := ComposeMessage(&SendOptions{
		From: "a@yahoo.com", To: []string{"b@example.com"}, Subject: "Docs",
		Body: "see attached", Attachments: []string{path},
	})
	if err != nil {
		t.Fatalf("ComposeMessage: %v", err)
	}

	msg, err := ParseMessage(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("ParseMessage: %v", err)
	}
	if !strings.Contains(msg.Body, "see attached") {
		t.Errorf("Body = %q, want it to contain the text", msg.Body)
	}
	if len(msg.Attachments) != 1 {
		t.Fatalf("Attachments = %d, want 1", len(msg.Attachments))
	}
	a := msg.Attachments[0]
	if a.Filename != filepath.Base(path) {
		t.Errorf("Filename = %q, want %q", a.Filename, filepath.Base(path))
	}
	if a.ContentType != "application/pdf" {
		t.Errorf("ContentType = %q, want application/pdf", a.ContentType)
	}
	if a.Size != len(payload) {
		t.Errorf("Size = %d, want %d", a.Size, len(payload))
	}
}

func TestComposeMessageWithoutAttachmentsStaysSinglePart(t *testing.T) {
	raw, err := ComposeMessage(&SendOptions{From: "a@yahoo.com", To: []string{"b@example.com"}, Subject: "Hi", Body: "hello"})
	if err != nil {
		t.Fatalf("ComposeMessage: %v", err)
	}
	if bytes.Contains(raw, []byte("multipart/")) {
		t.Errorf("message without attachments should stay text/plain, got multipart")
	}
}
