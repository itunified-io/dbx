package user_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/host/user"
	"github.com/stretchr/testify/assert"
)

func TestParseLast(t *testing.T) {
	content := `root     pts/0        10.0.0.1         Thu Apr 10 08:00   still logged in
admin    pts/1        10.0.0.2         Wed Apr  9 14:30 - 18:00  (03:30)

wtmp begins Wed Apr  2 00:00:00 2026
`
	records := user.ParseLast(content)
	assert.Len(t, records, 2)
	assert.Equal(t, "root", records[0].User)
	assert.Equal(t, "pts/0", records[0].Terminal)
	assert.Equal(t, "10.0.0.1", records[0].Source)
}
