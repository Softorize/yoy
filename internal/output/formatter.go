package output

import (
	"io"

	"github.com/Softorize/yoy/internal/yahoo"
)

// Formatter defines the output formatting interface.
type Formatter interface {
	FormatMessages(w io.Writer, messages []yahoo.Message) error
	FormatMessage(w io.Writer, message *yahoo.Message) error
	FormatFolders(w io.Writer, folders []yahoo.Folder) error
	FormatKeyValue(w io.Writer, data map[string]string) error
}

// NewFormatter creates a formatter based on format type.
func NewFormatter(format string, colorEnabled bool) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{}
	case "plain":
		return &PlainFormatter{}
	default:
		return &TableFormatter{color: colorEnabled}
	}
}
