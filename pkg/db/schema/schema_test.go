package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLConstants(t *testing.T) {
	assert.Contains(t, ListSQL, "dba_objects")
	assert.Contains(t, ObjectListSQL, ":1")
	assert.Contains(t, ObjectListByTypeSQL, ":2")
	assert.Contains(t, ObjectDescribeSQL, "object_name")
}
