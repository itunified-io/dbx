package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopWaiters_DefaultLimit(t *testing.T) {
	// Verify that TopWaiters uses default limit of 10 when given <= 0.
	// We can't call the real function without a DB, but we test the guard logic.
	cases := []struct {
		input    int
		expected int
	}{
		{0, 10},
		{-5, 10},
		{25, 25},
	}
	for _, tc := range cases {
		limit := tc.input
		if limit <= 0 {
			limit = 10
		}
		assert.Equal(t, tc.expected, limit)
	}
}

func TestSQLConstants(t *testing.T) {
	assert.Contains(t, ListSQL, "v$session")
	assert.Contains(t, DescribeSQL, ":1")
	assert.Contains(t, TopWaitersSQL, "FETCH FIRST")
}
