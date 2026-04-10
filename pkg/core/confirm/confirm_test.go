package confirm_test

import (
	"bytes"
	"testing"

	"github.com/itunified-io/dbx/pkg/core/confirm"
	"github.com/stretchr/testify/assert"
)

func TestReadOnlyNeedsNoConfirm(t *testing.T) {
	gate := confirm.New(nil, nil)
	err := gate.Check(confirm.LevelNone, "list sessions", "prod-orcl", false)
	assert.NoError(t, err)
}

func TestStandardConfirmWithFlag(t *testing.T) {
	gate := confirm.New(nil, nil)
	err := gate.Check(confirm.LevelStandard, "resize USERS to 50G", "prod-orcl", true)
	assert.NoError(t, err)
}

func TestStandardConfirmWithoutFlagBlocks(t *testing.T) {
	gate := confirm.New(nil, nil)
	err := gate.Check(confirm.LevelStandard, "resize USERS to 50G", "prod-orcl", false)
	assert.ErrorIs(t, err, confirm.ErrConfirmRequired)
}

func TestEchoBackSuccess(t *testing.T) {
	input := bytes.NewBufferString("RESIZE USERS 50G\n")
	output := &bytes.Buffer{}
	gate := confirm.New(input, output)

	err := gate.CheckEchoBack("RESIZE USERS 50G", "resize USERS tablespace to 50G on prod-orcl")
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "RESIZE USERS 50G")
}

func TestEchoBackWrongInput(t *testing.T) {
	input := bytes.NewBufferString("WRONG INPUT\n")
	output := &bytes.Buffer{}
	gate := confirm.New(input, output)

	err := gate.CheckEchoBack("RESIZE USERS 50G", "resize USERS")
	assert.ErrorIs(t, err, confirm.ErrConfirmMismatch)
}

func TestDoubleConfirmSuccess(t *testing.T) {
	input := bytes.NewBufferString("FAILOVER PROD-ORCL\nCONFIRM-FAILOVER\n")
	output := &bytes.Buffer{}
	gate := confirm.New(input, output)

	err := gate.CheckDoubleConfirm(
		"FAILOVER PROD-ORCL",
		"CONFIRM-FAILOVER",
		"failover prod-orcl to standby",
	)
	assert.NoError(t, err)
}

func TestDoubleConfirmSecondFails(t *testing.T) {
	input := bytes.NewBufferString("FAILOVER PROD-ORCL\nNOPE\n")
	output := &bytes.Buffer{}
	gate := confirm.New(input, output)

	err := gate.CheckDoubleConfirm(
		"FAILOVER PROD-ORCL",
		"CONFIRM-FAILOVER",
		"failover",
	)
	assert.ErrorIs(t, err, confirm.ErrConfirmMismatch)
}

func TestConfirmLevelConstants(t *testing.T) {
	assert.Equal(t, confirm.Level(0), confirm.LevelNone)
	assert.Equal(t, confirm.Level(1), confirm.LevelStandard)
	assert.Equal(t, confirm.Level(2), confirm.LevelEchoBack)
	assert.Equal(t, confirm.Level(3), confirm.LevelDoubleConfirm)
}
