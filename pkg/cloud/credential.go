package cloud

import (
	"context"
	"fmt"
)

// VaultReader is the minimal Vault interface needed by the credential loader.
type VaultReader interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
}

// AWSCredentials holds AWS authentication details loaded from Vault.
type AWSCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	AssumeRoleARN   string // Optional: for cross-account access
}

// AzureCredentials holds Azure Service Principal details loaded from Vault.
type AzureCredentials struct {
	TenantID       string
	ClientID       string
	ClientSecret   string
	SubscriptionID string
}

// OCICredentials holds OCI authentication details loaded from Vault.
type OCICredentials struct {
	TenancyOCID   string
	UserOCID      string
	Fingerprint   string
	PrivateKey    string
	Region        string
	CompartmentID string
}

// CredentialLoader loads cloud credentials from Vault.
type CredentialLoader struct {
	vault VaultReader
}

// NewCredentialLoader creates a new credential loader backed by Vault.
func NewCredentialLoader(vault VaultReader) *CredentialLoader {
	return &CredentialLoader{vault: vault}
}

// vaultPath returns the Vault path for a cloud provider profile.
func vaultPath(provider, profile string) string {
	return fmt.Sprintf("secret/data/cloud/%s/%s", provider, profile)
}

// LoadAWS loads AWS credentials for the given profile.
func (cl *CredentialLoader) LoadAWS(ctx context.Context, profile string) (*AWSCredentials, error) {
	data, err := cl.vault.ReadSecret(ctx, vaultPath("aws", profile))
	if err != nil {
		return nil, fmt.Errorf("load AWS credentials for profile %q: %w", profile, err)
	}

	creds := &AWSCredentials{}
	if v, ok := data["access_key_id"].(string); ok {
		creds.AccessKeyID = v
	}
	if v, ok := data["secret_access_key"].(string); ok {
		creds.SecretAccessKey = v
	}
	if v, ok := data["region"].(string); ok {
		creds.Region = v
	}
	if v, ok := data["assume_role_arn"].(string); ok {
		creds.AssumeRoleARN = v
	}
	return creds, nil
}

// LoadAzure loads Azure credentials for the given profile.
func (cl *CredentialLoader) LoadAzure(ctx context.Context, profile string) (*AzureCredentials, error) {
	data, err := cl.vault.ReadSecret(ctx, vaultPath("azure", profile))
	if err != nil {
		return nil, fmt.Errorf("load Azure credentials for profile %q: %w", profile, err)
	}

	creds := &AzureCredentials{}
	if v, ok := data["tenant_id"].(string); ok {
		creds.TenantID = v
	}
	if v, ok := data["client_id"].(string); ok {
		creds.ClientID = v
	}
	if v, ok := data["client_secret"].(string); ok {
		creds.ClientSecret = v
	}
	if v, ok := data["subscription_id"].(string); ok {
		creds.SubscriptionID = v
	}
	return creds, nil
}

// LoadOCI loads OCI credentials for the given profile.
func (cl *CredentialLoader) LoadOCI(ctx context.Context, profile string) (*OCICredentials, error) {
	data, err := cl.vault.ReadSecret(ctx, vaultPath("oci", profile))
	if err != nil {
		return nil, fmt.Errorf("load OCI credentials for profile %q: %w", profile, err)
	}

	creds := &OCICredentials{}
	if v, ok := data["tenancy_ocid"].(string); ok {
		creds.TenancyOCID = v
	}
	if v, ok := data["user_ocid"].(string); ok {
		creds.UserOCID = v
	}
	if v, ok := data["fingerprint"].(string); ok {
		creds.Fingerprint = v
	}
	if v, ok := data["private_key"].(string); ok {
		creds.PrivateKey = v
	}
	if v, ok := data["region"].(string); ok {
		creds.Region = v
	}
	if v, ok := data["compartment_id"].(string); ok {
		creds.CompartmentID = v
	}
	return creds, nil
}
