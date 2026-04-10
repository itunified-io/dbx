package vault_test

import (
	"context"
	"testing"
	"time"

	"github.com/itunified-io/dbx/pkg/core/vault"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialCaching(t *testing.T) {
	calls := 0
	fetcher := func(ctx context.Context, path string) (*vault.Credential, error) {
		calls++
		return &vault.Credential{
			Username: "sys",
			Password: "secret",
			LeaseTTL: 1 * time.Hour,
		}, nil
	}

	client := vault.NewClient(vault.WithFetcher(fetcher), vault.WithCacheTTL(5*time.Minute))

	ctx := context.Background()
	cred, err := client.GetCredential(ctx, "secret/data/oracle/prod-orcl")
	require.NoError(t, err)
	assert.Equal(t, "sys", cred.Username)

	cred2, err := client.GetCredential(ctx, "secret/data/oracle/prod-orcl")
	require.NoError(t, err)
	assert.Equal(t, "sys", cred2.Username)
	assert.Equal(t, 1, calls, "should only call fetcher once due to cache")
}

func TestFallbackToInline(t *testing.T) {
	client := vault.NewClient()

	cred := &vault.Credential{Username: "admin", Password: "inline-pw"}
	result := client.ResolveCredential(context.Background(), "inline", "", cred)
	assert.Equal(t, "admin", result.Username)
	assert.Equal(t, "inline-pw", result.Password)
}

func TestCredentialRedacted(t *testing.T) {
	c := vault.Credential{
		Username:      "sys",
		Password:      "secret123",
		ConnectString: "db-prod:1521/ORCL",
	}
	assert.Equal(t, "sys/***@db-prod:1521/ORCL", c.Redacted())
}
