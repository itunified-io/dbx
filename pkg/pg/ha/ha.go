// Package ha provides PostgreSQL high availability operations.
package ha

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// FailoverResult represents the result of a failover operation.
type FailoverResult struct {
	OldPrimary string `json:"old_primary"`
	NewPrimary string `json:"new_primary"`
	Duration   string `json:"duration"`
}

// SwitchoverResult represents the result of a switchover operation.
type SwitchoverResult struct {
	OldPrimary string `json:"old_primary"`
	NewPrimary string `json:"new_primary"`
	Duration   string `json:"duration"`
}

// ReplicaInfo represents streaming replication replica info.
type ReplicaInfo struct {
	ClientAddr     string     `json:"client_addr"`
	State          string     `json:"state"`
	SentLSN        string     `json:"sent_lsn"`
	ReplayLSN      string     `json:"replay_lsn"`
	ReplayLagBytes int64      `json:"replay_lag_bytes"`
	SyncState      string     `json:"sync_state"`
	ReplyTime      *time.Time `json:"reply_time"`
}

// Slot represents a replication slot.
type Slot struct {
	SlotName      string `json:"slot_name"`
	SlotType      string `json:"slot_type"`
	Active        bool   `json:"active"`
	RestartLSN    string `json:"restart_lsn"`
	RetainedBytes int64  `json:"retained_bytes"`
}

// ReadinessReport represents HA readiness assessment.
type ReadinessReport struct {
	Score       int      `json:"score"`       // 0-100
	ReplicaLag  int64    `json:"replica_lag_bytes"`
	ReplicaSync string   `json:"replica_sync_state"`
	Issues      []string `json:"issues"`
}

// SwitchoverPlanResult represents a planned switchover.
type SwitchoverPlanResult struct {
	Steps     []string `json:"steps"`
	Risks     []string `json:"risks"`
	Estimated string   `json:"estimated_duration"`
}

// LagInfo represents replica lag information.
type LagInfo struct {
	ClientAddr string `json:"client_addr"`
	LagBytes   int64  `json:"lag_bytes"`
	LagSize    string `json:"lag_size"`
}

// TimelineInfo represents WAL timeline information.
type TimelineInfo struct {
	TimelineID int64  `json:"timeline_id"`
	LSN        string `json:"lsn"`
	IsRecovery bool   `json:"is_in_recovery"`
}

const sqlReplicationStatus = `
SELECT client_addr::text, state, sent_lsn::text, replay_lsn::text,
       pg_wal_lsn_diff(sent_lsn, replay_lsn) AS replay_lag_bytes,
       sync_state, reply_time
FROM pg_stat_replication ORDER BY client_addr`

const sqlSlotList = `
SELECT slot_name, slot_type, active,
       COALESCE(restart_lsn::text, '') AS restart_lsn,
       pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn) AS retained_bytes
FROM pg_replication_slots ORDER BY slot_name`

const sqlReplicaLag = `
SELECT client_addr::text,
       pg_wal_lsn_diff(sent_lsn, replay_lsn) AS lag_bytes,
       pg_size_pretty(pg_wal_lsn_diff(sent_lsn, replay_lsn)) AS lag_size
FROM pg_stat_replication ORDER BY lag_bytes DESC`

const sqlTimelineHistory = `
SELECT pg_control_checkpoint(), pg_current_wal_lsn()::text, pg_is_in_recovery()`

// Failover triggers a CNPG cluster failover. Double-confirm-gated.
func Failover(ctx context.Context, _ pginternal.Querier, k8s *pginternal.K8sClient, params map[string]any) (*FailoverResult, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return nil, fmt.Errorf("confirm gate: set confirm=true to execute failover")
	}
	if dconfirm, _ := params["confirm_destructive"].(bool); !dconfirm {
		return nil, fmt.Errorf("double-confirm required: failover may cause data loss. Set confirm_destructive=true")
	}
	clusterName, _ := params["cluster"].(string)
	targetPod, _ := params["target"].(string)

	patch := map[string]any{
		"metadata": map[string]any{
			"annotations": map[string]any{
				"cnpg.io/failoverTarget": targetPod,
			},
		},
	}
	patchBytes, _ := json.Marshal(patch)
	if err := k8s.PatchCluster(ctx, clusterName, patchBytes); err != nil {
		return nil, fmt.Errorf("ha failover: %w", err)
	}
	return &FailoverResult{NewPrimary: targetPod}, nil
}

// Switchover triggers a CNPG cluster switchover. Confirm-gated.
func Switchover(ctx context.Context, _ pginternal.Querier, k8s *pginternal.K8sClient, params map[string]any) (*SwitchoverResult, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return nil, fmt.Errorf("confirm gate: set confirm=true to execute switchover")
	}
	clusterName, _ := params["cluster"].(string)
	targetPod, _ := params["target"].(string)

	patch := map[string]any{
		"metadata": map[string]any{
			"annotations": map[string]any{
				"cnpg.io/switchoverTarget": targetPod,
			},
		},
	}
	patchBytes, _ := json.Marshal(patch)
	if err := k8s.PatchCluster(ctx, clusterName, patchBytes); err != nil {
		return nil, fmt.Errorf("ha switchover: %w", err)
	}
	return &SwitchoverResult{NewPrimary: targetPod}, nil
}

