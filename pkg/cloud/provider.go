package cloud

import (
	"context"
	"fmt"
	"time"
)

// ProviderID identifies a supported cloud provider.
type ProviderID int

const (
	AWS    ProviderID = iota // Amazon Web Services
	Azure                    // Microsoft Azure
	OCI                      // Oracle Cloud Infrastructure
	OnPrem                   // On-premises (SSH-based, no VM creation)
)

func (p ProviderID) String() string {
	switch p {
	case AWS:
		return "aws"
	case Azure:
		return "azure"
	case OCI:
		return "oci"
	case OnPrem:
		return "onprem"
	default:
		return "unknown"
	}
}

// ParseProviderID converts a string to ProviderID.
func ParseProviderID(s string) (ProviderID, error) {
	switch s {
	case "aws":
		return AWS, nil
	case "azure":
		return Azure, nil
	case "oci":
		return OCI, nil
	case "onprem":
		return OnPrem, nil
	default:
		return -1, fmt.Errorf("unknown provider: %q", s)
	}
}

// InstanceState represents the lifecycle state of a cloud instance.
type InstanceState int

const (
	StatePending    InstanceState = iota
	StateRunning
	StateStopped
	StateTerminated
	StateStopping
	StateStarting
)

func (s InstanceState) String() string {
	switch s {
	case StatePending:
		return "pending"
	case StateRunning:
		return "running"
	case StateStopped:
		return "stopped"
	case StateTerminated:
		return "terminated"
	case StateStopping:
		return "stopping"
	case StateStarting:
		return "starting"
	default:
		return "unknown"
	}
}

// InstanceSpec defines the configuration for creating a cloud instance.
type InstanceSpec struct {
	Name            string            `yaml:"name" json:"name"`
	Role            string            `yaml:"role" json:"role"`                         // db_primary, db_standby, app
	InstanceType    string            `yaml:"instance_type" json:"instance_type"`
	ImageID         string            `yaml:"ami" json:"image_id"`                      // AMI (AWS), Image (Azure), Image (OCI)
	SubnetID        string            `yaml:"subnet" json:"subnet_id"`
	SecurityGroupID string            `yaml:"security_group" json:"security_group_id"`
	KeyPairName     string            `yaml:"key_pair" json:"key_pair_name"`
	Tags            map[string]string `yaml:"tags" json:"tags"`
	Storage         []VolumeSpec      `yaml:"storage" json:"storage"`
	UserData        string            `yaml:"user_data" json:"user_data"` // Cloud-init / UserData script

	// Azure-specific
	ResourceGroup string `yaml:"resource_group,omitempty" json:"resource_group,omitempty"`
	VMSize        string `yaml:"vm_size,omitempty" json:"vm_size,omitempty"`

	// OCI-specific
	Shape         string `yaml:"shape,omitempty" json:"shape,omitempty"`
	OCPUs         int    `yaml:"ocpus,omitempty" json:"ocpus,omitempty"`
	MemoryGB      int    `yaml:"memory_gb,omitempty" json:"memory_gb,omitempty"`
	CompartmentID string `yaml:"compartment_id,omitempty" json:"compartment_id,omitempty"`
}

// Validate checks required fields.
func (s *InstanceSpec) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("instance name is required")
	}
	if s.InstanceType == "" && s.VMSize == "" && s.Shape == "" {
		return fmt.Errorf("instance type (instance_type, vm_size, or shape) is required")
	}
	if s.ImageID == "" {
		return fmt.Errorf("image ID is required")
	}
	return nil
}

// Instance represents a running or stopped cloud instance.
type Instance struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Provider     ProviderID        `json:"provider"`
	State        InstanceState     `json:"state"`
	InstanceType string            `json:"instance_type"`
	PrivateIP    string            `json:"private_ip"`
	PublicIP     string            `json:"public_ip,omitempty"`
	SubnetID     string            `json:"subnet_id"`
	LaunchTime   time.Time         `json:"launch_time"`
	Tags         map[string]string `json:"tags"`
	Volumes      []Volume          `json:"volumes,omitempty"`
}

