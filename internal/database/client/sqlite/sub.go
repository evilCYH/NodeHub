package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/evilCYH/NodeHub/internal/database/interfaces"
	"github.com/evilCYH/NodeHub/internal/models/sub"
	"github.com/evilCYH/NodeHub/internal/utils/log"
)

type SubRepository struct {
	db *DB
}

func (db *DB) Sub() interfaces.SubRepository {
	return &SubRepository{db: db}
}

func (db *DB) SubRun() interfaces.SubRunRepository {
	return &SubRunRepository{db: db}
}

func (db *DB) NodeLog() interfaces.NodeLogRepository {
	return &NodeLogRepository{db: db}
}

func (db *DB) NodeUpdateLog() interfaces.NodeUpdateLogRepository {
	return &NodeUpdateLogRepository{db: db}
}

func (db *DB) NodeRegistry() interfaces.NodeRegistryRepository {
	return &NodeRegistryRepository{db: db}
}

func (r *SubRepository) Create(ctx context.Context, link *sub.Data) error {
	log.Debugf("Create sub")
	tx, err := r.db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	var nextSortOrder int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(sort_order) + 1, 0) FROM sub`).Scan(&nextSortOrder); err != nil {
		return fmt.Errorf("failed to query next sort_order: %w", err)
	}

	query := `INSERT INTO sub (sort_order, enable, name, tags, cron_expr, config, upload, download, total, expire, info_updated_at, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := tx.ExecContext(ctx, query,
		nextSortOrder,
		link.Enable,
		link.Name,
		link.Tags,
		link.CronExpr,
		link.Config,
		link.Upload,
		link.Download,
		link.Total,
		link.Expire,
		link.InfoUpdatedAt,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create sub: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get sub id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit create sub transaction: %w", err)
	}

	link.ID = uint16(id)
	link.SortOrder = nextSortOrder
	link.CreatedAt = now
	link.UpdatedAt = now

	return nil
}

func (r *SubRepository) GetByID(ctx context.Context, id uint16) (*sub.Data, error) {
	log.Debugf("Get sub by id")
	query := `SELECT id, sort_order, enable, name, tags, cron_expr, config, result, upload, download, total, expire, info_updated_at, created_at, updated_at
	          FROM sub WHERE id = ?`

	var s sub.Data
	err := r.db.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID,
		&s.SortOrder,
		&s.Enable,
		&s.Name,
		&s.Tags,
		&s.CronExpr,
		&s.Config,
		&s.Result,
		&s.Upload,
		&s.Download,
		&s.Total,
		&s.Expire,
		&s.InfoUpdatedAt,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sub by id: %w", err)
	}

	return &s, nil
}

func (r *SubRepository) Update(ctx context.Context, data *sub.Data) error {
	log.Debugf("Update sub")
	query := `UPDATE sub SET enable = ?, name = ?, tags = ?, cron_expr = ?, config = ?, result = ?, upload = ?, download = ?, total = ?, expire = ?, info_updated_at = ?, updated_at = ? WHERE id = ?`
	data.UpdatedAt = time.Now()
	_, err := r.db.db.ExecContext(ctx, query,
		data.Enable,
		data.Name,
		data.Tags,
		data.CronExpr,
		data.Config,
		data.Result,
		data.Upload,
		data.Download,
		data.Total,
		data.Expire,
		data.InfoUpdatedAt,
		data.UpdatedAt,
		data.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update sub: %w", err)
	}

	return nil
}

func (r *SubRepository) Delete(ctx context.Context, id uint16) error {
	log.Debugf("Delete sub")
	query := `DELETE FROM sub WHERE id = ?`

	_, err := r.db.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete sub: %w", err)
	}

	return nil
}

