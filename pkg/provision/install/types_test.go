package install

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallSpec_Validate(t *testing.T) {
	cases := []struct {
		name    string
		spec    InstallSpec
		wantErr string // empty string = no error expected
	}{
		{"missing target", InstallSpec{}, "target is required"},
		{"missing oracle_home", InstallSpec{Target: "ext3adm1"}, "oracle_home is required"},
		{"valid spec", InstallSpec{Target: "ext3adm1", OracleHome: "/u01/app/19c/grid"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate()
			if tc.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			}
		})
	}
}

func TestDetectionState_String(t *testing.T) {
	cases := []struct {
		state DetectionState
		want  string
	}{
		{DetectionStateAbsent, "absent"},
		{DetectionStatePartial, "partial"},
		{DetectionStateInstalled, "installed"},
		{DetectionState(99), "unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.state.String())
		})
	}
}
