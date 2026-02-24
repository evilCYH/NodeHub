package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/evilCYH/NodeHub/internal/database/interfaces"
	"github.com/evilCYH/NodeHub/internal/models/auth"
	"github.com/evilCYH/NodeHub/internal/utils/log"
)

// Get 获取认证信息
func (db *DB) Auth() interfaces.AuthRepository {
	return &AuthRepository{db: db}
}

// AuthRepository 认证数据访问实现
type AuthRepository struct {
	db *DB
}

// Get 获取认证信息
func (db *AuthRepository) Get(ctx context.Context) (*auth.Data, error) {
	log.Debugf("Get auth")
	query := `SELECT id, username, password FROM auth LIMIT 1`

	var authData auth.Data
	err := db.db.db.QueryRowContext(ctx, query).Scan(
		&authData.ID,
		&authData.UserName,
		&authData.Password,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get auth: %w", err)
	}

	return &authData, nil
}

// UpdateName 更新用户名
func (r *AuthRepository) UpdateName(ctx context.Context, name string) error {
	log.Debugf("UpdateName: %s", name)
	query := `UPDATE auth SET username = ? WHERE id = (SELECT id FROM auth LIMIT 1)`

	result, err := r.db.db.ExecContext(ctx, query, name)
	if err != nil {
		return fmt.Errorf("failed to update username: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no auth record found to update")
	}

	return nil
}

// UpdatePassword 更新密码
func (r *AuthRepository) UpdatePassword(ctx context.Context, hashPassword string) error {
	log.Debugf("UpdatePassword: %s", hashPassword)
	query := `UPDATE auth SET password = ? WHERE id = (SELECT id FROM auth LIMIT 1)`

	result, err := r.db.db.ExecContext(ctx, query, hashPassword)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no auth record found to update")
	}

	return nil
}

// Initialize 初始化认证信息
func (r *AuthRepository) Initialize(ctx context.Context, authData *auth.Data) error {
	log.Debugf("Initialize: %s", authData.UserName)
	query := `INSERT INTO auth (username, password) VALUES (?, ?)`
	_, err := r.db.db.ExecContext(ctx, query, authData.UserName, authData.Password)
	if err != nil {
		return fmt.Errorf("failed to initialize auth: %w", err)
	}

	return nil
}

// IsInitialized 验证是否已初始化
func (r *AuthRepository) IsInitialized(ctx context.Context) (bool, error) {
	log.Debugf("IsInitialized")
	query := `SELECT EXISTS(SELECT 1 FROM auth LIMIT 1)`

	var exists bool
	err := r.db.db.QueryRowContext(ctx, query).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check auth initialization: %w", err)
	}

	return exists, nil
}
