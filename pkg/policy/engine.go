package policy

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EngineOpts configures the policy engine.
type EngineOpts struct {
	Concurrency int // max parallel rule evaluations (default 4)
}

// Engine evaluates policies against targets.
type Engine struct {
	executors   map[string]CheckExecutor
	concurrency int
}

// NewEngine creates a new policy engine.
func NewEngine(opts EngineOpts) *Engine {
	c := opts.Concurrency
	if c <= 0 {
		c = 4
	}
	return &Engine{
		executors:   make(map[string]CheckExecutor),
		concurrency: c,
	}
}

// RegisterExecutor registers a check executor for a check type.
func (e *Engine) RegisterExecutor(checkType string, exec CheckExecutor) {
	e.executors[checkType] = exec
}

// Scan evaluates all rules in a policy against a target.
func (e *Engine) Scan(ctx context.Context, entityName, entityType string, p *Policy) (*ScanResult, error) {
	start := time.Now()

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("policy scan: %w", err)
	}

	results := make([]CheckResult, len(p.Rules))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, e.concurrency)

	var scanErr error

	for i, rule := range p.Rules {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("policy scan: %w", ctx.Err())
		default:
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, r Rule) {
			defer wg.Done()
			defer func() { <-sem }()

			cr := e.evaluateRule(ctx, r)
			mu.Lock()
			results[idx] = cr
			mu.Unlock()
		}(i, rule)
	}

	wg.Wait()

	if scanErr != nil {
		return nil, scanErr
	}

	summary := ScanSummary{Total: len(results)}
	for _, cr := range results {
		switch cr.Status {
		case "pass":
			summary.Passed++
		case "fail":
			summary.Failed++
		case "error":
			summary.Errors++
		case "skip":
			summary.Skipped++
		}
	}

	return &ScanResult{
		EntityName: entityName,
		EntityType: entityType,
		PolicyName: p.Metadata.Name,
		Framework:  p.Metadata.Framework,
		PolicySHA:  p.SHA256,
		ScanTime:   start,
		Duration:   time.Since(start),
		Results:    results,
		Summary:    summary,
	}, nil
}

func (e *Engine) evaluateRule(ctx context.Context, r Rule) CheckResult {
	exec, ok := e.executors[r.Check.Type]
	if !ok {
		return CheckResult{
			RuleID:      r.ID,
			Title:       r.Title,
			Severity:    r.Severity,
			Status:      "error",
			Message:     fmt.Sprintf("no executor for check type %q", r.Check.Type),
			EvaluatedAt: time.Now(),
		}
	}

	cr, err := exec.Execute(ctx, r.Check)
	if err != nil {
		return CheckResult{
			RuleID:      r.ID,
			Title:       r.Title,
			Severity:    r.Severity,
			Status:      "error",
			Message:     err.Error(),
			EvaluatedAt: time.Now(),
		}
	}

	cr.RuleID = r.ID
	cr.Title = r.Title
	cr.Severity = r.Severity
	return cr
}
