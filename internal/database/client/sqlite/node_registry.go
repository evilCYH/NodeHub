package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
)

type NodeRegistryRepository struct {
	db *DB
}

func (r *NodeRegistryRepository) Upsert(ctx context.Context, record *nodeModel.RegistryRecordDB) error {
	query := `INSERT INTO node_registry (
		sub_id, unique_key, raw, info, init_status, last_check_at, last_check_source, last_fail_reason, first_seen_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(sub_id, unique_key) DO UPDATE SET
		raw = excluded.raw,
		info = excluded.info,
		init_status = excluded.init_status,
		last_check_at = excluded.last_check_at,
		last_check_source = excluded.last_check_source,
		last_fail_reason = excluded.last_fail_reason,
		first_seen_at = COALESCE(node_registry.first_seen_at, excluded.first_seen_at),
		updated_at = excluded.updated_at`

	infoRaw := ""
	if record.Info != nil {
		b, err := json.Marshal(record.Info)
		if err != nil {
			return fmt.Errorf("failed to marshal node registry info: %w", err)
		}
		infoRaw = string(b)
	}

	firstSeenAt := record.FirstSeenAt
	if firstSeenAt.IsZero() {
		firstSeenAt = time.Now()
	}
	updatedAt := record.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}

	var lastCheckAt any = nil
	if !record.LastCheckAt.IsZero() {
		lastCheckAt = record.LastCheckAt
	}
	result, err := r.db.db.ExecContext(ctx, query,
		record.SubID,
		strconv.FormatUint(record.UniqueKey, 10),
		record.Raw,
		infoRaw,
		string(record.InitStatus),
		lastCheckAt,
		record.LastCheckSource,
		record.LastFailReason,
		firstSeenAt,
		updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert node registry: %w", err)
	}
	_, _ = result.RowsAffected()

	record.FirstSeenAt = firstSeenAt
	record.UpdatedAt = updatedAt
	return nil
}

func (r *NodeRegistryRepository) List(ctx context.Context) ([]nodeModel.RegistryRecordDB, error) {
	query := `SELECT sub_id, unique_key, raw, info, init_status, last_check_at, last_check_source, last_fail_reason, first_seen_at, updated_at
		FROM node_registry`
	rows, err := r.db.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list node registry: %w", err)
	}
	defer rows.Close()

	result := make([]nodeModel.RegistryRecordDB, 0)
	for rows.Next() {
		var item nodeModel.RegistryRecordDB
		var uniqueKeyRaw string
		var infoRaw sql.NullString
		var initStatusRaw string
		var lastCheckAt sql.NullTime
		var lastCheckSource sql.NullString
		var lastFailReason sql.NullString
		if err := rows.Scan(
			&item.SubID,
			&uniqueKeyRaw,
			&item.Raw,
			&infoRaw,
			&initStatusRaw,
			&lastCheckAt,
			&lastCheckSource,
			&lastFailReason,
			&item.FirstSeenAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan node registry: %w", err)
		}
		uniqueKey, err := strconv.ParseUint(uniqueKeyRaw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node registry unique_key: %w", err)
		}
		item.UniqueKey = uniqueKey
		item.InitStatus = nodeModel.InitStatus(initStatusRaw)
		if lastCheckAt.Valid {
			item.LastCheckAt = lastCheckAt.Time
		}
		if lastCheckSource.Valid {
			item.LastCheckSource = lastCheckSource.String
		}
		if lastFailReason.Valid {
			item.LastFailReason = lastFailReason.String
		}
		if infoRaw.Valid && infoRaw.String != "" {
			var info nodeModel.RegistryInfoSnapshot
			if err := json.Unmarshal([]byte(infoRaw.String), &info); err != nil {
				return nil, fmt.Errorf("failed to unmarshal node registry info: %w", err)
			}
			item.Info = &info
		}

		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate node registry rows: %w", err)
	}
	return result, nil
}

func (r *NodeRegistryRepository) DeleteBySubID(ctx context.Context, subID uint16) error {
	query := `DELETE FROM node_registry WHERE sub_id = ?`
	if _, err := r.db.db.ExecContext(ctx, query, subID); err != nil {
		return fmt.Errorf("failed to delete node registry by sub_id: %w", err)
	}
	return nil
}
