package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
)

type NodeLogRepository struct {
	db *DB
}

func (r *NodeLogRepository) Create(ctx context.Context, logEntry *nodeModel.NodeLog) error {
	query := `INSERT INTO node_log (sub_id, node_key, node_name, level, source, run_id, check_id, message, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	createdAt := logEntry.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	nodeKey := strconv.FormatUint(logEntry.NodeKey, 10)
	result, err := r.db.db.ExecContext(ctx, query,
		logEntry.SubID,
		nodeKey,
		logEntry.NodeName,
		logEntry.Level,
		string(logEntry.Source),
		logEntry.RunID,
		logEntry.CheckID,
		logEntry.Message,
		createdAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create node log: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get node log id: %w", err)
	}
	logEntry.ID = uint64(id)
	logEntry.CreatedAt = createdAt
	return nil
}

func (r *NodeLogRepository) Query(ctx context.Context, query nodeModel.NodeLogQuery) ([]nodeModel.NodeLog, int64, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 50
	}
	if query.PageSize > 200 {
		query.PageSize = 200
	}

	conditions := []string{"sub_id = ?"}
	args := []any{query.SubID}

	if query.Source != "" {
		conditions = append(conditions, "source = ?")
		args = append(args, query.Source)
	}
	if query.Level != "" {
		conditions = append(conditions, "level = ?")
		args = append(args, query.Level)
	}
	if query.CheckID != 0 {
		conditions = append(conditions, "check_id = ?")
		args = append(args, query.CheckID)
	}
	if query.Keyword != "" {
		conditions = append(conditions, "(node_name LIKE ? OR message LIKE ?)")
		like := "%" + query.Keyword + "%"
		args = append(args, like, like)
	}

	whereClause := strings.Join(conditions, " AND ")
	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM node_log WHERE %s", whereClause)
	var total int64
	if err := r.db.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count node logs: %w", err)
	}

	limit := query.PageSize
	offset := (query.Page - 1) * query.PageSize
	listQuery := fmt.Sprintf(`SELECT id, sub_id, node_key, node_name, level, source, run_id, check_id, message, created_at
		FROM node_log WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, whereClause)
	listArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.db.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query node logs: %w", err)
	}
	defer rows.Close()

	var logs []nodeModel.NodeLog
	for rows.Next() {
		var logEntry nodeModel.NodeLog
		var source sql.NullString
		var nodeKey sql.NullString
		if err := rows.Scan(
			&logEntry.ID,
			&logEntry.SubID,
			&nodeKey,
			&logEntry.NodeName,
			&logEntry.Level,
			&source,
			&logEntry.RunID,
			&logEntry.CheckID,
			&logEntry.Message,
			&logEntry.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan node log: %w", err)
		}
		if nodeKey.Valid {
			value, err := strconv.ParseUint(nodeKey.String, 10, 64)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to parse node key: %w", err)
			}
			logEntry.NodeKey = value
		}
		if source.Valid {
			logEntry.Source = nodeModel.LogSource(source.String)
		}
		logs = append(logs, logEntry)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate node logs: %w", err)
	}

	return logs, total, nil
}

func (r *NodeLogRepository) CleanupBefore(ctx context.Context, before time.Time) error {
	query := `DELETE FROM node_log WHERE created_at < ?`
	_, err := r.db.db.ExecContext(ctx, query, before)
	if err != nil {
		return fmt.Errorf("failed to cleanup node logs: %w", err)
	}
	return nil
}
