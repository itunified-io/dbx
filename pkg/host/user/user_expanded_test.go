package user_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/host/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePasswd(t *testing.T) {
	content := `root:x:0:0:root:/root:/bin/bash
oracle:x:1001:1001:Oracle DBA:/home/oracle:/bin/bash
postgres:x:1002:1002:PostgreSQL:/home/postgres:/bin/bash
nologin_user:x:1003:1003:NoLogin:/home/nologin:/sbin/nologin
`
	users, err := user.ParsePasswd(content)
	require.NoError(t, err)
	assert.Len(t, users, 4)
	assert.Equal(t, "root", users[0].Name)
	assert.Equal(t, 0, users[0].UID)
	assert.Equal(t, "/bin/bash", users[0].Shell)
	assert.True(t, users[0].HasLoginShell)
	assert.False(t, users[3].HasLoginShell)
}

func TestParseGroups(t *testing.T) {
	content := `root:x:0:
oinstall:x:1001:oracle
dba:x:1002:oracle,postgres
wheel:x:10:root,oracle
`
	groups, err := user.ParseGroups(content)
	require.NoError(t, err)
	assert.Len(t, groups, 4)
	assert.Equal(t, "dba", groups[2].Name)
	assert.Equal(t, []string{"oracle", "postgres"}, groups[2].Members)
}

func TestParseWho(t *testing.T) {
	output := `oracle   pts/0        2026-04-10 09:00 (10.10.0.100)
root     pts/1        2026-04-10 08:30 (10.10.0.1)
`
	sessions, err := user.ParseWho(output)
	require.NoError(t, err)
	assert.Len(t, sessions, 2)
	assert.Equal(t, "oracle", sessions[0].User)
	assert.Equal(t, "pts/0", sessions[0].Terminal)
	assert.Equal(t, "10.10.0.100", sessions[0].RemoteHost)
}

func TestParseSudoers(t *testing.T) {
	content := `root    ALL=(ALL:ALL) ALL
%wheel  ALL=(ALL:ALL) ALL
oracle  ALL=(ALL) NOPASSWD: /usr/bin/systemctl restart oracle*
dbmon   ALL=(ALL) NOPASSWD: /usr/bin/cat /proc/*, /usr/bin/free, /usr/bin/df *
`
	rules, err := user.ParseSudoers(content)
	require.NoError(t, err)
	assert.Len(t, rules, 4)
	assert.True(t, rules[2].NOPASSWD)
	assert.Contains(t, rules[2].Commands, "/usr/bin/systemctl restart oracle*")
}
