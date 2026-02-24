package sqlite

import (
	"fmt"
	"time"

	"github.com/evilCYH/NodeHub/internal/database/client/sqlite/migration"
	migModel "github.com/evilCYH/NodeHub/internal/database/migration"
)

func (db *DB) Migrate() error {
	migrations := migration.Get()
	if len(migrations) == 0 {
		return nil
	}
	if err := db.ensureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	appliedDates, err := db.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	var pendingMigrations []*migModel.Info
	for _, migration := range migrations {
		if !appliedDates[migration.Date] {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	if len(pendingMigrations) == 0 {
		return nil
	}

	return db.applyMigrations(pendingMigrations)
}

func (db *DB) ensureMigrationsTable() error {
	migrationTable := `
	CREATE TABLE IF NOT EXISTS "migration" (
		"date" INTEGER NOT NULL UNIQUE,
		"version" TEXT NOT NULL,
		"description" TEXT NOT NULL,
		"applied_at" DATETIME NOT NULL,
		PRIMARY KEY("date")
	);`

	_, err := db.db.Exec(migrationTable)
	if err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}
	return nil
}

func (db *DB) getAppliedMigrations() (map[uint64]bool, error) {
	appliedDates := make(map[uint64]bool)

	rows, err := db.db.Query("SELECT date FROM migration")
	if err != nil {
		return appliedDates, err
	}
	defer rows.Close()

	for rows.Next() {
		var date uint64
		if err := rows.Scan(&date); err != nil {
			return nil, fmt.Errorf("failed to scan migration date: %w", err)
		}
		appliedDates[date] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate migration rows: %w", err)
	}

	return appliedDates, nil
}

func (db *DB) applyMigrations(migrations []*migModel.Info) error {
	tx, err := db.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertStmt, err := tx.Prepare("INSERT INTO migration (date, version, description, applied_at) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer insertStmt.Close()

	for _, migration := range migrations {
		_, err = tx.Exec(migration.Content())
		if err != nil {
			return fmt.Errorf("failed to execute migration %d SQL: %w", migration.Date, err)
		}

		_, err = insertStmt.Exec(migration.Date, migration.Version, migration.Description, time.Now())
		if err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.Date, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migrations transaction: %w", err)
	}

	return nil
}
