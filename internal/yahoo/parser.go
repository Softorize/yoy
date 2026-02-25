package yahoo

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"strings"
	"time"

	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
)

// ParseMessage parses a raw email message into our Message type.
func ParseMessage(r io.Reader) (*Message, error) {
	mr, err := mail.CreateReader(r)
	if err != nil {
		return nil, fmt.Errorf("creating mail reader: %w", err)
	}
	defer mr.Close()

	header := mr.Header
	msg := &Message{}

	// Parse From.
	if addrs, err := header.AddressList("From"); err == nil && len(addrs) > 0 {
		msg.From = Address{Name: addrs[0].Name, Address: addrs[0].Address}
	}

	// Parse To.
	if addrs, err := header.AddressList("To"); err == nil {
		msg.To = make([]Address, len(addrs))
		for i, a := range addrs {
			msg.To[i] = Address{Name: a.Name, Address: a.Address}
		}
	}

	// Parse Cc.
	if addrs, err := header.AddressList("Cc"); err == nil {
		msg.Cc = make([]Address, len(addrs))
		for i, a := range addrs {
			msg.Cc[i] = Address{Name: a.Name, Address: a.Address}
		}
	}

	// Parse Reply-To.
	if addrs, err := header.AddressList("Reply-To"); err == nil {
		msg.ReplyTo = make([]Address, len(addrs))
		for i, a := range addrs {
			msg.ReplyTo[i] = Address{Name: a.Name, Address: a.Address}
		}
	}

	// Parse Subject.
	if subject, err := header.Subject(); err == nil {
		msg.Subject = subject
	}

	// Parse Date.
	if date, err := header.Date(); err == nil {
		msg.Date = date
	}

	// Parse Message-ID.
	if id, err := header.MessageID(); err == nil {
		msg.MessageID = id
	}

	// Parse In-Reply-To.
	if v := header.Get("In-Reply-To"); v != "" {
		msg.InReplyTo = strings.TrimSpace(v)
	}

	// Parse References.
	if v := header.Get("References"); v != "" {
		refs := strings.Fields(v)
		msg.References = refs
	}

	// Parse body parts.
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		switch h := part.Header.(type) {
		case *mail.InlineHeader:
			ct, _, _ := h.ContentType()
			body, err := io.ReadAll(part.Body)
			if err != nil {
				continue
			}
			switch {
			case strings.HasPrefix(ct, "text/plain"):
				msg.Body = string(body)
			case strings.HasPrefix(ct, "text/html"):
				msg.HTMLBody = string(body)
			}
		case *mail.AttachmentHeader:
			filename, _ := h.Filename()
			ct, _, _ := h.ContentType()
			body, err := io.ReadAll(part.Body)
			if err != nil {
				continue
			}
			msg.Attachments = append(msg.Attachments, Attachment{
				Filename:    filename,
				ContentType: ct,
				Size:        len(body),
			})
		}
	}

	// If no plain text body, use HTML body.
	if msg.Body == "" && msg.HTMLBody != "" {
		msg.Body = msg.HTMLBody
	}

	return msg, nil
}

// ComposeMessage creates a raw email message from SendOptions.
func ComposeMessage(opts *SendOptions) ([]byte, error) {
	var buf bytes.Buffer

	var h mail.Header
	h.SetAddressList("From", []*mail.Address{{Address: opts.From}})

	toAddrs := make([]*mail.Address, len(opts.To))
	for i, addr := range opts.To {
		toAddrs[i] = &mail.Address{Address: addr}
	}
	h.SetAddressList("To", toAddrs)

	if len(opts.Cc) > 0 {
		ccAddrs := make([]*mail.Address, len(opts.Cc))
		for i, addr := range opts.Cc {
			ccAddrs[i] = &mail.Address{Address: addr}
		}
		h.SetAddressList("Cc", ccAddrs)
	}

	h.SetSubject(opts.Subject)
	h.SetDate(time.Now())
	h.SetContentType("text/plain", map[string]string{"charset": "UTF-8"})

	if opts.ReplyTo != "" {
		h.Set("In-Reply-To", opts.ReplyTo)
	}

	for k, v := range opts.Headers {
		h.Set(k, v)
	}

	mw, err := mail.CreateSingleInlineWriter(&buf, h)
	if err != nil {
		return nil, fmt.Errorf("creating mail writer: %w", err)
	}

	if _, err := io.WriteString(mw, opts.Body); err != nil {
		return nil, fmt.Errorf("writing body: %w", err)
	}

	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("closing mail writer: %w", err)
	}

	return buf.Bytes(), nil
}

// DecodeRFC2047 decodes RFC 2047 encoded words in a string.
func DecodeRFC2047(s string) string {
	dec := new(mime.WordDecoder)
	decoded, err := dec.DecodeHeader(s)
	if err != nil {
		return s
	}
	return decoded
}