func (r *SubRepository) DeleteCascade(ctx context.Context, id uint16) (bool, error) {
	log.Debugf("Delete sub cascade")
	tx, err := r.db.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var existed int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM sub WHERE id = ? LIMIT 1`, id).Scan(&existed); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check sub existence: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM sub_run_event WHERE sub_id = ?`, id); err != nil {
		return false, fmt.Errorf("failed to delete sub run events: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM sub_run WHERE sub_id = ?`, id); err != nil {
		return false, fmt.Errorf("failed to delete sub runs: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM node_log WHERE sub_id = ?`, id); err != nil {
		return false, fmt.Errorf("failed to delete node logs: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM node_update_log WHERE sub_id = ?`, id); err != nil {
		return false, fmt.Errorf("failed to delete node update logs: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM node_registry WHERE sub_id = ?`, id); err != nil {
		return false, fmt.Errorf("failed to delete node registry: %w", err)
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM sub WHERE id = ?`, id)
	if err != nil {
		return false, fmt.Errorf("failed to delete sub: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get affected rows: %w", err)
	}
	if affected == 0 {
		return false, nil
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("failed to commit delete sub transaction: %w", err)
	}
	return true, nil
}

func (r *SubRepository) List(ctx context.Context) (*[]sub.Data, error) {
	log.Debugf("List sub")
	query := `SELECT id, sort_order, enable, name, tags, cron_expr, config, result, upload, download, total, expire, info_updated_at, created_at, updated_at
	          FROM sub ORDER BY sort_order ASC, id ASC`

	rows, err := r.db.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list sub: %w", err)
	}
	defer rows.Close()

	var subs []sub.Data
	for rows.Next() {
		var s sub.Data
		err := rows.Scan(
			&s.ID,
			&s.SortOrder,
			&s.Enable,
			&s.Name,
			&s.Tags,
			&s.CronExpr,
			&s.Config,
			&s.Result,
			&s.Upload,
			&s.Download,
			&s.Total,
			&s.Expire,
			&s.InfoUpdatedAt,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sub: %w", err)
		}
		subs = append(subs, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate subs: %w", err)
	}

	return &subs, nil
}

func (r *SubRepository) BatchCreate(ctx context.Context, links []*sub.Data) error {
	log.Debugf("Batch create %d subs", len(links))
	if len(links) == 0 {
		return nil
	}

	tx, err := r.db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()
	var nextSortOrder int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(sort_order) + 1, 0) FROM sub`).Scan(&nextSortOrder); err != nil {
		return fmt.Errorf("failed to query next sort_order: %w", err)
	}

	query := `INSERT INTO sub (sort_order, enable, name, tags, cron_expr, config, upload, download, total, expire, info_updated_at, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, link := range links {
		sortOrder := nextSortOrder + i
		result, err := stmt.ExecContext(ctx,
			sortOrder,
			link.Enable,
			link.Name,
			link.Tags,
			link.CronExpr,
			link.Config,
			link.Upload,
			link.Download,
			link.Total,
			link.Expire,
			link.InfoUpdatedAt,
			now,
			now,
		)
		if err != nil {
			return fmt.Errorf("failed to execute batch insert: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get sub id: %w", err)
		}

		link.ID = uint16(id)
		link.SortOrder = sortOrder
		link.CreatedAt = now
		link.UpdatedAt = now
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *SubRepository) UpdateSortOrder(ctx context.Context, orders []sub.SortOrderItem) error {
	log.Debugf("Update sub sort order, count=%d", len(orders))
	if len(orders) == 0 {
		return nil
	}

	tx, err := r.db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	stmt, err := tx.PrepareContext(ctx, `UPDATE sub SET sort_order = ?, updated_at = ? WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("failed to prepare update sort_order statement: %w", err)
	}
	defer stmt.Close()

	for _, order := range orders {
		result, err := stmt.ExecContext(ctx, order.SortOrder, now, order.ID)
		if err != nil {
			return fmt.Errorf("failed to update sort_order for sub %d: %w", order.ID, err)
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to check affected rows for sub %d: %w", order.ID, err)
		}
		if rowsAffected == 0 {
			return fmt.Errorf("sub not found: %d", order.ID)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit update sort_order transaction: %w", err)
	}
	return nil
}

func (r *SubRepository) UpdateSubInfo(ctx context.Context, id uint16, upload, download, total, expire int64, infoUpdatedAt *time.Time) error {
	log.Debugf("Update sub info")
	query := `UPDATE sub SET upload = ?, download = ?, total = ?, expire = ?, info_updated_at = ? WHERE id = ?`
	_, err := r.db.db.ExecContext(ctx, query, upload, download, total, expire, infoUpdatedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update sub info: %w", err)
	}
	return nil
}
