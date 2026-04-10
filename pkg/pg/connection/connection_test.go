package connection_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/connection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig(host, db string) connection.ProfileConfig {
	return connection.ProfileConfig{
		Host:     host,
		Port:     5432,
		Database: db,
		SSLMode:  "disable",
		User:     "test",
		Password: "secret",
	}
}

func TestAdd(t *testing.T) {
	reg := connection.NewRegistry()
	require.NoError(t, reg.Add("dev", newTestConfig("localhost", "devdb")))
	assert.Equal(t, "dev", reg.Active(), "first profile should auto-activate")
}

func TestAddDuplicate(t *testing.T) {
	reg := connection.NewRegistry()
	require.NoError(t, reg.Add("dev", newTestConfig("localhost", "devdb")))
	err := reg.Add("dev", newTestConfig("localhost", "devdb"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGet(t *testing.T) {
	reg := connection.NewRegistry()
	require.NoError(t, reg.Add("dev", newTestConfig("localhost", "devdb")))

	p, err := reg.Get("dev")
	require.NoError(t, err)
	assert.Equal(t, "dev", p.Name)
	assert.Equal(t, "devdb", p.Config.Database)
}

func TestGetMissing(t *testing.T) {
	reg := connection.NewRegistry()
	_, err := reg.Get("nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestActiveSwitch(t *testing.T) {
	reg := connection.NewRegistry()
	require.NoError(t, reg.Add("dev", newTestConfig("localhost", "devdb")))
	require.NoError(t, reg.Add("prod", newTestConfig("prod-host", "proddb")))

	assert.Equal(t, "dev", reg.Active())
	require.NoError(t, reg.Switch("prod"))
	assert.Equal(t, "prod", reg.Active())
}

func TestSwitchMissing(t *testing.T) {
	reg := connection.NewRegistry()
	err := reg.Switch("nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestList(t *testing.T) {
	reg := connection.NewRegistry()
	require.NoError(t, reg.Add("beta", newTestConfig("b-host", "bdb")))
	require.NoError(t, reg.Add("alpha", newTestConfig("a-host", "adb")))

	infos := reg.List()
	require.Len(t, infos, 2)
	assert.Equal(t, "alpha", infos[0].Name, "list should be sorted")
	assert.Equal(t, "beta", infos[1].Name)
	assert.False(t, infos[0].Connected)
}

func TestRemove(t *testing.T) {
	reg := connection.NewRegistry()
	require.NoError(t, reg.Add("dev", newTestConfig("localhost", "devdb")))
	require.NoError(t, reg.Add("prod", newTestConfig("prod-host", "proddb")))

	require.NoError(t, reg.Remove("dev"))
	assert.Equal(t, "prod", reg.Active(), "active should switch after removing active profile")

	_, err := reg.Get("dev")
	require.Error(t, err)
}

func TestRemoveMissing(t *testing.T) {
	reg := connection.NewRegistry()
	err := reg.Remove("nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestActivePoolNoProfile(t *testing.T) {
	reg := connection.NewRegistry()
	_, err := reg.ActivePool()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no active profile")
}

func TestActivePoolNotConnected(t *testing.T) {
	reg := connection.NewRegistry()
	require.NoError(t, reg.Add("dev", newTestConfig("localhost", "devdb")))
	_, err := reg.ActivePool()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestDSN(t *testing.T) {
	cfg := connection.ProfileConfig{
		Host:     "myhost",
		Port:     5433,
		Database: "mydb",
		SSLMode:  "require",
		User:     "admin",
		Password: "pass",
	}
	expected := "postgres://admin:pass@myhost:5433/mydb?sslmode=require"
	assert.Equal(t, expected, cfg.DSN())
}

func TestDSNDefaults(t *testing.T) {
	cfg := connection.ProfileConfig{
		Host:     "localhost",
		Database: "test",
		User:     "user",
		Password: "pw",
	}
	dsn := cfg.DSN()
	assert.Contains(t, dsn, ":5432/")
	assert.Contains(t, dsn, "sslmode=prefer")
}
