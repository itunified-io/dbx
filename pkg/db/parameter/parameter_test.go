package parameter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLConstants(t *testing.T) {
	assert.Contains(t, ListSQL, "v$parameter")
	assert.Contains(t, DescribeSQL, ":1")
	assert.Contains(t, ModifiedSQL, "isdefault")
	assert.Contains(t, HiddenSQL, "x$ksppi")
}
