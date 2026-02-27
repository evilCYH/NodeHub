package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
)

type NodeUpdateLogRepository struct {
	db *DB
}

func (r *NodeUpdateLogRepository) Create(ctx context.Context, logEntry *nodeModel.UpdateLog) error {
	query := `INSERT INTO node_update_log (sub_id, run_id, created_at, duration_ms, raw_count, candidate, duplicate, invalid, test_failed, accepted, merged, dropped, details)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	createdAt := logEntry.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	details := encodeDetails(logEntry.Details)
	result, err := r.db.db.ExecContext(ctx, query,
		logEntry.SubID,
		logEntry.RunID,
		createdAt,
		logEntry.DurationMs,
		logEntry.RawCount,
		logEntry.Candidate,
		logEntry.Duplicate,
		logEntry.Invalid,
		logEntry.TestFailed,
		logEntry.Accepted,
		logEntry.Merged,
		logEntry.Dropped,
		details,
	)
	if err != nil {
		return fmt.Errorf("failed to create node update log: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get node update log id: %w", err)
	}
	logEntry.ID = uint64(id)
	logEntry.CreatedAt = createdAt
	return nil
}

func (r *NodeUpdateLogRepository) List(ctx context.Context, subID uint16, limit int) ([]nodeModel.UpdateLog, error) {
	if limit <= 0 {
		limit = 5
	}
	query := `SELECT id, sub_id, run_id, created_at, duration_ms, raw_count, candidate, duplicate, invalid, test_failed, accepted, merged, dropped, details
	          FROM node_update_log WHERE sub_id = ? ORDER BY created_at DESC LIMIT ?`
	rows, err := r.db.db.QueryContext(ctx, query, subID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list node update logs: %w", err)
	}
	defer rows.Close()

	var logs []nodeModel.UpdateLog
	for rows.Next() {
		var logEntry nodeModel.UpdateLog
		var details sql.NullString
		if err := rows.Scan(
			&logEntry.ID,
			&logEntry.SubID,
			&logEntry.RunID,
			&logEntry.CreatedAt,
			&logEntry.DurationMs,
			&logEntry.RawCount,
			&logEntry.Candidate,
			&logEntry.Duplicate,
			&logEntry.Invalid,
			&logEntry.TestFailed,
			&logEntry.Accepted,
			&logEntry.Merged,
			&logEntry.Dropped,
			&details,
		); err != nil {
			return nil, fmt.Errorf("failed to scan node update log: %w", err)
		}
		if details.Valid {
			logEntry.Details = decodeDetails(details.String)
		}
		logs = append(logs, logEntry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate node update logs: %w", err)
	}
	return logs, nil
}

func (r *NodeUpdateLogRepository) ListByRunID(ctx context.Context, subID uint16, runID uint64) ([]nodeModel.UpdateLog, error) {
	query := `SELECT id, sub_id, run_id, created_at, duration_ms, raw_count, candidate, duplicate, invalid, test_failed, accepted, merged, dropped, details
	          FROM node_update_log WHERE sub_id = ? AND run_id = ? ORDER BY created_at DESC`
	rows, err := r.db.db.QueryContext(ctx, query, subID, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to list node update logs by run id: %w", err)
	}
	defer rows.Close()

	var logs []nodeModel.UpdateLog
	for rows.Next() {
		var logEntry nodeModel.UpdateLog
		var details sql.NullString
		if err := rows.Scan(
			&logEntry.ID,
			&logEntry.SubID,
			&logEntry.RunID,
			&logEntry.CreatedAt,
			&logEntry.DurationMs,
			&logEntry.RawCount,
			&logEntry.Candidate,
			&logEntry.Duplicate,
			&logEntry.Invalid,
			&logEntry.TestFailed,
			&logEntry.Accepted,
			&logEntry.Merged,
			&logEntry.Dropped,
			&details,
		); err != nil {
			return nil, fmt.Errorf("failed to scan node update log by run id: %w", err)
		}
		if details.Valid {
			logEntry.Details = decodeDetails(details.String)
		}
		logs = append(logs, logEntry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate node update logs by run id: %w", err)
	}
	return logs, nil
}

func (r *NodeUpdateLogRepository) CleanupBefore(ctx context.Context, before time.Time) error {
	query := `DELETE FROM node_update_log WHERE created_at < ?`
	_, err := r.db.db.ExecContext(ctx, query, before)
	if err != nil {
		return fmt.Errorf("failed to cleanup node update logs: %w", err)
	}
	return nil
}

func encodeDetails(details []string) string {
	if len(details) == 0 {
		return ""
	}
	return strings.Join(details, "\n")
}

func decodeDetails(raw string) []string {
	if raw == "" {
		return nil
	}
	return strings.Split(raw, "\n")
}