// VolumeSpec defines the configuration for creating a storage volume.
type VolumeSpec struct {
	Name      string `yaml:"name" json:"name"`
	SizeGB    int    `yaml:"size_gb" json:"size_gb"`
	Type      string `yaml:"type" json:"type"` // gp3, io2, Premium_LRS, etc.
	IOPS      int    `yaml:"iops,omitempty" json:"iops,omitempty"`
	MountPath string `yaml:"mount,omitempty" json:"mount_path,omitempty"`
}

// Validate checks required fields.
func (v *VolumeSpec) Validate() error {
	if v.SizeGB <= 0 {
		return fmt.Errorf("volume size must be > 0 (got %d)", v.SizeGB)
	}
	if v.Type == "" {
		return fmt.Errorf("volume type is required")
	}
	return nil
}

// Volume represents a storage volume attached to an instance.
type Volume struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	SizeGB     int    `json:"size_gb"`
	Type       string `json:"type"`
	IOPS       int    `json:"iops,omitempty"`
	State      string `json:"state"`
	InstanceID string `json:"instance_id,omitempty"`
	MountPath  string `json:"mount_path,omitempty"`
}

// SGSpec defines a security group / NSG / security list.
type SGSpec struct {
	Name         string        `yaml:"name" json:"name"`
	Description  string        `yaml:"description" json:"description"`
	VPCID        string        `yaml:"vpc_id" json:"vpc_id"`
	IngressRules []IngressRule `yaml:"ingress" json:"ingress_rules"`
	EgressRules  []EgressRule  `yaml:"egress" json:"egress_rules"`
}

// IngressRule defines an inbound traffic rule.
type IngressRule struct {
	Port        int    `yaml:"port" json:"port"`
	Protocol    string `yaml:"protocol" json:"protocol"`       // tcp, udp, icmp
	Source      string `yaml:"source" json:"source"`           // CIDR
	Description string `yaml:"description" json:"description"`
}

// EgressRule defines an outbound traffic rule.
type EgressRule struct {
	Port        int    `yaml:"port" json:"port"`
	Protocol    string `yaml:"protocol" json:"protocol"`
	Destination string `yaml:"destination" json:"destination"` // CIDR
	Description string `yaml:"description" json:"description"`
}

// SecurityGroup represents a created security group.
type SecurityGroup struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	VPCID string `json:"vpc_id"`
}

// LBSpec defines a load balancer.
type LBSpec struct {
	Name      string       `yaml:"name" json:"name"`
	Type      string       `yaml:"type" json:"type"` // nlb, alb, azure_lb, oci_lb
	Listeners []LBListener `yaml:"listeners" json:"listeners"`
	SubnetIDs []string     `yaml:"subnet_ids" json:"subnet_ids"`
	Internal  bool         `yaml:"internal" json:"internal"`
}

// LBListener defines a load balancer listener.
type LBListener struct {
	Port       int    `yaml:"port" json:"port"`
	TargetPort int    `yaml:"target_port" json:"target_port"`
	Protocol   string `yaml:"protocol" json:"protocol"` // TCP, HTTP, HTTPS
}

// LoadBalancer represents a created load balancer.
type LoadBalancer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	DNSName string `json:"dns_name,omitempty"`
	Type    string `json:"type"`
	State   string `json:"state"`
}

// ManagedDBSpec defines a managed database service (RDS, Azure Flexible, Autonomous DB).
type ManagedDBSpec struct {
	Name                string `yaml:"name" json:"name"`
	Engine              string `yaml:"engine" json:"engine"`                   // oracle, postgres
	Version             string `yaml:"version" json:"version"`
	InstanceClass       string `yaml:"instance_class" json:"instance_class"`   // db.r6i.2xlarge, GP_Standard_D8s_v3, etc.
	StorageGB           int    `yaml:"storage_gb" json:"storage_gb"`
	HAMode              string `yaml:"ha_mode" json:"ha_mode"`                 // multi_az, zone_redundant, none
	BackupRetentionDays int    `yaml:"backup_retention_days" json:"backup_retention_days"`
	SubnetGroupID       string `yaml:"subnet_group_id" json:"subnet_group_id"`
	SecurityGroupID     string `yaml:"security_group_id" json:"security_group_id"`

	// Azure-specific
	SKU                string `yaml:"sku,omitempty" json:"sku,omitempty"`
	GeoRedundantBackup bool   `yaml:"geo_redundant_backup,omitempty" json:"geo_redundant_backup,omitempty"`

	// OCI-specific (Autonomous DB)
	WorkloadType string `yaml:"workload_type,omitempty" json:"workload_type,omitempty"` // OLTP, DW, APEX, JSON
	CPUCount     int    `yaml:"cpu_count,omitempty" json:"cpu_count,omitempty"`
	StorageTB    int    `yaml:"storage_tb,omitempty" json:"storage_tb,omitempty"`
	AutoScaling  bool   `yaml:"auto_scaling,omitempty" json:"auto_scaling,omitempty"`
	LicenseType  string `yaml:"license_type,omitempty" json:"license_type,omitempty"` // LICENSE_INCLUDED, BRING_YOUR_OWN
}

