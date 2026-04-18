package policy

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
)

// ReportJSON generates a JSON report from a scan result.
func ReportJSON(sr *ScanResult) ([]byte, error) {
	return json.MarshalIndent(sr, "", "  ")
}

// ReportCSV generates a CSV report from a scan result.
func ReportCSV(sr *ScanResult) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"rule_id", "title", "severity", "status", "actual", "expected", "message"})
	for _, r := range sr.Results {
		_ = w.Write([]string{r.RuleID, r.Title, r.Severity, r.Status, r.Actual, r.Expected, r.Message})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

// FleetReportCSV generates a CSV report aggregating multiple scan results.
func FleetReportCSV(results []*ScanResult) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"entity_name", "entity_type", "policy", "framework", "rule_id", "title", "severity", "status", "actual", "expected"})
	for _, sr := range results {
		for _, r := range sr.Results {
			_ = w.Write([]string{sr.EntityName, sr.EntityType, sr.PolicyName, sr.Framework, r.RuleID, r.Title, r.Severity, r.Status, r.Actual, r.Expected})
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

var htmlTemplate = template.Must(template.New("report").Funcs(template.FuncMap{
	"upper": strings.ToUpper,
}).Parse(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Policy Report: {{.PolicyName}}</title>
<style>
  body { font-family: sans-serif; margin: 2em; }
  table { border-collapse: collapse; width: 100%; }
  th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
  th { background: #333; color: #fff; }
  .pass { color: #2e7d32; font-weight: bold; }
  .fail { color: #c62828; font-weight: bold; }
  .error { color: #e65100; font-weight: bold; }
  .skip { color: #9e9e9e; }
  .critical { background: #ffcdd2; }
  .high { background: #fff3e0; }
  .summary { display: flex; gap: 2em; margin: 1em 0; }
  .summary div { padding: 1em; border-radius: 4px; }
</style></head><body>
<h1>{{.PolicyName}}</h1>
<p>Entity: <strong>{{.EntityName}}</strong> ({{.EntityType}}) | Framework: {{.Framework}} | SHA: {{.PolicySHA}}</p>
<p>Scanned: {{.ScanTime.Format "2006-01-02 15:04:05 UTC"}} | Duration: {{.Duration}}</p>
<div class="summary">
  <div style="background:#e8f5e9">Passed: {{.Summary.Passed}}</div>
  <div style="background:#ffebee">Failed: {{.Summary.Failed}}</div>
  <div style="background:#fff3e0">Errors: {{.Summary.Errors}}</div>
  <div style="background:#f5f5f5">Skipped: {{.Summary.Skipped}}</div>
  <div style="background:#e3f2fd">Total: {{.Summary.Total}}</div>
</div>
<table>
<tr><th>Rule ID</th><th>Title</th><th>Severity</th><th>Status</th><th>Actual</th><th>Expected</th><th>Message</th></tr>
{{range .Results}}
<tr class="{{.Severity}}"><td>{{.RuleID}}</td><td>{{.Title}}</td><td>{{.Severity}}</td><td class="{{.Status}}">{{upper .Status}}</td><td>{{.Actual}}</td><td>{{.Expected}}</td><td>{{.Message}}</td></tr>
{{end}}
</table></body></html>`))

// ReportHTML generates an HTML report from a scan result.
func ReportHTML(sr *ScanResult) ([]byte, error) {
	var buf bytes.Buffer
	if err := htmlTemplate.Execute(&buf, sr); err != nil {
		return nil, fmt.Errorf("html report: %w", err)
	}
	return buf.Bytes(), nil
}

// ComplianceScore computes 0-100 compliance percentage from a scan result.
func ComplianceScore(r *ScanResult) float64 {
	applicable := r.Summary.Total - r.Summary.Skipped
	if applicable == 0 {
		return 100.0
	}
	return float64(r.Summary.Passed) / float64(applicable) * 100
}
