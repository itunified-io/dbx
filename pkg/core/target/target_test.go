package target_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/itunified-io/dbx/pkg/core/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOracleTarget(t *testing.T) {
	yaml := `
name: prod-orcl
type: oracle_database
description: "Production ORCL"
oracle_license:
  edition: enterprise
  options:
    - partitioning
    - diagnostics_pack
  oem_packs:
    - diagnostics
primary:
  host: db-prod.example.com
  port: 1521
  service: ORCL
  credential: vault
  vault_path: secret/data/oracle/prod-orcl
standby:
  host: db-standby.example.com
  port: 1521
  service: ORCL
  role: physical_standby
ssh:
  host: db-prod.example.com
  user: oracle
  key_path: ~/.ssh/oracle_ed25519
`
	tgt, err := target.Parse([]byte(yaml))
	require.NoError(t, err)
	assert.Equal(t, "prod-orcl", tgt.Name)
	assert.Equal(t, target.TypeOracleDatabase, tgt.Type)
	assert.Equal(t, "enterprise", tgt.OracleLicense.Edition)
	assert.Contains(t, tgt.OracleLicense.Options, "diagnostics_pack")
	assert.Equal(t, "db-prod.example.com", tgt.Primary.Host)
	assert.Equal(t, 1521, tgt.Primary.Port)
	assert.Equal(t, "oracle", tgt.SSH.User)
}

func TestParsePostgresTarget(t *testing.T) {
	yaml := `
name: prod-pg
type: pg_database
description: "Production PostgreSQL 17"
primary:
  host: pg-prod.example.com
  port: 5432
  database: appdb
  sslmode: verify-full
  credential: vault
  vault_path: secret/data/postgres/prod-pg
replica:
  host: pg-replica.example.com
  port: 5432
  database: appdb
  role: standby
cnpg:
  cluster_name: pg-prod-cluster
  namespace: cnpg-system
  k8s_context: k8s-prod
`
	tgt, err := target.Parse([]byte(yaml))
	require.NoError(t, err)
	assert.Equal(t, "prod-pg", tgt.Name)
	assert.Equal(t, target.TypePGDatabase, tgt.Type)
	assert.Equal(t, "pg-prod.example.com", tgt.Primary.Host)
	assert.Equal(t, 5432, tgt.Primary.Port)
	assert.Equal(t, "appdb", tgt.Primary.Database)
	assert.Equal(t, "pg-prod-cluster", tgt.CNPG.ClusterName)
}

func TestEntityTypeConstants(t *testing.T) {
	assert.Equal(t, "oracle_database", string(target.TypeOracleDatabase))
	assert.Equal(t, "rac_database", string(target.TypeRACDatabase))
	assert.Equal(t, "oracle_listener", string(target.TypeOracleListener))
	assert.Equal(t, "oracle_asm", string(target.TypeOracleASM))
	assert.Equal(t, "oracle_host", string(target.TypeOracleHost))
	assert.Equal(t, "pg_database", string(target.TypePGDatabase))
	assert.Equal(t, "pg_cluster", string(target.TypePGCluster))
	assert.Equal(t, "host", string(target.TypeHost))
}

func TestIsOracle(t *testing.T) {
	tgt := &target.Target{Type: target.TypeOracleDatabase}
	assert.True(t, tgt.IsOracle())
	assert.False(t, tgt.IsPostgres())

	pgTgt := &target.Target{Type: target.TypePGDatabase}
	assert.False(t, pgTgt.IsOracle())
	assert.True(t, pgTgt.IsPostgres())
}

func TestRegistryLoadFromDir(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "prod-orcl.yaml"), []byte(`
name: prod-orcl
type: oracle_database
primary:
  host: db-prod.example.com
  port: 1521
  service: ORCL
`), 0644)

	os.WriteFile(filepath.Join(dir, "prod-pg.yaml"), []byte(`
name: prod-pg
type: pg_database
primary:
  host: pg-prod.example.com
  port: 5432
  database: appdb
`), 0644)

	reg, err := target.NewRegistry(dir)
	require.NoError(t, err)
	assert.Equal(t, 2, reg.Count())

	tgt, err := reg.Get("prod-orcl")
	require.NoError(t, err)
	assert.Equal(t, target.TypeOracleDatabase, tgt.Type)

	tgt, err = reg.Get("prod-pg")
	require.NoError(t, err)
	assert.Equal(t, target.TypePGDatabase, tgt.Type)
}

func TestRegistryGetNotFound(t *testing.T) {
	dir := t.TempDir()
	reg, err := target.NewRegistry(dir)
	require.NoError(t, err)

	_, err = reg.Get("nonexistent")
	assert.ErrorIs(t, err, target.ErrTargetNotFound)
}

func TestRegistryListByType(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "a.yaml"), []byte("name: a\ntype: oracle_database\n"), 0644)
	os.WriteFile(filepath.Join(dir, "b.yaml"), []byte("name: b\ntype: pg_database\n"), 0644)
	os.WriteFile(filepath.Join(dir, "c.yaml"), []byte("name: c\ntype: oracle_database\n"), 0644)

	reg, err := target.NewRegistry(dir)
	require.NoError(t, err)

	oracleTargets := reg.ListByType(target.TypeOracleDatabase)
	assert.Len(t, oracleTargets, 2)

	pgTargets := reg.ListByType(target.TypePGDatabase)
	assert.Len(t, pgTargets, 1)
}

func TestRegistryResolveType(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "prod-orcl.yaml"), []byte("name: prod-orcl\ntype: oracle_database\n"), 0644)

	reg, err := target.NewRegistry(dir)
	require.NoError(t, err)

	resolved, err := reg.Resolve("prod-orcl", "")
	require.NoError(t, err)
	assert.Equal(t, target.TypeOracleDatabase, resolved.Type)

	resolved, err = reg.Resolve("prod-orcl", "oracle_database")
	require.NoError(t, err)
	assert.Equal(t, target.TypeOracleDatabase, resolved.Type)

	_, err = reg.Resolve("prod-orcl", "pg_database")
	assert.Error(t, err)
}
