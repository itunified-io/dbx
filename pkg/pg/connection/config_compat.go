package connection

import (
	"encoding/json"
	"fmt"

	"github.com/itunified-io/dbx/pkg/core/target"
)

// CompatConfig is the JSON-serializable form of a multi-profile connection config.
type CompatConfig struct {
	Profiles []CompatProfile `json:"profiles"`
	Active   string          `json:"active"`
}

// CompatProfile is the JSON-serializable form of a single profile.
type CompatProfile struct {
	Name   string        `json:"name"`
	Config ProfileConfig `json:"config"`
}

// GenerateConfigJSON derives a CompatConfig from a Target and returns its
// JSON representation. It creates a "primary" profile from Target.Primary
// and an optional "replica" profile from Target.Replica.
func GenerateConfigJSON(t *target.Target) ([]byte, error) {
	if t.Primary == nil {
		return nil, fmt.Errorf("target %q has no primary endpoint", t.Name)
	}

	cc := CompatConfig{Active: "primary"}

	cc.Profiles = append(cc.Profiles, CompatProfile{
		Name:   "primary",
		Config: endpointToConfig(t.Primary),
	})

	if t.Replica != nil {
		cc.Profiles = append(cc.Profiles, CompatProfile{
			Name:   "replica",
			Config: endpointToConfig(t.Replica),
		})
	}

	return json.MarshalIndent(cc, "", "  ")
}

// endpointToConfig converts a target.Endpoint to a ProfileConfig.
func endpointToConfig(ep *target.Endpoint) ProfileConfig {
	port := ep.Port
	if port == 0 {
		port = 5432
	}
	sslMode := ep.SSLMode
	if sslMode == "" {
		sslMode = "prefer"
	}
	return ProfileConfig{
		Host:     ep.Host,
		Port:     port,
		Database: ep.Database,
		SSLMode:  sslMode,
	}
}