// ManagedDB represents a managed database instance.
type ManagedDB struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Engine   string `json:"engine"`
	Version  string `json:"version"`
	State    string `json:"state"`
	Endpoint string `json:"endpoint"`
	Port     int    `json:"port"`
}

// CostEstimate holds monthly cost breakdown.
type CostEstimate struct {
	ComputeMonthly   float64        `json:"compute_monthly"`
	StorageMonthly   float64        `json:"storage_monthly"`
	NetworkMonthly   float64        `json:"network_monthly"`
	ManagedDBMonthly float64        `json:"managed_db_monthly"`
	Currency         string         `json:"currency"`
	Details          []CostLineItem `json:"details,omitempty"`
}

// MonthlyTotal returns the sum of all cost components.
func (c *CostEstimate) MonthlyTotal() float64 {
	return c.ComputeMonthly + c.StorageMonthly + c.NetworkMonthly + c.ManagedDBMonthly
}

// CostLineItem is a single line item in a cost estimate.
type CostLineItem struct {
	Category    string  `json:"category"`    // compute, storage, network, managed_db
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Monthly     float64 `json:"monthly"`
}

// Snapshot represents a point-in-time backup of a volume or instance.
type Snapshot struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	VolumeID   string    `json:"volume_id"`
	InstanceID string    `json:"instance_id"`
	SizeGB     int       `json:"size_gb"`
	State      string    `json:"state"`
	CreatedAt  time.Time `json:"created_at"`
}

// CloudProvider is the core interface for all cloud infrastructure operations.
// Implementations: aws/, azure/, oci/, onprem/
type CloudProvider interface {
	// Instance lifecycle
	CreateInstance(ctx context.Context, spec InstanceSpec) (*Instance, error)
	TerminateInstance(ctx context.Context, instanceID string) error
	GetInstance(ctx context.Context, instanceID string) (*Instance, error)
	ListInstances(ctx context.Context, filters map[string]string) ([]*Instance, error)
	StartInstance(ctx context.Context, instanceID string) error
	StopInstance(ctx context.Context, instanceID string) error

	// Storage
	CreateVolume(ctx context.Context, spec VolumeSpec) (*Volume, error)
	AttachVolume(ctx context.Context, volumeID, instanceID string) error
	DetachVolume(ctx context.Context, volumeID string) error

	// Network
	CreateSecurityGroup(ctx context.Context, spec SGSpec) (*SecurityGroup, error)
	AuthorizeIngress(ctx context.Context, sgID string, rule IngressRule) error

	// Load Balancer
	CreateLoadBalancer(ctx context.Context, spec LBSpec) (*LoadBalancer, error)
	RegisterTarget(ctx context.Context, lbID, instanceID string, port int) error

	// Managed Database (optional -- not all providers/configs use this)
	CreateManagedDB(ctx context.Context, spec ManagedDBSpec) (*ManagedDB, error)
	GetManagedDB(ctx context.Context, dbID string) (*ManagedDB, error)

	// Cost Estimation
	EstimateCost(ctx context.Context, spec InstanceSpec) (*CostEstimate, error)
}

// ErrNotImplemented returns an error for methods not yet wired to real cloud SDKs.
func ErrNotImplemented(method string) error {
	return fmt.Errorf("%s: not yet implemented (mock-only)", method)
}
