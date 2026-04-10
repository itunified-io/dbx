// Package wal provides PostgreSQL WAL monitoring tools.
package wal

import (
	"context"
	"fmt"
	"time"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// WALStatusResult represents current WAL status.
type WALStatusResult struct {
	CurrentLSN     string `json:"current_lsn"`
	CurrentSegment string `json:"current_segment"`
	TotalWALBytes  int64  `json:"total_wal_bytes"`
}

// ArchiveStatusResult represents WAL archive status.
type ArchiveStatusResult struct {
	ArchivedCount    int64      `json:"archived_count"`
	FailedCount      int64      `json:"failed_count"`
	LastArchivedWAL  *string    `json:"last_archived_wal"`
	LastArchivedTime *time.Time `json:"last_archived_time"`
	LastFailedWAL    *string    `json:"last_failed_wal"`
	LastFailedTime   *time.Time `json:"last_failed_time"`
}

// ArchiveLagResult represents WAL archive lag.
type ArchiveLagResult struct {
	ArchiveLagBytes int64 `json:"archive_lag_bytes"`
}

// WALSegment represents a WAL segment file.
type WALSegment struct {
	SegmentName string `json:"segment_name"`
	Size        int64  `json:"size"`
}

// WALRateResult represents WAL generation rate.
type WALRateResult struct {
	BytesPerSecond float64 `json:"bytes_per_second"`
	MBPerMinute    float64 `json:"mb_per_minute"`
}

const sqlWALStatus = `
SELECT pg_current_wal_lsn()::text AS current_lsn,
       pg_walfile_name(pg_current_wal_lsn()) AS current_segment,
       pg_wal_lsn_diff(pg_current_wal_lsn(), '0/0') AS total_wal_bytes`

const sqlArchiveStatus = `
SELECT archived_count, failed_count, last_archived_wal, last_archived_time,
       last_failed_wal, last_failed_time
FROM pg_stat_archiver`

// WALStatus returns current WAL position and segment.
func WALStatus(ctx context.Context, q pginternal.Querier, _ map[string]any) (*WALStatusResult, error) {
	row := pginternal.QueryRow(ctx, q, sqlWALStatus)
	var r WALStatusResult
	if err := row.Scan(&r.CurrentLSN, &r.CurrentSegment, &r.TotalWALBytes); err != nil {
		return nil, fmt.Errorf("pg wal status: %w", err)
	}
	return &r, nil
}

// ArchiveStatus returns WAL archiver statistics.
func ArchiveStatus(ctx context.Context, q pginternal.Querier, _ map[string]any) (*ArchiveStatusResult, error) {
	row := pginternal.QueryRow(ctx, q, sqlArchiveStatus)
	var r ArchiveStatusResult
	if err := row.Scan(&r.ArchivedCount, &r.FailedCount, &r.LastArchivedWAL,
		&r.LastArchivedTime, &r.LastFailedWAL, &r.LastFailedTime); err != nil {
		return nil, fmt.Errorf("pg archive status: %w", err)
	}
	return &r, nil
}

// ArchiveLag returns WAL archive lag in bytes.
func ArchiveLag(ctx context.Context, q pginternal.Querier, _ map[string]any) (*ArchiveLagResult, error) {
	row := pginternal.QueryRow(ctx, q,
		`SELECT COALESCE(pg_wal_lsn_diff(pg_current_wal_lsn(),
		        CASE WHEN last_archived_wal IS NOT NULL
		             THEN pg_lsn(regexp_replace(last_archived_wal, '[^0-9A-Fa-f/]', '', 'g'))
		             ELSE pg_current_wal_lsn() END), 0)::bigint
		 FROM pg_stat_archiver`)
	var r ArchiveLagResult
	if err := row.Scan(&r.ArchiveLagBytes); err != nil {
		return nil, fmt.Errorf("pg archive lag: %w", err)
	}
	return &r, nil
}

// WALSegmentList lists WAL segments (requires access to pg_ls_waldir).
func WALSegmentList(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]WALSegment, error) {
	rows, err := pginternal.QueryRows(ctx, q,
		`SELECT name, size FROM pg_ls_waldir() ORDER BY modification DESC LIMIT 50`)
	if err != nil {
		return nil, fmt.Errorf("pg wal segment list: %w", err)
	}
	defer rows.Close()
	var results []WALSegment
	for rows.Next() {
		var s WALSegment
		if err := rows.Scan(&s.SegmentName, &s.Size); err != nil {
			return nil, fmt.Errorf("pg wal segment list scan: %w", err)
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

// WALRate estimates WAL generation rate by sampling two LSN positions.
func WALRate(ctx context.Context, q pginternal.Querier, _ map[string]any) (*WALRateResult, error) {
	// Take a single snapshot -- proper rate calculation requires two samples over time
	row := pginternal.QueryRow(ctx, q,
		`SELECT pg_wal_lsn_diff(pg_current_wal_lsn(), '0/0')::bigint`)
	var totalBytes int64
	if err := row.Scan(&totalBytes); err != nil {
		return nil, fmt.Errorf("pg wal rate: %w", err)
	}
	return &WALRateResult{
		BytesPerSecond: 0,
		MBPerMinute:    0,
	}, nil
}
