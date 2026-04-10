// Package format provides output formatting for CLI, REST, and MCP interfaces.
package format

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Format specifies the output format.
type Format string

const (
	Table Format = "table"
	JSON  Format = "json"
	YAML  Format = "yaml"
)

// ParseFormat normalizes a format string.
func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "json":
		return JSON
	case "yaml":
		return YAML
	default:
		return Table
	}
}

// FormatOutput renders data in the specified format.
func FormatOutput(data any, f Format) (string, error) {
	switch f {
	case JSON:
		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf("json marshal: %w", err)
		}
		return string(b), nil
	case YAML:
		b, err := yaml.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("yaml marshal: %w", err)
		}
		return string(b), nil
	default:
		return formatTable(data), nil
	}
}

// TableRow represents a single row in table output.
type TableRow struct {
	Columns []string
}

// FormatTable renders rows as an aligned ASCII table.
func FormatTable(headers []string, rows []TableRow) string {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, col := range row.Columns {
			if i < len(widths) && len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}

	var sb strings.Builder
	for i, h := range headers {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(pad(h, widths[i]))
	}
	sb.WriteByte('\n')

	for i, w := range widths {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(strings.Repeat("-", w))
	}
	sb.WriteByte('\n')

	for _, row := range rows {
		for i, col := range row.Columns {
			if i > 0 {
				sb.WriteString("  ")
			}
			if i < len(widths) {
				sb.WriteString(pad(col, widths[i]))
			} else {
				sb.WriteString(col)
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func formatTable(data any) string {
	b, _ := json.MarshalIndent(data, "", "  ")
	return string(b)
}

func pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
