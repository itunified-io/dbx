package log_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/host/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseJournalJSON(t *testing.T) {
	output := `{"__REALTIME_TIMESTAMP":"1712744400000000","_SYSTEMD_UNIT":"sshd.service","PRIORITY":"3","MESSAGE":"Failed password for invalid user admin from 10.20.0.50 port 45678 ssh2","SYSLOG_IDENTIFIER":"sshd"}
{"__REALTIME_TIMESTAMP":"1712744500000000","_SYSTEMD_UNIT":"kernel","PRIORITY":"4","MESSAGE":"EXT4-fs error (device sda1): ext4_lookup:1856: inode #12345: comm","SYSLOG_IDENTIFIER":"kernel"}
`
	entries, err := log.ParseJournalJSON(output)
	require.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "sshd.service", entries[0].Unit)
	assert.Equal(t, 3, entries[0].Priority)
	assert.Contains(t, entries[0].Message, "Failed password")
}

func TestParseAuthLog(t *testing.T) {
	output := `Apr 10 09:00:01 db-prod sshd[12345]: Accepted publickey for oracle from 10.10.0.100 port 54321 ssh2
Apr 10 09:01:02 db-prod sshd[12346]: Failed password for invalid user admin from 10.20.0.50 port 45678 ssh2
Apr 10 09:01:03 db-prod sshd[12347]: Failed password for invalid user admin from 10.20.0.50 port 45679 ssh2
Apr 10 09:02:04 db-prod sshd[12348]: Accepted password for root from 10.10.0.1 port 22222 ssh2
`
	summary, err := log.ParseAuthLog(output)
	require.NoError(t, err)
	assert.Equal(t, 2, summary.SuccessCount)
	assert.Equal(t, 2, summary.FailedCount)
	assert.Equal(t, 1, len(summary.FailedSources))
	assert.Equal(t, 2, summary.FailedSources["10.20.0.50"])
}

func TestFilterBySeverity(t *testing.T) {
	entries := []log.Entry{
		{Priority: 3, Message: "error 1"},
		{Priority: 4, Message: "warning 1"},
		{Priority: 6, Message: "info 1"},
		{Priority: 3, Message: "error 2"},
	}
	errors := log.FilterBySeverity(entries, 3)
	assert.Len(t, errors, 2)
}

func TestFilterByUnit(t *testing.T) {
	entries := []log.Entry{
		{Unit: "sshd.service", Message: "msg1"},
		{Unit: "docker.service", Message: "msg2"},
		{Unit: "sshd.service", Message: "msg3"},
	}
	sshd := log.FilterByUnit(entries, "sshd.service")
	assert.Len(t, sshd, 2)
}
