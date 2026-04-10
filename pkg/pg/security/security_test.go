package security_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/security"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSLAudit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	ver := "TLSv1.3"
	cipher := "TLS_AES_256_GCM_SHA384"
	bits := int32(256)
	mock.ExpectQuery("SELECT s.pid").
		WillReturnRows(pgxmock.NewRows([]string{
			"pid", "ssl", "version", "cipher", "bits", "client_dn",
		}).
			AddRow(int32(100), true, &ver, &cipher, &bits, "").
			AddRow(int32(101), false, nil, nil, nil, ""))

	results, err := security.SSLAudit(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, results[0].SSL)
	assert.False(t, results[1].SSL)
}

func TestHBARules(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT line_number").
		WillReturnRows(pgxmock.NewRows([]string{
			"line_number", "type", "database", "user_name", "address", "auth_method",
		}).
			AddRow(int32(1), "host", "all", "all", "127.0.0.1/32", "scram-sha-256").
			AddRow(int32(2), "host", "all", "all", "0.0.0.0/0", "trust"))

	rules, err := security.HBARules(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, rules, 2)
	assert.Equal(t, "trust", rules[1].AuthMethod)
	assert.Equal(t, "HIGH", rules[1].RiskLevel)
}
