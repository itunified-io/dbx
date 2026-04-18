package cloud

import (
	"context"
	"fmt"
	"strings"
)

// EstimateBlueprint computes the total monthly cost for a blueprint by delegating
// to each instance's provider-specific EstimateCost method and summing results.
func EstimateBlueprint(ctx context.Context, bp *Blueprint, reg *ProviderRegistry) (*CostEstimate, error) {
	providerID, err := ParseProviderID(bp.Metadata.Provider)
	if err != nil {
		return nil, fmt.Errorf("blueprint provider: %w", err)
	}

	provider, err := reg.Get(providerID)
	if err != nil {
		return nil, fmt.Errorf("provider for %s: %w", bp.Metadata.Provider, err)
	}

	total := &CostEstimate{Currency: "USD"}

	for _, inst := range bp.Infrastructure.Instances {
		spec := InstanceSpec{
			Name:         inst.Name,
			InstanceType: inst.InstanceType,
			ImageID:      inst.AMI,
			Storage:      inst.Storage,
			// Azure/OCI fields
			VMSize:   inst.VMSize,
			Shape:    inst.Shape,
			OCPUs:    inst.OCPUs,
			MemoryGB: inst.MemoryGB,
		}

		est, err := provider.EstimateCost(ctx, spec)
		if err != nil {
			return nil, fmt.Errorf("estimate cost for %s: %w", inst.Name, err)
		}

		total.ComputeMonthly += est.ComputeMonthly
		total.StorageMonthly += est.StorageMonthly
		total.NetworkMonthly += est.NetworkMonthly
		total.ManagedDBMonthly += est.ManagedDBMonthly
		total.Details = append(total.Details, est.Details...)
	}

	// Add LB estimated cost if present
	if bp.Infrastructure.LoadBalancer != nil {
		lbCost := 16.20 // Base NLB/LB cost estimate
		total.NetworkMonthly += lbCost
		total.Details = append(total.Details, CostLineItem{
			Category:    "network",
			Description: bp.Infrastructure.LoadBalancer.Type + " base cost",
			Monthly:     lbCost,
		})
	}

	return total, nil
}

// FormatCostReport produces a human-readable cost breakdown string.
func FormatCostReport(est *CostEstimate, provider string, dbxLicenseMonthly float64) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("%s Monthly Cost Estimate:\n", strings.ToUpper(provider)))
	b.WriteString(fmt.Sprintf("  Compute:    %s %s\n", est.Currency, formatMoney(est.ComputeMonthly)))
	b.WriteString(fmt.Sprintf("  Storage:    %s %s\n", est.Currency, formatMoney(est.StorageMonthly)))
	b.WriteString(fmt.Sprintf("  Network:    %s %s\n", est.Currency, formatMoney(est.NetworkMonthly)))
	if est.ManagedDBMonthly > 0 {
		b.WriteString(fmt.Sprintf("  Managed DB: %s %s\n", est.Currency, formatMoney(est.ManagedDBMonthly)))
	}
	cloudTotal := est.MonthlyTotal()
	b.WriteString(fmt.Sprintf("  Total %s:  %s %s/mo\n", strings.ToUpper(provider), est.Currency, formatMoney(cloudTotal)))

	if dbxLicenseMonthly > 0 {
		b.WriteString(fmt.Sprintf("  dbx License: %s %s/mo\n", est.Currency, formatMoney(dbxLicenseMonthly)))
		b.WriteString(fmt.Sprintf("  Grand Total: %s %s/mo\n", est.Currency, formatMoney(cloudTotal+dbxLicenseMonthly)))
	}

	if len(est.Details) > 0 {
		b.WriteString("\nBreakdown:\n")
		for _, item := range est.Details {
			b.WriteString(fmt.Sprintf("  [%s] %s: %s %s\n", item.Category, item.Description, est.Currency, formatMoney(item.Monthly)))
		}
	}

	return b.String()
}

// formatMoney formats a float as a comma-separated money string (e.g., "1,459.20").
func formatMoney(amount float64) string {
	s := fmt.Sprintf("%.2f", amount)
	parts := strings.Split(s, ".")
	intPart := parts[0]

	// Insert commas
	n := len(intPart)
	if n <= 3 {
		return s
	}
	var result strings.Builder
	for i, c := range intPart {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}
	return result.String() + "." + parts[1]
}
