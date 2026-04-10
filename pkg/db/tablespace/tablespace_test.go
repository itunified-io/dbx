package tablespace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLConstants(t *testing.T) {
	assert.Contains(t, ListSQL, "dba_tablespaces")
	assert.Contains(t, DescribeSQL, ":1")
	assert.Contains(t, UsageSummarySQL, "SUM")
}
