package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	subModel "github.com/evilCYH/NodeHub/internal/models/sub"
)

type SubRunRepository struct {
	db *DB
}

func (r *SubRunRepository) CreateRun(ctx context.Context, run *subModel.RunLog) error {
	query := `INSERT INTO sub_run (sub_id, status, message, raw_count, accepted, created_at, duration_ms)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	createdAt := run.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	result, err := r.db.db.ExecContext(ctx, query,
		run.SubID,
		run.Status,
		run.Message,
		run.RawCount,
		run.Accepted,
		createdAt,
		run.DurationMs,
	)
	if err != nil {
		return fmt.Errorf("failed to create sub run: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get sub run id: %w", err)
	}
	run.ID = uint64(id)
	run.CreatedAt = createdAt
	return nil
}

func (r *SubRunRepository) UpdateRun(ctx context.Context, run *subModel.RunLog) error {
	if run.ID == 0 {
		return fmt.Errorf("sub run id is required")
	}
	query := `UPDATE sub_run SET status = ?, message = ?, raw_count = ?, accepted = ?, duration_ms = ? WHERE id = ?`
	_, err := r.db.db.ExecContext(ctx, query,
		run.Status,
		run.Message,
		run.RawCount,
		run.Accepted,
		run.DurationMs,
		run.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update sub run: %w", err)
	}
	return nil
}

func (r *SubRunRepository) CreateEvent(ctx context.Context, event *subModel.RunEvent) error {
	query := `INSERT INTO sub_run_event (run_id, sub_id, step, level, message, created_at)
	          VALUES (?, ?, ?, ?, ?, ?)`
	createdAt := event.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	result, err := r.db.db.ExecContext(ctx, query,
		event.RunID,
		event.SubID,
		event.Step,
		event.Level,
		event.Message,
		createdAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create sub run event: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get sub run event id: %w", err)
	}
	event.ID = uint64(id)
	event.CreatedAt = createdAt
	return nil
}

func (r *SubRunRepository) ListRuns(ctx context.Context, subID uint16, limit int) ([]subModel.RunLog, error) {
	if limit <= 0 {
		limit = 10
	}
	query := `SELECT id, sub_id, status, message, raw_count, accepted, created_at, duration_ms
	          FROM sub_run WHERE sub_id = ? ORDER BY created_at DESC LIMIT ?`
	rows, err := r.db.db.QueryContext(ctx, query, subID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list sub runs: %w", err)
	}
	defer rows.Close()
	var runs []subModel.RunLog
	for rows.Next() {
		var run subModel.RunLog
		var message sql.NullString
		if err := rows.Scan(
			&run.ID,
			&run.SubID,
			&run.Status,
			&message,
			&run.RawCount,
			&run.Accepted,
			&run.CreatedAt,
			&run.DurationMs,
		); err != nil {
			return nil, fmt.Errorf("failed to scan sub run: %w", err)
		}
		if message.Valid {
			run.Message = message.String
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate sub runs: %w", err)
	}
	return runs, nil
}

func (r *SubRunRepository) ListEvents(ctx context.Context, runID uint64, subID uint16) ([]subModel.RunEvent, error) {
	query := `SELECT id, run_id, sub_id, step, level, message, created_at
	          FROM sub_run_event WHERE run_id = ? AND sub_id = ? ORDER BY created_at ASC`
	rows, err := r.db.db.QueryContext(ctx, query, runID, subID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sub run events: %w", err)
	}
	defer rows.Close()
	var events []subModel.RunEvent
	for rows.Next() {
		var event subModel.RunEvent
		if err := rows.Scan(
			&event.ID,
			&event.RunID,
			&event.SubID,
			&event.Step,
			&event.Level,
			&event.Message,
			&event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan sub run event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate sub run events: %w", err)
	}
	return events, nil
}

func (r *SubRunRepository) CleanupBefore(ctx context.Context, before time.Time) error {
	_, err := r.db.db.ExecContext(ctx, `DELETE FROM sub_run_event WHERE created_at < ?`, before)
	if err != nil {
		return fmt.Errorf("failed to cleanup sub run events: %w", err)
	}
	_, err = r.db.db.ExecContext(ctx, `DELETE FROM sub_run WHERE created_at < ?`, before)
	if err != nil {
		return fmt.Errorf("failed to cleanup sub runs: %w", err)
	}
	return nil
}
