package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/evilCYH/NodeHub/internal/database/interfaces"
	"github.com/evilCYH/NodeHub/internal/models/check"
	"github.com/evilCYH/NodeHub/internal/utils/log"
)

type CheckRepository struct {
	db *DB
}

func (db *DB) Check() interfaces.CheckRepository {
	return &CheckRepository{db: db}
}

func (r *CheckRepository) Create(ctx context.Context, t *check.Data) error {
	log.Debugf("Create check")
	query := `INSERT INTO check_task (enable, name, task, config, result)
	          VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.db.ExecContext(ctx, query,
		t.Enable,
		t.Name,
		t.Task,
		t.Config,
		t.Result,
	)

	if err != nil {
		return fmt.Errorf("failed to create check: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get check id: %w", err)
	}
	t.ID = uint16(id)
	return nil
}

func (r *CheckRepository) GetByID(ctx context.Context, id uint16) (*check.Data, error) {
	log.Debugf("Get check by id")
	query := `SELECT id, enable, name, task, config, result
	          FROM check_task WHERE id = ?`

	var t check.Data
	err := r.db.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.Enable,
		&t.Name,
		&t.Task,
		&t.Config,
		&t.Result,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get check by id: %w", err)
	}

	return &t, nil
}

func (r *CheckRepository) Update(ctx context.Context, t *check.Data) error {
	log.Debugf("Update check")
	query := `UPDATE check_task SET enable = ?, name = ?, task = ?, config = ?, result = ? WHERE id = ?`

	_, err := r.db.db.ExecContext(ctx, query,
		t.Enable,
		t.Name,
		t.Task,
		t.Config,
		t.Result,
		t.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update check: %w", err)
	}

	return nil
}

func (r *CheckRepository) Delete(ctx context.Context, id uint16) error {
	log.Debugf("Delete check")
	query := `DELETE FROM check_task WHERE id = ?`

	_, err := r.db.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete check: %w", err)
	}

	return nil
}

func (r *CheckRepository) List(ctx context.Context) (*[]check.Data, error) {
	log.Debugf("List check")
	query := `SELECT id, enable, name, task, config, result
	          FROM check_task ORDER BY id DESC`

	rows, err := r.db.db.QueryContext(ctx, query)
	if err != nil {
		log.Errorf("failed to list checks: %v", err)
		return nil, fmt.Errorf("failed to list checks: %w", err)
	}
	defer rows.Close()

	var checks []check.Data
	for rows.Next() {
		var t check.Data
		err := rows.Scan(
			&t.ID,
			&t.Enable,
			&t.Name,
			&t.Task,
			&t.Config,
			&t.Result,
		)
		if err != nil {
			log.Errorf("failed to scan check: %v", err)
			return nil, fmt.Errorf("failed to scan check: %w", err)
		}
		checks = append(checks, t)
	}

	if err = rows.Err(); err != nil {
		log.Errorf("failed to iterate checks: %v", err)
		return nil, fmt.Errorf("failed to iterate checks: %w", err)
	}

	return &checks, nil
}
