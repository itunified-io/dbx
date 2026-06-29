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

// ADR-0047: an EchoBack op must NOT proceed on the boolean alone — the caller
// must restate the target identifier. A bare ConfirmFlag=true is not enough.
func TestPipelineEchoBackBooleanOnlyBlocks(t *testing.T) {
	p := pipeline.New(pipeline.WithConfirmGate(confirm.New(nil, nil)))

	executed := false
	req := &pipeline.Request{
		Interface:    "mcp",
		User:         "admin",
		ToolName:     "pg_drop_table",
		Target:       &target.Target{Name: "users"},
		ConfirmLevel: confirm.LevelEchoBack,
		ConfirmFlag:  true, // boolean only, no restatement
	}
	_, err := p.Execute(context.Background(), req, func(ctx context.Context, r *pipeline.Request) (any, error) {
		executed = true
		return nil, nil
	})
	assert.ErrorIs(t, err, confirm.ErrConfirmRequired)
	assert.False(t, executed, "execFn must not run when confirmation is missing")
}

func TestPipelineEchoBackWrongIdentifierBlocks(t *testing.T) {
	p := pipeline.New(pipeline.WithConfirmGate(confirm.New(nil, nil)))

	executed := false
	req := &pipeline.Request{
		Interface:    "mcp",
		User:         "admin",
		ToolName:     "pg_drop_table",
		Target:       &target.Target{Name: "users"},
		ConfirmLevel: confirm.LevelEchoBack,
		ConfirmFlag:  true,
		Params:       map[string]string{"confirm_target": "orders"}, // wrong table
	}
	_, err := p.Execute(context.Background(), req, func(ctx context.Context, r *pipeline.Request) (any, error) {
		executed = true
		return nil, nil
	})
	assert.ErrorIs(t, err, confirm.ErrConfirmMismatch)
	assert.False(t, executed)
}

func TestPipelineEchoBackCorrectIdentifierProceeds(t *testing.T) {
	p := pipeline.New(pipeline.WithConfirmGate(confirm.New(nil, nil)))

	req := &pipeline.Request{
		Interface:    "mcp",
		User:         "admin",
		ToolName:     "pg_drop_table",
		Target:       &target.Target{Name: "users"},
		ConfirmLevel: confirm.LevelEchoBack,
		Params:       map[string]string{"confirm_target": "users"}, // restated identifier
	}
	resp, err := p.Execute(context.Background(), req, func(ctx context.Context, r *pipeline.Request) (any, error) {
		return "dropped", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "dropped", resp.Result)
}

// ADR-0047: a DoubleConfirm op requires both the restated identifier and the
// derived second phrase. Boolean alone, or a wrong factor, must block.
func TestPipelineDoubleConfirmBooleanOnlyBlocks(t *testing.T) {
	p := pipeline.New(pipeline.WithConfirmGate(confirm.New(nil, nil)))

	req := &pipeline.Request{
		Interface:    "mcp",
		User:         "admin",
		ToolName:     "cnpg_cluster_delete",
		Target:       &target.Target{Name: "prod-cluster"},
		ConfirmLevel: confirm.LevelDoubleConfirm,
		ConfirmFlag:  true,
	}
	_, err := p.Execute(context.Background(), req, func(ctx context.Context, r *pipeline.Request) (any, error) {
		return nil, nil
	})
	assert.ErrorIs(t, err, confirm.ErrConfirmRequired)
}

func TestPipelineDoubleConfirmWrongSecondFactorBlocks(t *testing.T) {
	p := pipeline.New(pipeline.WithConfirmGate(confirm.New(nil, nil)))

	req := &pipeline.Request{
		Interface:    "mcp",
		User:         "admin",
		ToolName:     "cnpg_cluster_delete",
		Target:       &target.Target{Name: "prod-cluster"},
		ConfirmLevel: confirm.LevelDoubleConfirm,
		Params: map[string]string{
			"confirm_target": "prod-cluster",
			"confirm_phrase": "WRONG",
		},
	}
	_, err := p.Execute(context.Background(), req, func(ctx context.Context, r *pipeline.Request) (any, error) {
		return nil, nil
	})
	assert.ErrorIs(t, err, confirm.ErrConfirmMismatch)
}

func TestPipelineDoubleConfirmCorrectFactorsProceeds(t *testing.T) {
	p := pipeline.New(pipeline.WithConfirmGate(confirm.New(nil, nil)))

	req := &pipeline.Request{
		Interface:    "mcp",
		User:         "admin",
		ToolName:     "cnpg_cluster_delete",
		Target:       &target.Target{Name: "prod-cluster"},
		ConfirmLevel: confirm.LevelDoubleConfirm,
		Params: map[string]string{
			"confirm_target": "prod-cluster",
			"confirm_phrase": "CONFIRM-prod-cluster",
		},
	}
	resp, err := p.Execute(context.Background(), req, func(ctx context.Context, r *pipeline.Request) (any, error) {
		return "deleted", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "deleted", resp.Result)
}
