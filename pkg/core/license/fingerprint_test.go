package license_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/core/license"
	"github.com/stretchr/testify/assert"
)

func TestGenerateFingerprint(t *testing.T) {
	fp := license.GenerateFingerprint()
	assert.Len(t, fp, 64, "fingerprint is SHA256 hex = 64 chars")

	// Same machine should produce same fingerprint
	fp2 := license.GenerateFingerprint()
	assert.Equal(t, fp, fp2)
}
