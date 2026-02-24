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

func (r *SubRepository) Create(ctx context.Context, link *sub.Data) error {
	log.Debugf("Create sub")
	query := `INSERT INTO sub (enable, name, tags, cron_expr, config, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	result, err := r.db.db.ExecContext(ctx, query,
		link.Enable,
		link.Name,
		link.Tags,
		link.CronExpr,
		link.Config,
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

	link.ID = uint16(id)
	link.CreatedAt = now
	link.UpdatedAt = now

	return nil
}

func (r *SubRepository) GetByID(ctx context.Context, id uint16) (*sub.Data, error) {
	log.Debugf("Get sub by id")
	query := `SELECT id, enable, name, tags, cron_expr, config, result, created_at, updated_at
	          FROM sub WHERE id = ?`

	var s sub.Data
	err := r.db.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID,
		&s.Enable,
		&s.Name,
		&s.Tags,
		&s.CronExpr,
		&s.Config,
		&s.Result,
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
	query := `UPDATE sub SET enable = ?, name = ?, tags = ?, cron_expr = ?, config = ?, result = ?, updated_at = ? WHERE id = ?`
	data.UpdatedAt = time.Now()
	_, err := r.db.db.ExecContext(ctx, query,
		data.Enable,
		data.Name,
		data.Tags,
		data.CronExpr,
		data.Config,
		data.Result,
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

func (r *SubRepository) List(ctx context.Context) (*[]sub.Data, error) {
	log.Debugf("List sub")
	query := `SELECT id, enable, name, tags, cron_expr, config, result, created_at, updated_at
	          FROM sub ORDER BY id DESC`

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
			&s.Enable,
			&s.Name,
			&s.Tags,
			&s.CronExpr,
			&s.Config,
			&s.Result,
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

	query := `INSERT INTO sub (enable, name, tags, cron_expr, config, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, link := range links {
		result, err := stmt.ExecContext(ctx,
			link.Enable,
			link.Name,
			link.Tags,
			link.CronExpr,
			link.Config,
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
		link.CreatedAt = now
		link.UpdatedAt = now
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
