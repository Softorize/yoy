package yahoo

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"

	"github.com/Softorize/yoy/internal/auth"
	"github.com/Softorize/yoy/internal/config"
	yoyerrors "github.com/Softorize/yoy/internal/errors"
)

// IMAPClient wraps the IMAP connection for Yahoo Mail.
type IMAPClient struct {
	client *imapclient.Client
	email  string
}

// NewIMAPClient creates and authenticates a new IMAP connection to Yahoo Mail.
func NewIMAPClient(ctx context.Context, email string) (*IMAPClient, error) {
	creds, err := auth.LoadCredentials()
	if err != nil {
		return nil, yoyerrors.Wrap("loading credentials", err, yoyerrors.ExitAuth).
			WithHint("Run 'yoy auth login' to authenticate.")
	}

	addr := fmt.Sprintf("%s:%d", config.DefaultIMAPHost, config.DefaultIMAPPort)
	c, err := imapclient.DialTLS(addr, nil)
	if err != nil {
		return nil, yoyerrors.Wrap("connecting to IMAP server", err, yoyerrors.ExitNetwork).
			WithHint("Check your internet connection and try again.")
	}

	switch creds.Method {
	case auth.AuthMethodAppPassword:
		if err := c.Login(email, creds.AppPassword).Wait(); err != nil {
			c.Close()
			return nil, yoyerrors.Wrap("IMAP login failed", err, yoyerrors.ExitAuth).
				WithHint("Check your app password or generate a new one at https://login.yahoo.com/account/security")
		}
	default:
		accessToken, err := auth.GetAccessToken(ctx)
		if err != nil {
			c.Close()
			return nil, yoyerrors.Wrap("getting access token", err, yoyerrors.ExitAuth).
				WithHint("Run 'yoy auth login' to re-authenticate.")
		}
		saslClient := auth.NewXOAuth2Client(email, accessToken)
		if err := c.Authenticate(saslClient); err != nil {
			c.Close()
			return nil, yoyerrors.Wrap("IMAP authentication failed", err, yoyerrors.ExitAuth).
				WithHint("Run 'yoy auth login' to re-authenticate.")
		}
	}

	return &IMAPClient{client: c, email: email}, nil
}

// Close closes the IMAP connection.
func (ic *IMAPClient) Close() error {
	if ic.client != nil {
		return ic.client.Close()
	}
	return nil
}

// ListFolders returns all mail folders.
func (ic *IMAPClient) ListFolders() ([]Folder, error) {
	listCmd := ic.client.List("", "*", nil)
	var folders []Folder

	for {
		mbox := listCmd.Next()
		if mbox == nil {
			break
		}
		folders = append(folders, Folder{
			Name: mbox.Mailbox,
		})
	}

	if err := listCmd.Close(); err != nil {
		return nil, yoyerrors.FromIMAPError(err)
	}

	// Get message counts for each folder.
	for i := range folders {
		data, err := ic.client.Status(folders[i].Name, &imap.StatusOptions{
			NumMessages: true,
			NumUnseen:   true,
		}).Wait()
		if err != nil {
			continue
		}
		if data.NumMessages != nil {
			folders[i].Messages = *data.NumMessages
		}
		if data.NumUnseen != nil {
			folders[i].Unseen = *data.NumUnseen
		}
	}

	sort.Slice(folders, func(i, j int) bool {
		return folders[i].Name < folders[j].Name
	})

	return folders, nil
}

// CreateFolder creates a new mail folder.
func (ic *IMAPClient) CreateFolder(name string) error {
	if err := ic.client.Create(name, nil).Wait(); err != nil {
		return yoyerrors.FromIMAPError(err)
	}
	return nil
}

// DeleteFolder deletes a mail folder.
func (ic *IMAPClient) DeleteFolder(name string) error {
	if err := ic.client.Delete(name).Wait(); err != nil {
		return yoyerrors.FromIMAPError(err)
	}
	return nil
}

