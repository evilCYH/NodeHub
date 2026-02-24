package sqlite

import (
	"context"
	"fmt"

	"github.com/evilCYH/NodeHub/internal/database/interfaces"
	"github.com/evilCYH/NodeHub/internal/models/setting"
	"github.com/evilCYH/NodeHub/internal/utils/log"
)

func (db *DB) Setting() interfaces.SettingRepository {
	return &SystemConfigRepository{db: db}
}

type SystemConfigRepository struct {
	db *DB
}

func (r *SystemConfigRepository) Create(ctx context.Context, configs *[]setting.Setting) error {
	if configs == nil || len(*configs) == 0 {
		return nil
	}
	tx, err := r.db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO setting (key, value)
	          VALUES (?, ?)`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, config := range *configs {
		_, err := stmt.ExecContext(ctx,
			config.Key,
			config.Value,
		)
		if err != nil {
			return fmt.Errorf("failed to create system config key '%s': %w", config.Key, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *SystemConfigRepository) GetAll(ctx context.Context) (*[]setting.Setting, error) {
	log.Debugf("GetAll")
	query := `SELECT key, value
	          FROM setting ORDER BY key`

	rows, err := r.db.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all configs: %w", err)
	}
	defer rows.Close()

	var configs []setting.Setting
	for rows.Next() {
		var config setting.Setting
		if err := rows.Scan(
			&config.Key,
			&config.Value,
		); err != nil {
			return nil, fmt.Errorf("failed to scan config: %w", err)
		}
		configs = append(configs, config)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate configs: %w", err)
	}

	return &configs, nil
}

func (r *SystemConfigRepository) GetByKey(ctx context.Context, keys []string) (*[]setting.Setting, error) {
	log.Debugf("GetByKey: %v", keys)
	if len(keys) == 0 {
		return &[]setting.Setting{}, nil
	}

	args := make([]interface{}, len(keys))
	inClause := ""
	for i, key := range keys {
		if i > 0 {
			inClause += ","
		}
		inClause += "?"
		args[i] = key
	}
	query := `SELECT key, value
	          FROM setting WHERE key IN (` + inClause + `) ORDER BY key`

	rows, err := r.db.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query configs by keys: %w", err)
	}
	defer rows.Close()

	var configs []setting.Setting
	for rows.Next() {
		var config setting.Setting
		if err := rows.Scan(
			&config.Key,
			&config.Value,
		); err != nil {
			return nil, fmt.Errorf("failed to scan config: %w", err)
		}
		configs = append(configs, config)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate configs: %w", err)
	}

	return &configs, nil
}

func (r *SystemConfigRepository) Update(ctx context.Context, data *[]setting.Setting) error {
	log.Debugf("Update: %v", data)
	if data == nil || len(*data) == 0 {
		return nil
	}

	tx, err := r.db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `UPDATE setting SET value = ? WHERE key = ?`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	for _, updateData := range *data {
		result, err := stmt.ExecContext(ctx,
			updateData.Value,
			updateData.Key,
		)
		if err != nil {
			return fmt.Errorf("failed to update system config key '%s': %w", updateData.Key, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected for key '%s': %w", updateData.Key, err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("no config found with key '%s'", updateData.Key)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
