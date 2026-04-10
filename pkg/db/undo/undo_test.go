package undo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLConstants(t *testing.T) {
	assert.Contains(t, ListSQL, "dba_undo_extents")
	assert.Contains(t, SegmentInfoSQL, "v$rollstat")
}
