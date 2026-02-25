package yahoo

import "time"

// Address represents an email address.
type Address struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address"`
}

// Attachment represents an email attachment.
type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int    `json:"size"`
}

// Message represents an email message.
type Message struct {
	UID         uint32       `json:"uid"`
	MessageID   string       `json:"message_id,omitempty"`
	Subject     string       `json:"subject"`
	From        Address      `json:"from"`
	To          []Address    `json:"to"`
	Cc          []Address    `json:"cc,omitempty"`
	ReplyTo     []Address    `json:"reply_to,omitempty"`
	Date        time.Time    `json:"date"`
	Body        string       `json:"body,omitempty"`
	HTMLBody    string       `json:"html_body,omitempty"`
	Flags       []string     `json:"flags,omitempty"`
	Seen        bool         `json:"seen"`
	Flagged     bool         `json:"flagged"`
	Attachments []Attachment `json:"attachments,omitempty"`
	InReplyTo   string       `json:"in_reply_to,omitempty"`
	References  []string     `json:"references,omitempty"`
}

// Folder represents a mail folder.
type Folder struct {
	Name     string `json:"name"`
	Messages uint32 `json:"messages"`
	Unseen   uint32 `json:"unseen"`
}

// SendOptions holds options for sending an email.
type SendOptions struct {
	From    string
	To      []string
	Cc      []string
	Bcc     []string
	Subject string
	Body    string
	ReplyTo string
	Headers map[string]string
}