// ListMessages lists messages in a folder.
func (ic *IMAPClient) ListMessages(folder string, limit uint32) ([]Message, error) {
	mbox, err := ic.client.Select(folder, nil).Wait()
	if err != nil {
		return nil, yoyerrors.FromIMAPError(err)
	}

	if mbox.NumMessages == 0 {
		return nil, nil
	}

	// Calculate range: fetch the last N messages.
	start := uint32(1)
	if mbox.NumMessages > limit {
		start = mbox.NumMessages - limit + 1
	}

	var seqSet imap.SeqSet
	seqSet.AddRange(start, mbox.NumMessages)

	fetchOptions := &imap.FetchOptions{
		UID:      true,
		Flags:    true,
		Envelope: true,
	}

	fetchCmd := ic.client.Fetch(seqSet, fetchOptions)
	var messages []Message

	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		m := messageFromFetchData(msg)
		messages = append(messages, m)
	}

	if err := fetchCmd.Close(); err != nil {
		return nil, yoyerrors.FromIMAPError(err)
	}

	// Reverse to show newest first.
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// SearchMessages searches for messages matching a query in a folder.
func (ic *IMAPClient) SearchMessages(folder, query string) ([]Message, error) {
	if _, err := ic.client.Select(folder, nil).Wait(); err != nil {
		return nil, yoyerrors.FromIMAPError(err)
	}

	// Search in subject and from headers using OR.
	criteria := &imap.SearchCriteria{
		Or: [][2]imap.SearchCriteria{
			{
				imap.SearchCriteria{Header: []imap.SearchCriteriaHeaderField{{Key: "Subject", Value: query}}},
				imap.SearchCriteria{Header: []imap.SearchCriteriaHeaderField{{Key: "From", Value: query}}},
			},
		},
	}

	searchData, err := ic.client.Search(criteria, nil).Wait()
	if err != nil {
		return nil, yoyerrors.FromIMAPError(err)
	}

	seqNums := searchData.AllSeqNums()
	if len(seqNums) == 0 {
		return nil, nil
	}

	var seqSet imap.SeqSet
	seqSet.AddNum(seqNums...)

	fetchOptions := &imap.FetchOptions{
		UID:      true,
		Flags:    true,
		Envelope: true,
	}

	fetchCmd := ic.client.Fetch(seqSet, fetchOptions)
	var messages []Message

	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}
		m := messageFromFetchData(msg)
		messages = append(messages, m)
	}

	if err := fetchCmd.Close(); err != nil {
		return nil, yoyerrors.FromIMAPError(err)
	}

	// Reverse to show newest first.
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// ReadMessage fetches a complete message by UID.
func (ic *IMAPClient) ReadMessage(folder string, uid uint32) (*Message, error) {
	if _, err := ic.client.Select(folder, nil).Wait(); err != nil {
		return nil, yoyerrors.FromIMAPError(err)
	}

	uidSet := imap.UIDSetNum(imap.UID(uid))

	fetchOptions := &imap.FetchOptions{
		UID:         true,
		Flags:       true,
		Envelope:    true,
		BodySection: []*imap.FetchItemBodySection{{}},
	}

	fetchCmd := ic.client.Fetch(uidSet, fetchOptions)

	msg := fetchCmd.Next()
	if msg == nil {
		if err := fetchCmd.Close(); err != nil {
			return nil, yoyerrors.FromIMAPError(err)
		}
		return nil, yoyerrors.New(fmt.Sprintf("message UID %d not found", uid), yoyerrors.ExitNotFound)
	}

	m := messageFromFetchData(msg)

	if err := fetchCmd.Close(); err != nil {
		return nil, yoyerrors.FromIMAPError(err)
	}

	return &m, nil
}

// DeleteMessage marks a message as deleted and expunges.
func (ic *IMAPClient) DeleteMessage(folder string, uid uint32) error {
	if _, err := ic.client.Select(folder, nil).Wait(); err != nil {
		return yoyerrors.FromIMAPError(err)
	}

	uidSet := imap.UIDSetNum(imap.UID(uid))

	storeCmd := ic.client.Store(uidSet, &imap.StoreFlags{
		Op:    imap.StoreFlagsAdd,
		Flags: []imap.Flag{imap.FlagDeleted},
	}, nil)
	if err := storeCmd.Close(); err != nil {
		return yoyerrors.FromIMAPError(err)
	}

	expungeCmd := ic.client.Expunge()
	for expungeCmd.Next() != 0 {
		// Drain expunge notifications.
	}
	if err := expungeCmd.Close(); err != nil {
		return yoyerrors.FromIMAPError(err)
	}

	return nil
}

