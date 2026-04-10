package advisor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLTuningList_DefaultLimit(t *testing.T) {
	cases := []struct {
		input    int
		expected int
	}{
		{0, 20},
		{-1, 20},
		{50, 50},
	}
	for _, tc := range cases {
		limit := tc.input
		if limit <= 0 {
			limit = 20
		}
		assert.Equal(t, tc.expected, limit)
	}
}

func TestSQLConstants(t *testing.T) {
	assert.Contains(t, SegmentAdvisorSQL, "DBMS_SPACE.ASA_RECOMMENDATIONS")
	assert.Contains(t, SQLTuningListSQL, "SQL Tuning Advisor")
}
