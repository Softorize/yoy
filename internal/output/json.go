package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/Softorize/yoy/internal/yahoo"
)

// JSONFormatter outputs data as pretty-printed JSON.
type JSONFormatter struct{}

func writeJSON(w io.Writer, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

func (f *JSONFormatter) FormatMessages(w io.Writer, messages []yahoo.Message) error {
	return writeJSON(w, messages)
}

func (f *JSONFormatter) FormatMessage(w io.Writer, message *yahoo.Message) error {
	return writeJSON(w, message)
}

func (f *JSONFormatter) FormatFolders(w io.Writer, folders []yahoo.Folder) error {
	return writeJSON(w, folders)
}

func (f *JSONFormatter) FormatKeyValue(w io.Writer, data map[string]string) error {
	return writeJSON(w, data)
}
