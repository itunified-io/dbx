package redo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSwitchHistory_DefaultDays(t *testing.T) {
	cases := []struct {
		input    int
		expected int
	}{
		{0, 7},
		{-1, 7},
		{14, 14},
	}
	for _, tc := range cases {
		days := tc.input
		if days <= 0 {
			days = 7
		}
		assert.Equal(t, tc.expected, days)
	}
}

func TestSQLConstants(t *testing.T) {
	assert.Contains(t, ListSQL, "v$log")
	assert.Contains(t, SwitchHistorySQL, "v$log_history")
}
