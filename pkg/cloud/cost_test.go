package cloud_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEstimateBlueprint_AWS(t *testing.T) {
	bp, err := cloud.ParseBlueprint([]byte(validAWSBlueprint))
	require.NoError(t, err)

	// Wire up mock AWS provider for cost estimation
	reg := cloud.NewProviderRegistry()
	reg.Register(cloud.AWS, &mockProvider{id: cloud.AWS})

	est, err := cloud.EstimateBlueprint(context.Background(), bp, reg)
	require.NoError(t, err)

	// mockProvider.EstimateCost returns ComputeMonthly=100.0 per instance
	// 2 instances = 200.0 compute + 16.20 LB network
	assert.Greater(t, est.ComputeMonthly, 0.0)
	assert.Greater(t, est.MonthlyTotal(), 0.0)
	assert.Equal(t, "USD", est.Currency)
}

func TestEstimateBlueprint_UnknownProvider(t *testing.T) {
	bp := &cloud.Blueprint{
		Metadata: cloud.BlueprintMetadata{
			Name:     "test",
			Provider: "gcp",
			Profile:  "prod",
		},
	}
	reg := cloud.NewProviderRegistry()

	_, err := cloud.EstimateBlueprint(context.Background(), bp, reg)
	assert.Error(t, err)
}

func TestFormatCostReport(t *testing.T) {
	est := &cloud.CostEstimate{
		ComputeMonthly: 1459.20,
		StorageMonthly: 987.60,
		NetworkMonthly: 66.20,
		Currency:       "USD",
		Details: []cloud.CostLineItem{
			{Category: "compute", Description: "2x r6i.2xlarge (on-demand)", Monthly: 1459.20},
			{Category: "storage", Description: "2x 500GB io2 + misc", Monthly: 987.60},
			{Category: "network", Description: "NLB + data transfer", Monthly: 66.20},
		},
	}
	report := cloud.FormatCostReport(est, "aws", 83.00)
	assert.Contains(t, report, "Compute")
	assert.Contains(t, report, "1,459.20")
	assert.Contains(t, report, "Grand Total")
	assert.Contains(t, report, "dbx License")
}
