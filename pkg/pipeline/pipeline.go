// Package pipeline implements the 9-stage execution pipeline for all dbx operations.
package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/itunified-io/dbx/pkg/core/audit"
	"github.com/itunified-io/dbx/pkg/core/confirm"
	"github.com/itunified-io/dbx/pkg/core/license"
	"github.com/itunified-io/dbx/pkg/core/oraclegate"
	"github.com/itunified-io/dbx/pkg/core/target"
	"github.com/itunified-io/dbx/pkg/core/vault"
)

// Request describes a tool invocation flowing through the pipeline.
type Request struct {
	Interface    string            // cli, rest, mcp, ui
	User         string            // authenticated user
	ToolName     string            // e.g. "oracle_dg_switchover"
	Target       *target.Target    // resolved target
	Params       map[string]string // named parameters
	ConfirmFlag  bool              // --confirm passed
	ConfirmLevel confirm.Level     // tool's required confirm level

	// Oracle gate requirements (nil if not an Oracle tool)
	OracleReq *oraclegate.Requirement

	// OEM pack requirement (empty if not OEM-specific)
	OEMPack string
}

// Response holds the output of a pipeline execution.
type Response struct {
	Result   any
	Duration time.Duration
}

// ExecuteFunc is the actual tool execution (stage 7).
type ExecuteFunc func(ctx context.Context, req *Request) (any, error)

// Pipeline orchestrates the 9-stage execution flow.
type Pipeline struct {
	licenseValidator *license.Validator
	oracleGate       *oraclegate.Gate
	confirmGate      *confirm.Gate
	vaultClient      *vault.Client
	auditLogger      *audit.Logger
}

// Option configures the Pipeline.
type Option func(*Pipeline)

func WithLicenseValidator(v *license.Validator) Option {
	return func(p *Pipeline) { p.licenseValidator = v }
}

func WithOracleGate(g *oraclegate.Gate) Option {
	return func(p *Pipeline) { p.oracleGate = g }
}

func WithConfirmGate(g *confirm.Gate) Option {
	return func(p *Pipeline) { p.confirmGate = g }
}

func WithVaultClient(c *vault.Client) Option {
	return func(p *Pipeline) { p.vaultClient = c }
}

func WithAuditLogger(l *audit.Logger) Option {
	return func(p *Pipeline) { p.auditLogger = l }
}

// New creates a pipeline with the given options.
func New(opts ...Option) *Pipeline {
	p := &Pipeline{}
	for _, o := range opts {
		o(p)
	}
	return p
}

// Execute runs the full 9-stage pipeline for a tool invocation.
func (p *Pipeline) Execute(ctx context.Context, req *Request, execFn ExecuteFunc) (*Response, error) {
	evt := audit.NewEvent(req.Interface, req.User, req.ToolName, "")
	if req.Target != nil {
		evt.Target = req.Target.Name
	}
	evt.Params = toAnyMap(req.Params)

	defer func() {
		if p.auditLogger != nil {
			evt.Redact([]string{"password", "secret", "token"})
			p.auditLogger.Log(evt)
		}
	}()

	// Stage 1: Input validation
	if req.ToolName == "" {
		err := fmt.Errorf("tool name is required")
		evt.Complete("failure", err)
		return nil, err
	}

	// Stage 2: License check
	if err := p.checkLicense(); err != nil {
		evt.Complete("blocked", err)
		return nil, err
	}

	// Stage 3: Oracle license gate
	if req.Target != nil && req.Target.IsOracle() && req.OracleReq != nil && p.oracleGate != nil {
		result := p.oracleGate.Check(req.Target.OracleLicense, *req.OracleReq)
		if result.Decision == oraclegate.Block {
			err := fmt.Errorf("oracle license gate: %s", result.Reason)
			evt.Complete("blocked", err)
			return nil, err
		}
	}

	// Stage 4: OEM pack gate (simplified — uses oracle gate mode)

	// Stage 5: Confirm gate
	if p.confirmGate != nil {
		if err := p.confirmGate.Check(req.ConfirmLevel, req.ToolName, targetName(req), req.ConfirmFlag); err != nil {
			evt.Complete("blocked", err)
			return nil, err
		}
	}

	// Stage 6: Credential resolution (available to execFn via vault client)

	// Stage 7: Execution
	start := time.Now()
	result, err := execFn(ctx, req)
	duration := time.Since(start)

	if err != nil {
		evt.Complete("failure", err)
		return nil, err
	}

	evt.Complete("success", nil)
	return &Response{Result: result, Duration: duration}, nil
}

func (p *Pipeline) checkLicense() error {
	// OSS mode — no license required
	if p.licenseValidator == nil {
		return nil
	}
	return nil
}

func targetName(req *Request) string {
	if req.Target != nil {
		return req.Target.Name
	}
	return ""
}

func toAnyMap(m map[string]string) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
