package cloud_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/itunified-io/dbx/pkg/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockVaultClient implements a minimal Vault reader for tests.
type mockVaultClient struct {
	secrets map[string]map[string]interface{}
}

func (m *mockVaultClient) ReadSecret(_ context.Context, path string) (map[string]interface{}, error) {
	if s, ok := m.secrets[path]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("secret not found: %s", path)
}

func TestCredentialLoader_AWS(t *testing.T) {
	vault := &mockVaultClient{
		secrets: map[string]map[string]interface{}{
			"secret/data/cloud/aws/prod": {
				"access_key_id":     "AKIAEXAMPLE",
				"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLE",
				"region":            "eu-west-1",
			},
		},
	}

	loader := cloud.NewCredentialLoader(vault)
	creds, err := loader.LoadAWS(context.Background(), "prod")
	require.NoError(t, err)
	assert.Equal(t, "AKIAEXAMPLE", creds.AccessKeyID)
	assert.Equal(t, "eu-west-1", creds.Region)
}

func TestCredentialLoader_Azure(t *testing.T) {
	vault := &mockVaultClient{
		secrets: map[string]map[string]interface{}{
			"secret/data/cloud/azure/prod": {
				"tenant_id":       "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
				"client_id":       "yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
				"client_secret":   "test-secret",
				"subscription_id": "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz",
			},
		},
	}

	loader := cloud.NewCredentialLoader(vault)
	creds, err := loader.LoadAzure(context.Background(), "prod")
	require.NoError(t, err)
	assert.Equal(t, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", creds.TenantID)
	assert.Equal(t, "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz", creds.SubscriptionID)
}

func TestCredentialLoader_OCI(t *testing.T) {
	vault := &mockVaultClient{
		secrets: map[string]map[string]interface{}{
			"secret/data/cloud/oci/prod": {
				"tenancy_ocid": "ocid1.tenancy.oc1..xxxx",
				"user_ocid":    "ocid1.user.oc1..xxxx",
				"fingerprint":  "aa:bb:cc:dd:ee:ff",
				"private_key":  "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
				"region":       "eu-frankfurt-1",
			},
		},
	}

	loader := cloud.NewCredentialLoader(vault)
	creds, err := loader.LoadOCI(context.Background(), "prod")
	require.NoError(t, err)
	assert.Equal(t, "ocid1.tenancy.oc1..xxxx", creds.TenancyOCID)
	assert.Equal(t, "eu-frankfurt-1", creds.Region)
}

func TestCredentialLoader_NotFound(t *testing.T) {
	vault := &mockVaultClient{secrets: map[string]map[string]interface{}{}}

	loader := cloud.NewCredentialLoader(vault)
	_, err := loader.LoadAWS(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "secret not found")
}
