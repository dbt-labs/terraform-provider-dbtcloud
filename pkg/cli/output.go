package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// noColor returns true if the NO_COLOR environment variable is set.
func noColor() bool {
	_, ok := os.LookupEnv("NO_COLOR")
	return ok
}

// newStyle creates a new lipgloss.Style, returning an empty style if NO_COLOR is set.
func newStyle() lipgloss.Style {
	return lipgloss.NewStyle()
}

// HeaderStyle returns a bold, blue style for headers.
func HeaderStyle() lipgloss.Style {
	s := newStyle().Bold(true)
	if !noColor() {
		s = s.Foreground(lipgloss.Color("12"))
	}
	return s
}

// ErrorStyle returns a bold, red style for errors.
func ErrorStyle() lipgloss.Style {
	s := newStyle().Bold(true)
	if !noColor() {
		s = s.Foreground(lipgloss.Color("9"))
	}
	return s
}

// SuccessStyle returns a bold, green style for success messages.
func SuccessStyle() lipgloss.Style {
	s := newStyle().Bold(true)
	if !noColor() {
		s = s.Foreground(lipgloss.Color("10"))
	}
	return s
}

// MutedStyle returns a grey style for muted/secondary text.
func MutedStyle() lipgloss.Style {
	s := newStyle()
	if !noColor() {
		s = s.Foreground(lipgloss.Color("8"))
	}
	return s
}

// PrintTable renders a styled table to stdout with the given headers and rows.
// If the terminal is narrower than the table, columns are shrunk and values truncated.
func PrintTable(headers []string, rows [][]string) {
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		Headers(headers...).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return HeaderStyle()
			}
			return newStyle()
		})

	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		t = t.Width(w).Wrap(false)
	}

	fmt.Println(t)
}

// KeyValue represents a single key-value pair for display.
type KeyValue struct {
	Key   string
	Value string
}

// PrintKeyValue renders key-value pairs with aligned keys to stdout.
func PrintKeyValue(pairs []KeyValue) {
	if len(pairs) == 0 {
		return
	}

	// Find the longest key for alignment.
	maxKeyLen := 0
	for _, p := range pairs {
		if len(p.Key) > maxKeyLen {
			maxKeyLen = len(p.Key)
		}
	}

	keyStyle := HeaderStyle()
	mutedSep := MutedStyle()

	for _, p := range pairs {
		paddedKey := p.Key + strings.Repeat(" ", maxKeyLen-len(p.Key))
		fmt.Printf("%s %s %s\n",
			keyStyle.Render(paddedKey),
			mutedSep.Render(":"),
			p.Value,
		)
	}
}

// PrintJSON outputs pretty-printed JSON to stdout.
func PrintJSON(data any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

// PrintYAML outputs YAML to stdout.
// It marshals via JSON first so that json struct tags are respected for field names.
func PrintYAML(data any) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}
	var jsonObj any
	if err := json.Unmarshal(jsonBytes, &jsonObj); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}
	enc := yaml.NewEncoder(os.Stdout)
	defer enc.Close()
	if err := enc.Encode(jsonObj); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}
	return nil
}

// FormatOutput routes output to JSON, YAML, or table format based on the format string.
// The tableFunc is called for the default table/human-readable output.
func FormatOutput(format string, data any, tableFunc func()) error {
	switch strings.ToLower(format) {
	case "json":
		return PrintJSON(data)
	case "yaml":
		return PrintYAML(data)
	default:
		tableFunc()
		return nil
	}
}
