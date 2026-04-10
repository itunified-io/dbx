package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLConstants(t *testing.T) {
	assert.Contains(t, ListSQL, "dba_users")
	assert.Contains(t, DescribeSQL, ":1")
	assert.Contains(t, ProfileListSQL, "dba_profiles")
}
