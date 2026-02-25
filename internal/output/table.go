package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Softorize/yoy/internal/yahoo"
	"github.com/olekukonko/tablewriter"
)

// TableFormatter outputs data as formatted tables.
type TableFormatter struct {
	color bool
}

func (f *TableFormatter) newTable(w io.Writer, headers []string) *tablewriter.Table {
	table := tablewriter.NewWriter(w)
	if f.color {
		colored := make([]string, len(headers))
		for i, h := range headers {
			colored[i] = Colorize(h, BoldCyan, true)
		}
		table.SetHeader(colored)
	} else {
		table.SetHeader(headers)
	}
	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("  ")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)
	return table
}

func (f *TableFormatter) FormatMessages(w io.Writer, messages []yahoo.Message) error {
	table := f.newTable(w, []string{"UID", "Date", "From", "Subject", "Flags"})
	for _, m := range messages {
		from := m.From.Address
		if m.From.Name != "" {
			from = m.From.Name
		}
		// Truncate subject to 50 chars.
		subject := m.Subject
		if len(subject) > 50 {
			subject = subject[:47] + "..."
		}
		flags := strings.Join(m.Flags, ",")

		uid := fmt.Sprintf("%d", m.UID)
		row := []string{uid, m.Date.Format("2006-01-02 15:04"), from, subject, flags}

		if f.color && !m.Seen {
			for i := range row {
				row[i] = Colorize(row[i], Bold, true)
			}
		}

		table.Append(row)
	}
	table.Render()
	return nil
}

func (f *TableFormatter) FormatMessage(w io.Writer, message *yahoo.Message) error {
	from := message.From.Address
	if message.From.Name != "" {
		from = fmt.Sprintf("%s <%s>", message.From.Name, message.From.Address)
	}

	toAddrs := make([]string, len(message.To))
	for i, a := range message.To {
		if a.Name != "" {
			toAddrs[i] = fmt.Sprintf("%s <%s>", a.Name, a.Address)
		} else {
			toAddrs[i] = a.Address
		}
	}

	fmt.Fprintf(w, "UID:     %d\n", message.UID)
	fmt.Fprintf(w, "Date:    %s\n", message.Date.Format("2006-01-02 15:04:05 -0700"))
	fmt.Fprintf(w, "From:    %s\n", from)
	fmt.Fprintf(w, "To:      %s\n", strings.Join(toAddrs, ", "))
	if len(message.Cc) > 0 {
		ccAddrs := make([]string, len(message.Cc))
		for i, a := range message.Cc {
			if a.Name != "" {
				ccAddrs[i] = fmt.Sprintf("%s <%s>", a.Name, a.Address)
			} else {
				ccAddrs[i] = a.Address
			}
		}
		fmt.Fprintf(w, "Cc:      %s\n", strings.Join(ccAddrs, ", "))
	}
	fmt.Fprintf(w, "Subject: %s\n", message.Subject)
	if len(message.Flags) > 0 {
		fmt.Fprintf(w, "Flags:   %s\n", strings.Join(message.Flags, ", "))
	}
	if len(message.Attachments) > 0 {
		names := make([]string, len(message.Attachments))
		for i, a := range message.Attachments {
			names[i] = a.Filename
		}
		fmt.Fprintf(w, "Attach:  %s\n", strings.Join(names, ", "))
	}
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, message.Body)
	return nil
}

func (f *TableFormatter) FormatFolders(w io.Writer, folders []yahoo.Folder) error {
	table := f.newTable(w, []string{"Name", "Messages", "Unseen"})
	for _, folder := range folders {
		table.Append([]string{
			folder.Name,
			fmt.Sprintf("%d", folder.Messages),
			fmt.Sprintf("%d", folder.Unseen),
		})
	}
	table.Render()
	return nil
}

func (f *TableFormatter) FormatKeyValue(w io.Writer, data map[string]string) error {
	table := f.newTable(w, []string{"Field", "Value"})
	for key, val := range data {
		k := key
		if f.color {
			k = Colorize(key, Bold, true)
		}
		table.Append([]string{k, val})
	}
	table.Render()
	return nil
}
