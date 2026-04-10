package format_test

import (
	"testing"

	"github.com/itunified-io/dbx/internal/format"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFormat(t *testing.T) {
	assert.Equal(t, format.JSON, format.ParseFormat("json"))
	assert.Equal(t, format.YAML, format.ParseFormat("yaml"))
	assert.Equal(t, format.Table, format.ParseFormat("table"))
	assert.Equal(t, format.Table, format.ParseFormat(""))
}

func TestFormatJSON(t *testing.T) {
	data := map[string]string{"name": "prod-orcl", "type": "oracle_database"}
	out, err := format.FormatOutput(data, format.JSON)
	require.NoError(t, err)
	assert.Contains(t, out, `"name": "prod-orcl"`)
}

func TestFormatYAML(t *testing.T) {
	data := map[string]string{"name": "prod-orcl"}
	out, err := format.FormatOutput(data, format.YAML)
	require.NoError(t, err)
	assert.Contains(t, out, "name: prod-orcl")
}

func TestFormatTable(t *testing.T) {
	headers := []string{"NAME", "TYPE", "STATUS"}
	rows := []format.TableRow{
		{Columns: []string{"prod-orcl", "oracle_database", "active"}},
		{Columns: []string{"prod-pg", "pg_database", "active"}},
	}
	out := format.FormatTable(headers, rows)
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "prod-orcl")
	assert.Contains(t, out, "prod-pg")
	assert.Contains(t, out, "---")
}
