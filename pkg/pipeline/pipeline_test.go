package pipeline_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/core/audit"
	"github.com/itunified-io/dbx/pkg/core/confirm"
	"github.com/itunified-io/dbx/pkg/core/oraclegate"
	"github.com/itunified-io/dbx/pkg/core/target"
	"github.com/itunified-io/dbx/pkg/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelineBasicExecution(t *testing.T) {
	auditBuf := &bytes.Buffer{}
	p := pipeline.New(
		pipeline.WithAuditLogger(audit.NewLogger(audit.WithStdoutSink(auditBuf))),
	)

	req := &pipeline.Request{
		Interface: "cli",
		User:      "admin",
		ToolName:  "target_list",
		Params:    map[string]string{"entity_type": "oracle_database"},
	}

	resp, err := p.Execute(context.Background(), req, func(ctx context.Context, r *pipeline.Request) (any, error) {
		return []string{"prod-orcl", "dev-orcl"}, nil
	})
	require.NoError(t, err)
	assert.NotNil(t, resp.Result)
	assert.NotEmpty(t, auditBuf.String())
}

func TestPipelineBlocksWithoutToolName(t *testing.T) {
	p := pipeline.New()

	req := &pipeline.Request{
		Interface: "cli",
		User:      "admin",
	}

	_, err := p.Execute(context.Background(), req, func(ctx context.Context, r *pipeline.Request) (any, error) {
		return nil, nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool name is required")
}

func TestPipelineOracleGateBlocks(t *testing.T) {
	p := pipeline.New(
		pipeline.WithOracleGate(oraclegate.New("strict")),
	)

	req := &pipeline.Request{
		Interface: "cli",
		User:      "admin",
		ToolName:  "oracle_ash_report",
		Target: &target.Target{
			Name: "prod-orcl",
			Type: target.TypeOracleDatabase,
			OracleLicense: &target.OracleLicense{
				Edition: "standard2",
			},
		},
		OracleReq: &oraclegate.Requirement{
			Edition: "enterprise",
			Options: []string{"diagnostics_pack"},
		},
	}

	_, err := p.Execute(context.Background(), req, func(ctx context.Context, r *pipeline.Request) (any, error) {
		return nil, nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "oracle license gate")
}

func TestPipelineConfirmGateBlocks(t *testing.T) {
	p := pipeline.New(
		pipeline.WithConfirmGate(confirm.New(nil, nil)),
	)

	req := &pipeline.Request{
		Interface:    "cli",
		User:         "admin",
		ToolName:     "dg_switchover",
		ConfirmLevel: confirm.LevelStandard,
		ConfirmFlag:  false,
	}

	_, err := p.Execute(context.Background(), req, func(ctx context.Context, r *pipeline.Request) (any, error) {
		return nil, nil
	})
	assert.ErrorIs(t, err, confirm.ErrConfirmRequired)
}

func TestPipelineConfirmFlagBypasses(t *testing.T) {
	p := pipeline.New(
		pipeline.WithConfirmGate(confirm.New(nil, nil)),
	)

	req := &pipeline.Request{
		Interface:    "cli",
		User:         "admin",
		ToolName:     "dg_switchover",
		ConfirmLevel: confirm.LevelStandard,
		ConfirmFlag:  true,
	}

	resp, err := p.Execute(context.Background(), req, func(ctx context.Context, r *pipeline.Request) (any, error) {
		return "switchover complete", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "switchover complete", resp.Result)
}
