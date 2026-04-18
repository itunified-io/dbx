package cloud

import "fmt"

// Recommendation holds an instance type recommendation for a workload profile.
type Recommendation struct {
	Profile      string     `json:"profile"`
	Provider     ProviderID `json:"provider"`
	InstanceType string     `json:"instance_type"`
	VCPUs        int        `json:"vcpus"`
	MemoryGB     int        `json:"memory_gb"`
	StorageGB    int        `json:"storage_gb"`
	StorageType  string     `json:"storage_type"`
	IOPS         int        `json:"iops"`
	Description  string     `json:"description"`
}

// workloadProfile defines recommended specs per provider.
type workloadProfile struct {
	Description string
	AWS         providerSpec
	Azure       providerSpec
	OCI         providerSpec
}

type providerSpec struct {
	InstanceType string
	VCPUs        int
	MemoryGB     int
	StorageGB    int
	StorageType  string
	IOPS         int
}

var profiles = map[string]workloadProfile{
	"oracle_oltp_small": {
		Description: "Oracle OLTP < 100 sessions",
		AWS:         providerSpec{"r6i.xlarge", 4, 32, 200, "io2", 5000},
		Azure:       providerSpec{"Standard_E4s_v5", 4, 32, 200, "Premium_LRS", 5000},
		OCI:         providerSpec{"VM.Standard.E4.Flex", 4, 64, 200, "block", 5000},
	},
	"oracle_oltp_medium": {
		Description: "Oracle OLTP 100-500 sessions",
		AWS:         providerSpec{"r6i.2xlarge", 8, 64, 500, "io2", 10000},
		Azure:       providerSpec{"Standard_E8s_v5", 8, 64, 500, "Premium_LRS", 10000},
		OCI:         providerSpec{"VM.Standard.E4.Flex", 8, 128, 500, "block", 10000},
	},
	"oracle_oltp_large": {
		Description: "Oracle OLTP 500+ sessions",
		AWS:         providerSpec{"r6i.4xlarge", 16, 128, 1024, "io2", 20000},
		Azure:       providerSpec{"Standard_E16s_v5", 16, 128, 1024, "Premium_LRS", 20000},
		OCI:         providerSpec{"VM.Standard.E4.Flex", 16, 256, 1024, "block", 20000},
	},
	"pg_oltp_small": {
		Description: "PostgreSQL OLTP < 200 connections",
		AWS:         providerSpec{"r6i.xlarge", 4, 32, 200, "gp3", 3000},
		Azure:       providerSpec{"Standard_E4s_v5", 4, 32, 200, "Premium_LRS", 3000},
		OCI:         providerSpec{"VM.Standard.E4.Flex", 4, 64, 200, "block", 3000},
	},
	"pg_oltp_large": {
		Description: "PostgreSQL OLTP 200+ connections",
		AWS:         providerSpec{"r6i.4xlarge", 16, 128, 1024, "io2", 20000},
		Azure:       providerSpec{"Standard_E16s_v5", 16, 128, 1024, "Premium_LRS", 20000},
		OCI:         providerSpec{"VM.Standard.E4.Flex", 16, 256, 1024, "block", 20000},
	},
}

// Recommend returns instance type recommendation for a workload profile and provider.
func Recommend(profileName string, provider ProviderID) (*Recommendation, error) {
	profile, ok := profiles[profileName]
	if !ok {
		return nil, fmt.Errorf("unknown workload profile: %q", profileName)
	}

	var spec providerSpec
	switch provider {
	case AWS:
		spec = profile.AWS
	case Azure:
		spec = profile.Azure
	case OCI:
		spec = profile.OCI
	default:
		return nil, fmt.Errorf("recommendations not available for provider %s", provider)
	}

	return &Recommendation{
		Profile:      profileName,
		Provider:     provider,
		InstanceType: spec.InstanceType,
		VCPUs:        spec.VCPUs,
		MemoryGB:     spec.MemoryGB,
		StorageGB:    spec.StorageGB,
		StorageType:  spec.StorageType,
		IOPS:         spec.IOPS,
		Description:  profile.Description,
	}, nil
}

// ListWorkloadProfiles returns all available workload profile names.
func ListWorkloadProfiles() []string {
	names := make([]string, 0, len(profiles))
	for k := range profiles {
		names = append(names, k)
	}
	return names
}
