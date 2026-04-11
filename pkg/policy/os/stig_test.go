package os_test

import (
	"testing"

	pos "github.com/itunified-io/dbx/pkg/policy/os"
	"github.com/stretchr/testify/assert"
)

func TestSTIGSeverityMap(t *testing.T) {
	assert.Equal(t, "critical", pos.STIGSeverityMap["CAT I"])
	assert.Equal(t, "high", pos.STIGSeverityMap["CAT II"])
	assert.Equal(t, "medium", pos.STIGSeverityMap["CAT III"])
}
