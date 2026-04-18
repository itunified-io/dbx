package cloud_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validAWSBlueprint = `
metadata:
  name: AWS Oracle Production Setup
  provider: aws
  profile: prod
  ticket: JIRA-6789

infrastructure:
  network:
    vpc: existing
    vpc_id: vpc-0abc123def456
    subnet_ids:
      db_primary: subnet-0abc123
      db_standby: subnet-0def456
    security_group:
      name: dbx-oracle-prod
      ingress:
        - port: 1521
          source: 10.0.0.0/16
          description: "Oracle TNS"
        - port: 22
          source: 10.0.0.0/16
          description: "SSH (dbx-agent)"
      egress:
        - port: 443
          destination: 0.0.0.0/0
          description: "dbx-central heartbeat"

  instances:
    - name: ora-prod-01
      role: db_primary
      instance_type: r6i.2xlarge
      ami: ami-oracle-linux-8-latest
      subnet: db_primary
      storage:
        - name: root
          size_gb: 100
          type: gp3
          iops: 3000
        - name: data
          size_gb: 500
          type: io2
          iops: 10000
          mount: /u01
        - name: redo
          size_gb: 100
          type: io2
          iops: 5000
          mount: /u02
        - name: backup
          size_gb: 1000
          type: gp3
          mount: /u03
      tags:
        Environment: production
        Engine: oracle
        ManagedBy: dbx

    - name: ora-prod-02
      role: db_standby
      instance_type: r6i.2xlarge
      ami: ami-oracle-linux-8-latest
      subnet: db_standby
      storage:
        - name: root
          size_gb: 100
          type: gp3
          iops: 3000
        - name: data
          size_gb: 500
          type: io2
          iops: 10000
          mount: /u01

  load_balancer:
    type: nlb
    name: dbx-oracle-prod-lb
    listeners:
      - port: 1521
        target_port: 1521
        protocol: TCP
        targets:
          - ora-prod-01
          - ora-prod-02

database:
  engine: oracle
  version: "19c"
  gold_image: oracle/19c-april-2026.yaml
  parameter_profile: parameters/oracle/prod-19c.yaml
  template: oracle-database-prod

monitoring:
  agent: auto_install
  template: oracle-database-prod
  alert_routing:
    slack: "#dba-prod-alerts"
    pagerduty: "oracle-prod-oncall"
`

func TestParseBlueprint_ValidAWS(t *testing.T) {
	bp, err := cloud.ParseBlueprint([]byte(validAWSBlueprint))
	require.NoError(t, err)
	assert.Equal(t, "AWS Oracle Production Setup", bp.Metadata.Name)
	assert.Equal(t, "aws", bp.Metadata.Provider)
	assert.Equal(t, "prod", bp.Metadata.Profile)
	assert.Equal(t, "JIRA-6789", bp.Metadata.Ticket)
	assert.Len(t, bp.Infrastructure.Instances, 2)
	assert.Equal(t, "ora-prod-01", bp.Infrastructure.Instances[0].Name)
	assert.Equal(t, "db_primary", bp.Infrastructure.Instances[0].Role)
	assert.Len(t, bp.Infrastructure.Instances[0].Storage, 4)
	assert.Equal(t, "oracle", bp.Database.Engine)
	assert.Equal(t, "19c", bp.Database.Version)
}

func TestParseBlueprint_Validation(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name:    "missing metadata.name",
			yaml:    "metadata:\n  provider: aws\n  profile: prod\n",
			wantErr: "metadata.name is required",
		},
		{
			name:    "missing metadata.provider",
			yaml:    "metadata:\n  name: test\n  profile: prod\n",
			wantErr: "metadata.provider is required",
		},
		{
			name:    "missing metadata.profile",
			yaml:    "metadata:\n  name: test\n  provider: aws\n",
			wantErr: "metadata.profile is required",
		},
		{
			name:    "invalid provider",
			yaml:    "metadata:\n  name: test\n  provider: gcp\n  profile: prod\n",
			wantErr: "unsupported provider",
		},
		{
			name:    "instance with zero storage size",
			yaml:    "metadata:\n  name: test\n  provider: aws\n  profile: prod\ninfrastructure:\n  instances:\n    - name: test\n      instance_type: t3.micro\n      ami: ami-123\n      storage:\n        - name: bad\n          size_gb: 0\n          type: gp3\n",
			wantErr: "volume size must be > 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cloud.ParseBlueprint([]byte(tt.yaml))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestParseBlueprint_AzureManagedDB(t *testing.T) {
	yaml := `
metadata:
  name: Azure PG HA Production
  provider: azure
  profile: prod

infrastructure:
  resource_group: rg-dbx-prod-westeurope

  managed_database:
    type: azure_flexible_server
    name: pg-prod-flex
    sku: GP_Standard_D8s_v3
    version: "16"
    ha_mode: ZoneRedundant
    backup_retention_days: 35
    geo_redundant_backup: true
    storage_gb: 512
    vnet_integration: snet-db

database:
  engine: postgres
  version: "16"
`
	bp, err := cloud.ParseBlueprint([]byte(yaml))
	require.NoError(t, err)
	assert.Equal(t, "azure", bp.Metadata.Provider)
	require.NotNil(t, bp.Infrastructure.ManagedDatabase)
	assert.Equal(t, "azure_flexible_server", bp.Infrastructure.ManagedDatabase.Type)
	assert.Equal(t, "ZoneRedundant", bp.Infrastructure.ManagedDatabase.HAMode)
	assert.Equal(t, true, bp.Infrastructure.ManagedDatabase.GeoRedundantBackup)
}

func TestBlueprint_InstanceNames(t *testing.T) {
	bp, err := cloud.ParseBlueprint([]byte(validAWSBlueprint))
	require.NoError(t, err)
	names := bp.InstanceNames()
	assert.Equal(t, []string{"ora-prod-01", "ora-prod-02"}, names)
}

func TestBlueprint_TotalStorageGB(t *testing.T) {
	bp, err := cloud.ParseBlueprint([]byte(validAWSBlueprint))
	require.NoError(t, err)
	// ora-prod-01: 100+500+100+1000 = 1700, ora-prod-02: 100+500 = 600 -> 2300 total
	assert.Equal(t, 2300, bp.TotalStorageGB())
}
