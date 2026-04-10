// Package replication provides PostgreSQL replication monitoring tools.
package replication

import (
	"context"
	"fmt"
	"time"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// ReplicaStatus represents a streaming replication replica.
type ReplicaStatus struct {
	ClientAddr    string     `json:"client_addr"`
	State         string     `json:"state"`
	SentLSN       string     `json:"sent_lsn"`
	WriteLSN      string     `json:"write_lsn"`
	FlushLSN      string     `json:"flush_lsn"`
	ReplayLSN     string     `json:"replay_lsn"`
	ReplayLagBytes int64     `json:"replay_lag_bytes"`
	SyncState     string     `json:"sync_state"`
	ReplyTime     *time.Time `json:"reply_time"`
}

// ReplicationSlot represents a replication slot.
type ReplicationSlot struct {
	SlotName       string `json:"slot_name"`
	Plugin         string `json:"plugin"`
	SlotType       string `json:"slot_type"`
	Database       string `json:"database"`
	Active         bool   `json:"active"`
	RestartLSN     string `json:"restart_lsn"`
	ConfirmedFlush string `json:"confirmed_flush_lsn"`
	RetainedBytes  int64  `json:"retained_bytes"`
}

// SlotLagInfo represents replication slot lag details.
type SlotLagInfo struct {
	SlotName      string `json:"slot_name"`
	Active        bool   `json:"active"`
	RetainedBytes int64  `json:"retained_bytes"`
	RetainedSize  string `json:"retained_size"`
}

// Publication represents a logical replication publication.
type Publication struct {
	PubName   string `json:"pub_name"`
	PubOwner  string `json:"pub_owner"`
	AllTables bool   `json:"all_tables"`
	PubInsert bool   `json:"pub_insert"`
	PubUpdate bool   `json:"pub_update"`
	PubDelete bool   `json:"pub_delete"`
}

const sqlStreamingStatus = `
SELECT client_addr::text, state, sent_lsn::text, write_lsn::text, flush_lsn::text, replay_lsn::text,
       pg_wal_lsn_diff(sent_lsn, replay_lsn) AS replay_lag_bytes,
       sync_state, reply_time
FROM pg_stat_replication ORDER BY client_addr`

const sqlSlotList = `
SELECT slot_name, COALESCE(plugin, '') AS plugin, slot_type, COALESCE(database, '') AS database, active,
       COALESCE(restart_lsn::text, '') AS restart_lsn,
       COALESCE(confirmed_flush_lsn::text, '') AS confirmed_flush_lsn,
       pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn) AS retained_bytes
FROM pg_replication_slots ORDER BY slot_name`

const sqlSlotLag = `
SELECT slot_name, active,
       pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn) AS retained_bytes,
       pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn)) AS retained_size
FROM pg_replication_slots ORDER BY retained_bytes DESC`

const sqlPublicationList = `
SELECT pubname, pg_get_userbyid(pubowner) AS pubowner,
       puballtables, pubinsert, pubupdate, pubdelete
FROM pg_publication ORDER BY pubname`

// StreamingStatus returns streaming replication replica status.
func StreamingStatus(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]ReplicaStatus, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlStreamingStatus)
	if err != nil {
		return nil, fmt.Errorf("pg streaming status: %w", err)
	}
	defer rows.Close()
	var results []ReplicaStatus
	for rows.Next() {
		var r ReplicaStatus
		if err := rows.Scan(&r.ClientAddr, &r.State, &r.SentLSN, &r.WriteLSN,
			&r.FlushLSN, &r.ReplayLSN, &r.ReplayLagBytes, &r.SyncState, &r.ReplyTime); err != nil {
			return nil, fmt.Errorf("pg streaming status scan: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// SlotList returns all replication slots.
func SlotList(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]ReplicationSlot, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlSlotList)
	if err != nil {
		return nil, fmt.Errorf("pg slot list: %w", err)
	}
	defer rows.Close()
	var results []ReplicationSlot
	for rows.Next() {
		var s ReplicationSlot
		if err := rows.Scan(&s.SlotName, &s.Plugin, &s.SlotType, &s.Database,
			&s.Active, &s.RestartLSN, &s.ConfirmedFlush, &s.RetainedBytes); err != nil {
			return nil, fmt.Errorf("pg slot list scan: %w", err)
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

// SlotLag returns replication slot lag information.
func SlotLag(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]SlotLagInfo, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlSlotLag)
	if err != nil {
		return nil, fmt.Errorf("pg slot lag: %w", err)
	}
	defer rows.Close()
	var results []SlotLagInfo
	for rows.Next() {
		var s SlotLagInfo
		if err := rows.Scan(&s.SlotName, &s.Active, &s.RetainedBytes, &s.RetainedSize); err != nil {
			return nil, fmt.Errorf("pg slot lag scan: %w", err)
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

// PublicationList returns logical replication publications.
func PublicationList(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]Publication, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlPublicationList)
	if err != nil {
		return nil, fmt.Errorf("pg publication list: %w", err)
	}
	defer rows.Close()
	var results []Publication
	for rows.Next() {
		var p Publication
		if err := rows.Scan(&p.PubName, &p.PubOwner, &p.AllTables,
			&p.PubInsert, &p.PubUpdate, &p.PubDelete); err != nil {
			return nil, fmt.Errorf("pg publication list scan: %w", err)
		}
		results = append(results, p)
	}
	return results, rows.Err()
}
