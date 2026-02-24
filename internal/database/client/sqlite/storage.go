package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/evilCYH/NodeHub/internal/database/interfaces"
	"github.com/evilCYH/NodeHub/internal/models/storage"
	"github.com/evilCYH/NodeHub/internal/utils/log"
)

// StorageRepository 存储配置数据访问实现
type StorageRepository struct {
	db *DB
}

// newStorageRepository 创建存储配置仓库
func (db *DB) Storage() interfaces.StorageRepository {
	return &StorageRepository{db: db}
}

// Create 创建存储配置
func (r *StorageRepository) Create(ctx context.Context, config *storage.Data) error {
	log.Debugf("Create storage config")
	query := `INSERT INTO storage (name, type, config)
	          VALUES (?, ?, ?)`

	result, err := r.db.db.ExecContext(ctx, query,
		config.Name,
		config.Type,
		config.Config,
	)

	if err != nil {
		return fmt.Errorf("failed to create storage config: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get storage config id: %w", err)
	}

	config.ID = uint16(id)

	return nil
}

// GetByID 根据ID获取存储配置
func (r *StorageRepository) GetByID(ctx context.Context, id uint16) (*storage.Data, error) {
	log.Debugf("Get storage config by id")
	query := `SELECT id, name, type, config
	          FROM storage WHERE id = ?`

	var config storage.Data
	err := r.db.db.QueryRowContext(ctx, query, id).Scan(
		&config.ID,
		&config.Name,
		&config.Type,
		&config.Config,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get storage config by id: %w", err)
	}

	return &config, nil
}

// Update 更新存储配置
func (r *StorageRepository) Update(ctx context.Context, config *storage.Data) error {
	log.Debugf("Update storage config")
	query := `UPDATE storage SET name = ?, type = ?, config = ? WHERE id = ?`

	_, err := r.db.db.ExecContext(ctx, query,
		config.Name,
		config.Type,
		config.Config,
		config.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update storage config: %w", err)
	}

	return nil
}

// Delete 删除存储配置
func (r *StorageRepository) Delete(ctx context.Context, id uint16) error {
	log.Debugf("Delete storage config")
	query := `DELETE FROM storage WHERE id = ?`

	_, err := r.db.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete storage config: %w", err)
	}

	return nil
}

// List 获取存储配置列表
func (r *StorageRepository) List(ctx context.Context) (*[]storage.Data, error) {
	log.Debugf("List storage configs")
	query := `SELECT id, name, type, config
	          FROM storage`

	var configs []storage.Data
	rows, err := r.db.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list storage configs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var config storage.Data
		err := rows.Scan(
			&config.ID,
			&config.Name,
			&config.Type,
			&config.Config,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan storage config: %w", err)
		}
		configs = append(configs, config)
	}

	return &configs, nil
}