// ReplicationStatus returns streaming replication info.
func ReplicationStatus(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]ReplicaInfo, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlReplicationStatus)
	if err != nil {
		return nil, fmt.Errorf("ha replication status: %w", err)
	}
	defer rows.Close()
	var results []ReplicaInfo
	for rows.Next() {
		var r ReplicaInfo
		if err := rows.Scan(&r.ClientAddr, &r.State, &r.SentLSN, &r.ReplayLSN,
			&r.ReplayLagBytes, &r.SyncState, &r.ReplyTime); err != nil {
			return nil, fmt.Errorf("ha replication status scan: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// SlotList returns replication slots.
func SlotList(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]Slot, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlSlotList)
	if err != nil {
		return nil, fmt.Errorf("ha slot list: %w", err)
	}
	defer rows.Close()
	var results []Slot
	for rows.Next() {
		var s Slot
		if err := rows.Scan(&s.SlotName, &s.SlotType, &s.Active, &s.RestartLSN, &s.RetainedBytes); err != nil {
			return nil, fmt.Errorf("ha slot list scan: %w", err)
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

// SlotCreate creates a replication slot. Confirm-gated.
func SlotCreate(ctx context.Context, q pginternal.Querier, params map[string]any) (*Slot, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return nil, fmt.Errorf("confirm gate: set confirm=true to create replication slot")
	}
	slotName, _ := params["slot_name"].(string)
	slotType, _ := params["slot_type"].(string)
	if slotType == "" {
		slotType = "physical"
	}

	var sql string
	if slotType == "logical" {
		plugin, _ := params["plugin"].(string)
		if plugin == "" {
			plugin = "pgoutput"
		}
		sql = fmt.Sprintf("SELECT pg_create_logical_replication_slot('%s', '%s')", slotName, plugin)
	} else {
		sql = fmt.Sprintf("SELECT pg_create_physical_replication_slot('%s')", slotName)
	}
	_, err := pginternal.Exec(ctx, q, sql)
	if err != nil {
		return nil, fmt.Errorf("ha slot create: %w", err)
	}
	return &Slot{SlotName: slotName, SlotType: slotType, Active: false}, nil
}

// SlotDrop drops a replication slot. Confirm-gated.
func SlotDrop(ctx context.Context, q pginternal.Querier, params map[string]any) (bool, error) {
	if confirmed, _ := params["confirm"].(bool); !confirmed {
		return false, fmt.Errorf("confirm gate: set confirm=true to drop replication slot")
	}
	slotName, _ := params["slot_name"].(string)
	_, err := pginternal.Exec(ctx, q, fmt.Sprintf("SELECT pg_drop_replication_slot('%s')", slotName))
	if err != nil {
		return false, fmt.Errorf("ha slot drop: %w", err)
	}
	return true, nil
}

// ReplicaLag returns lag information for all replicas.
func ReplicaLag(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]LagInfo, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlReplicaLag)
	if err != nil {
		return nil, fmt.Errorf("ha replica lag: %w", err)
	}
	defer rows.Close()
	var results []LagInfo
	for rows.Next() {
		var l LagInfo
		if err := rows.Scan(&l.ClientAddr, &l.LagBytes, &l.LagSize); err != nil {
			return nil, fmt.Errorf("ha replica lag scan: %w", err)
		}
		results = append(results, l)
	}
	return results, rows.Err()
}

// TimelineHistory returns WAL timeline information.
func TimelineHistory(ctx context.Context, q pginternal.Querier, _ map[string]any) (*TimelineInfo, error) {
	row := pginternal.QueryRow(ctx, q, "SELECT pg_current_wal_lsn()::text, pg_is_in_recovery()")
	var info TimelineInfo
	if err := row.Scan(&info.LSN, &info.IsRecovery); err != nil {
		return nil, fmt.Errorf("ha timeline history: %w", err)
	}
	return &info, nil
}

// ReadinessScore assesses HA readiness. Returns a score 0-100.
func ReadinessScore(ctx context.Context, q pginternal.Querier, _ *pginternal.K8sClient, _ map[string]any) (*ReadinessReport, error) {
	report := &ReadinessReport{Score: 100}

	replicas, err := ReplicationStatus(ctx, q, nil)
	if err != nil {
		report.Issues = append(report.Issues, "Cannot check replication: "+err.Error())
		report.Score -= 30
	} else if len(replicas) == 0 {
		report.Issues = append(report.Issues, "No replicas found")
		report.Score -= 50
	} else {
		for _, r := range replicas {
			if r.ReplayLagBytes > 10*1024*1024 { // 10MB
				report.Issues = append(report.Issues, fmt.Sprintf("Replica %s has %.1f MB lag", r.ClientAddr, float64(r.ReplayLagBytes)/1024/1024))
				report.Score -= 20
			}
			report.ReplicaLag = r.ReplayLagBytes
			report.ReplicaSync = r.SyncState
		}
	}

	if report.Score < 0 {
		report.Score = 0
	}
	return report, nil
}

// SwitchoverPlan creates a switchover execution plan.
func SwitchoverPlan(_ context.Context, _ pginternal.Querier, _ *pginternal.K8sClient, params map[string]any) (*SwitchoverPlanResult, error) {
	target, _ := params["target"].(string)
	return &SwitchoverPlanResult{
		Steps: []string{
			"1. Verify replica " + target + " is healthy and in sync",
			"2. Fence primary to stop writes",
			"3. Wait for replay lag to reach zero",
			"4. Promote " + target + " to primary",
			"5. Update connection endpoints",
			"6. Verify new primary is accepting writes",
		},
		Risks:     []string{"Brief write unavailability during switchover"},
		Estimated: "30-60 seconds",
	}, nil
}
