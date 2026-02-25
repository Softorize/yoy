package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Softorize/yoy/internal/yahoo"
)

// PlainFormatter outputs data as tab-separated values.
type PlainFormatter struct{}

func writeTSV(w io.Writer, headers []string, rows [][]string) error {
	if _, err := fmt.Fprintln(w, strings.Join(headers, "\t")); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintln(w, strings.Join(row, "\t")); err != nil {
			return err
		}
	}
	return nil
}

func (f *PlainFormatter) FormatMessages(w io.Writer, messages []yahoo.Message) error {
	rows := make([][]string, len(messages))
	for i, m := range messages {
		from := m.From.Address
		if m.From.Name != "" {
			from = m.From.Name
		}
		rows[i] = []string{
			fmt.Sprintf("%d", m.UID),
			m.Date.Format("2006-01-02 15:04"),
			from,
			m.Subject,
			strings.Join(m.Flags, ","),
		}
	}
	return writeTSV(w, []string{"UID", "Date", "From", "Subject", "Flags"}, rows)
}

func (f *PlainFormatter) FormatMessage(w io.Writer, message *yahoo.Message) error {
	from := message.From.Address
	if message.From.Name != "" {
		from = fmt.Sprintf("%s <%s>", message.From.Name, message.From.Address)
	}
	fmt.Fprintf(w, "UID\t%d\n", message.UID)
	fmt.Fprintf(w, "Date\t%s\n", message.Date.Format("2006-01-02 15:04:05 -0700"))
	fmt.Fprintf(w, "From\t%s\n", from)
	fmt.Fprintf(w, "Subject\t%s\n", message.Subject)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, message.Body)
	return nil
}

func (f *PlainFormatter) FormatFolders(w io.Writer, folders []yahoo.Folder) error {
	rows := make([][]string, len(folders))
	for i, folder := range folders {
		rows[i] = []string{
			folder.Name,
			fmt.Sprintf("%d", folder.Messages),
			fmt.Sprintf("%d", folder.Unseen),
		}
	}
	return writeTSV(w, []string{"Name", "Messages", "Unseen"}, rows)
}

func (f *PlainFormatter) FormatKeyValue(w io.Writer, data map[string]string) error {
	for key, val := range data {
		if _, err := fmt.Fprintf(w, "%s\t%s\n", key, val); err != nil {
			return err
		}
	}
	return nil
}
