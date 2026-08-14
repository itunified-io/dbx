package root

import (
	"fmt"

	"github.com/spf13/cobra"
)

// resolveTarget returns the leaf-local --target if set, else the inherited
// parent persistent --target, else a usage error. The error is intended to
// surface BEFORE Validate so the caller sees a cobra-style "required flag"
// message instead of a deeper runtime spec.Validate() failure.
//
// Lived in provision_install.go until the install primitives moved to openfpp.
// It stayed because `db sql` uses it too — the helper was never specific to
// provisioning, it just happened to be written there first.
func resolveTarget(cmd *cobra.Command, localTarget string) (string, error) {
	if localTarget != "" {
		return localTarget, nil
	}
	if p := cmd.InheritedFlags().Lookup("target"); p != nil && p.Value.String() != "" {
		return p.Value.String(), nil
	}
	return "", fmt.Errorf("required flag --target not set (pass on leaf or on parent command)")
}