// MoveMessage moves a message to a different folder.
func (ic *IMAPClient) MoveMessage(folder string, uid uint32, destFolder string) error {
	if _, err := ic.client.Select(folder, nil).Wait(); err != nil {
		return yoyerrors.FromIMAPError(err)
	}

	uidSet := imap.UIDSetNum(imap.UID(uid))

	if _, err := ic.client.Move(uidSet, destFolder).Wait(); err != nil {
		return yoyerrors.FromIMAPError(err)
	}

	return nil
}

// SetFlags sets flags on a message.
func (ic *IMAPClient) SetFlags(folder string, uid uint32, flags []imap.Flag, add bool) error {
	if _, err := ic.client.Select(folder, nil).Wait(); err != nil {
		return yoyerrors.FromIMAPError(err)
	}

	uidSet := imap.UIDSetNum(imap.UID(uid))

	op := imap.StoreFlagsAdd
	if !add {
		op = imap.StoreFlagsDel
	}

	storeCmd := ic.client.Store(uidSet, &imap.StoreFlags{
		Op:    op,
		Flags: flags,
	}, nil)
	return storeCmd.Close()
}

// StarMessage adds the \Flagged flag to a message.
func (ic *IMAPClient) StarMessage(folder string, uid uint32) error {
	return ic.SetFlags(folder, uid, []imap.Flag{imap.FlagFlagged}, true)
}

// UnstarMessage removes the \Flagged flag from a message.
func (ic *IMAPClient) UnstarMessage(folder string, uid uint32) error {
	return ic.SetFlags(folder, uid, []imap.Flag{imap.FlagFlagged}, false)
}

// MarkRead adds the \Seen flag to a message.
func (ic *IMAPClient) MarkRead(folder string, uid uint32) error {
	return ic.SetFlags(folder, uid, []imap.Flag{imap.FlagSeen}, true)
}

// MarkUnread removes the \Seen flag from a message.
func (ic *IMAPClient) MarkUnread(folder string, uid uint32) error {
	return ic.SetFlags(folder, uid, []imap.Flag{imap.FlagSeen}, false)
}

// messageFromFetchData extracts a Message from IMAP fetch data.
func messageFromFetchData(msg *imapclient.FetchMessageData) Message {
	m := Message{}

	for {
		item := msg.Next()
		if item == nil {
			break
		}

		switch data := item.(type) {
		case imapclient.FetchItemDataUID:
			m.UID = uint32(data.UID)
		case imapclient.FetchItemDataFlags:
			m.Flags = flagsToStrings(data.Flags)
			m.Seen = containsFlag(data.Flags, imap.FlagSeen)
			m.Flagged = containsFlag(data.Flags, imap.FlagFlagged)
		case imapclient.FetchItemDataEnvelope:
			env := data.Envelope
			if env == nil {
				continue
			}
			m.Subject = env.Subject
			m.Date = env.Date
			m.MessageID = env.MessageID
			if len(env.InReplyTo) > 0 {
				m.InReplyTo = env.InReplyTo[0]
			}

			if len(env.From) > 0 {
				m.From = Address{Name: env.From[0].Name, Address: env.From[0].Addr()}
			}
			m.To = envelopeAddresses(env.To)
			m.Cc = envelopeAddresses(env.Cc)
			m.ReplyTo = envelopeAddresses(env.ReplyTo)
		case imapclient.FetchItemDataBodySection:
			body, err := io.ReadAll(data.Literal)
			if err != nil {
				continue
			}
			parsed, err := ParseMessage(io.NopCloser(strings.NewReader(string(body))))
			if err == nil {
				m.Body = parsed.Body
				m.HTMLBody = parsed.HTMLBody
				m.Attachments = parsed.Attachments
				if parsed.InReplyTo != "" {
					m.InReplyTo = parsed.InReplyTo
				}
				if len(parsed.References) > 0 {
					m.References = parsed.References
				}
			}
		}
	}

	return m
}

func envelopeAddresses(addrs []imap.Address) []Address {
	if len(addrs) == 0 {
		return nil
	}
	result := make([]Address, len(addrs))
	for i, a := range addrs {
		result[i] = Address{Name: a.Name, Address: a.Addr()}
	}
	return result
}

func flagsToStrings(flags []imap.Flag) []string {
	result := make([]string, len(flags))
	for i, f := range flags {
		result[i] = strings.TrimPrefix(string(f), "\\")
	}
	return result
}

func containsFlag(flags []imap.Flag, flag imap.Flag) bool {
	for _, f := range flags {
		if f == flag {
			return true
		}
	}
	return false
}
