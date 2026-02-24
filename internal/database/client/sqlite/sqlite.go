package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/evilCYH/NodeHub/internal/database/interfaces"
	_ "modernc.org/sqlite"
)

// DB SQLite数据库连接包装器
type DB struct {
	db *sql.DB
}

// New 创建新的SQLite数据库连接
func New(databasePath string) (interfaces.Repository, error) {
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	if err := enablePragmas(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set pragmas: %w", err)
	}
	repository := DB{db: db}

	return &repository, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.db.Close()
}

// enablePragmas 启用SQLite优化选项
func enablePragmas(db *sql.DB) error {
	pragmas := map[string]string{
		"journal_mode":       "WAL",    // 启用WAL模式提高并发性能
		"synchronous":        "NORMAL", // 平衡性能和安全性
		"cache_size":         "-64000", // 64MB缓存
		"foreign_keys":       "ON",     // 启用外键约束
		"temp_store":         "MEMORY", // 临时表存储在内存中
		"busy_timeout":       "5000",   // 5秒忙等待超时
		"wal_autocheckpoint": "1000",   // WAL自动检查点
		"optimize":           "",       // 优化数据库
	}

	for key, value := range pragmas {
		var query string
		if value == "" {
			query = fmt.Sprintf("PRAGMA %s", key)
		} else {
			query = fmt.Sprintf("PRAGMA %s = %s", key, value)
		}

		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute pragma %s: %w", query, err)
		}
	}

	return nil
}
