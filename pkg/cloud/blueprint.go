package cloud

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Blueprint represents a complete provisioning blueprint (YAML file).
type Blueprint struct {
	Metadata       BlueprintMetadata       `yaml:"metadata"`
	Infrastructure BlueprintInfrastructure `yaml:"infrastructure"`
	Database       BlueprintDatabase       `yaml:"database,omitempty"`
	Monitoring     BlueprintMonitoring     `yaml:"monitoring,omitempty"`
}

// BlueprintMetadata identifies the blueprint.
type BlueprintMetadata struct {
	Name     string `yaml:"name"`
	Provider string `yaml:"provider"` // aws, azure, oci, onprem
	Profile  string `yaml:"profile"`  // Vault credential profile
	Ticket   string `yaml:"ticket"`   // Change ticket reference
}

// BlueprintInfrastructure describes what to provision.
type BlueprintInfrastructure struct {
	Network         *BlueprintNetwork     `yaml:"network,omitempty"`
	Instances       []BlueprintInstance    `yaml:"instances,omitempty"`
	LoadBalancer    *BlueprintLoadBalancer `yaml:"load_balancer,omitempty"`
	ManagedDatabase *BlueprintManagedDB   `yaml:"managed_database,omitempty"`
	ResourceGroup   string                `yaml:"resource_group,omitempty"` // Azure
	Compartment     string                `yaml:"compartment,omitempty"`    // OCI
}

// BlueprintNetwork describes network configuration.
type BlueprintNetwork struct {
	VPC           string            `yaml:"vpc"`                    // existing | create
	VPCID         string            `yaml:"vpc_id,omitempty"`       // If existing
	SubnetIDs     map[string]string `yaml:"subnet_ids,omitempty"`
	SecurityGroup *BlueprintSG      `yaml:"security_group,omitempty"`
}

// BlueprintSG describes a security group to create.
type BlueprintSG struct {
	Name    string        `yaml:"name"`
	Ingress []IngressRule `yaml:"ingress,omitempty"`
	Egress  []EgressRule  `yaml:"egress,omitempty"`
}

// BlueprintInstance describes a single compute instance.
type BlueprintInstance struct {
	Name         string            `yaml:"name"`
	Role         string            `yaml:"role"`
	InstanceType string            `yaml:"instance_type"`
	AMI          string            `yaml:"ami"`    // AWS AMI / Azure image / OCI image
	Subnet       string            `yaml:"subnet"` // Reference to subnet_ids key
	Storage      []VolumeSpec      `yaml:"storage,omitempty"`
	Tags         map[string]string `yaml:"tags,omitempty"`

	// Azure
	VMSize string         `yaml:"vm_size,omitempty"`
	Image  *AzureImageRef `yaml:"image,omitempty"`

	// OCI
	Shape    string `yaml:"shape,omitempty"`
	OCPUs    int    `yaml:"ocpus,omitempty"`
	MemoryGB int    `yaml:"memory_gb,omitempty"`
}

// AzureImageRef defines an Azure marketplace image.
type AzureImageRef struct {
	Publisher string `yaml:"publisher"`
	Offer     string `yaml:"offer"`
	SKU       string `yaml:"sku"`
}

// BlueprintLoadBalancer describes a load balancer.
type BlueprintLoadBalancer struct {
	Type      string              `yaml:"type"` // nlb, alb, azure_lb, oci_lb
	Name      string              `yaml:"name"`
	Listeners []BlueprintListener `yaml:"listeners"`
}

// BlueprintListener describes a load balancer listener with target references.
type BlueprintListener struct {
	Port       int      `yaml:"port"`
	TargetPort int      `yaml:"target_port"`
	Protocol   string   `yaml:"protocol"`
	Targets    []string `yaml:"targets"` // References to instance names
}

// BlueprintManagedDB describes a managed database service.
type BlueprintManagedDB struct {
	Type                string `yaml:"type"` // rds, azure_flexible_server, autonomous, db_system
	Name                string `yaml:"name"`
	SKU                 string `yaml:"sku,omitempty"`
	InstanceClass       string `yaml:"instance_class,omitempty"`
	Version             string `yaml:"version"`
	HAMode              string `yaml:"ha_mode,omitempty"`
	BackupRetentionDays int    `yaml:"backup_retention_days,omitempty"`
	GeoRedundantBackup  bool   `yaml:"geo_redundant_backup,omitempty"`
	StorageGB           int    `yaml:"storage_gb,omitempty"`
	VNetIntegration     string `yaml:"vnet_integration,omitempty"`

	// OCI Autonomous
	WorkloadType string `yaml:"workload_type,omitempty"`
	CPUCount     int    `yaml:"cpu_count,omitempty"`
	StorageTB    int    `yaml:"storage_tb,omitempty"`
	AutoScaling  bool   `yaml:"auto_scaling,omitempty"`
	LicenseType  string `yaml:"license_type,omitempty"`
}

// BlueprintDatabase describes post-provisioning database configuration.
type BlueprintDatabase struct {
	Engine           string `yaml:"engine"`
	Version          string `yaml:"version"`
	GoldImage        string `yaml:"gold_image,omitempty"`
	ParameterProfile string `yaml:"parameter_profile,omitempty"`
	Template         string `yaml:"template,omitempty"`
}

// BlueprintMonitoring describes monitoring configuration.
type BlueprintMonitoring struct {
	Agent        string            `yaml:"agent"`    // auto_install, manual
	Template     string            `yaml:"template"`
	AlertRouting map[string]string `yaml:"alert_routing,omitempty"`
}

// ParseBlueprint parses YAML bytes into a Blueprint and validates it.
func ParseBlueprint(data []byte) (*Blueprint, error) {
	var bp Blueprint
	if err := yaml.Unmarshal(data, &bp); err != nil {
		return nil, fmt.Errorf("invalid blueprint YAML: %w", err)
	}
	if err := bp.Validate(); err != nil {
		return nil, err
	}
	return &bp, nil
}

// Validate checks all required fields and cross-references.
func (bp *Blueprint) Validate() error {
	// Metadata
	if bp.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if bp.Metadata.Provider == "" {
		return fmt.Errorf("metadata.provider is required")
	}
	if bp.Metadata.Profile == "" {
		return fmt.Errorf("metadata.profile is required")
	}

	// Validate provider
	supported := map[string]bool{"aws": true, "azure": true, "oci": true, "onprem": true}
	if !supported[bp.Metadata.Provider] {
		return fmt.Errorf("unsupported provider %q (supported: aws, azure, oci, onprem)", bp.Metadata.Provider)
	}

	// Validate instances
	for i, inst := range bp.Infrastructure.Instances {
		if inst.Name == "" {
			return fmt.Errorf("infrastructure.instances[%d].name is required", i)
		}
		for j, vol := range inst.Storage {
			if err := vol.Validate(); err != nil {
				return fmt.Errorf("infrastructure.instances[%d].storage[%d]: %w", i, j, err)
			}
		}
	}

	return nil
}

// InstanceNames returns the names of all instances in the blueprint.
func (bp *Blueprint) InstanceNames() []string {
	names := make([]string, len(bp.Infrastructure.Instances))
	for i, inst := range bp.Infrastructure.Instances {
		names[i] = inst.Name
	}
	return names
}

// TotalStorageGB returns the total storage across all instances.
func (bp *Blueprint) TotalStorageGB() int {
	total := 0
	for _, inst := range bp.Infrastructure.Instances {
		for _, vol := range inst.Storage {
			total += vol.SizeGB
		}
	}
	return total
}
