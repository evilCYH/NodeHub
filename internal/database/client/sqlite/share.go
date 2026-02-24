package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/evilCYH/NodeHub/internal/database/interfaces"
	"github.com/evilCYH/NodeHub/internal/models/share"
	"github.com/evilCYH/NodeHub/internal/utils/log"
)

type ShareRepository struct {
	db *DB
}

func (db *DB) Share() interfaces.ShareRepository {
	return &ShareRepository{db: db}
}

func (r *ShareRepository) Create(ctx context.Context, shareData *share.Data) error {
	log.Debugf("Create share")
	query := `INSERT INTO share (enable, name, token, gen, max_access_count, expires)
	          VALUES (?, ?, ?, ?, ?, ?)`

	result, err := r.db.db.ExecContext(ctx, query,
		shareData.Enable,
		shareData.Name,
		shareData.Token,
		shareData.Gen,
		shareData.MaxAccessCount,
		shareData.Expires,
	)

	if err != nil {
		return fmt.Errorf("failed to create share link: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get share link id: %w", err)
	}

	shareData.ID = uint16(id)

	return nil
}

func (r *ShareRepository) GetByID(ctx context.Context, id uint16) (*share.Data, error) {
	log.Debugf("Get share by id")
	query := `SELECT id, enable, name, token, gen, access_count, expires, max_access_count
	          FROM share WHERE id = ?`

	var shareData share.Data
	err := r.db.db.QueryRowContext(ctx, query, id).Scan(
		&shareData.ID,
		&shareData.Enable,
		&shareData.Name,
		&shareData.Token,
		&shareData.Gen,
		&shareData.AccessCount,
		&shareData.Expires,
		&shareData.MaxAccessCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get share link by id: %w", err)
	}

	return &shareData, nil
}

func (r *ShareRepository) Update(ctx context.Context, shareData *share.Data) error {
	log.Debugf("Update share")
	query := `UPDATE share SET enable = ?, name = ?, token = ?, gen = ?, access_count = ?, max_access_count = ?, expires = ? WHERE id = ?`

	_, err := r.db.db.ExecContext(ctx, query,
		shareData.Enable,
		shareData.Name,
		shareData.Token,
		shareData.Gen,
		shareData.AccessCount,
		shareData.MaxAccessCount,
		shareData.Expires,
		shareData.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update share link: %w", err)
	}

	return nil
}

func (r *ShareRepository) UpdateAccessCount(ctx context.Context, shareLinks *[]share.UpdateAccessCountDB) error {
	if shareLinks == nil || len(*shareLinks) == 0 {
		return nil
	}
	log.Debugf("Batch update share access count for %d items", len(*shareLinks))

	tx, err := r.db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `UPDATE share SET access_count = ? WHERE id = ?`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, shareLink := range *shareLinks {
		_, err := stmt.ExecContext(ctx, shareLink.AccessCount, shareLink.ID)
		if err != nil {
			return fmt.Errorf("failed to update share access count for id %d: %w", shareLink.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *ShareRepository) Delete(ctx context.Context, id uint16) error {
	log.Debugf("Delete share")
	query := `DELETE FROM share WHERE id = ?`

	_, err := r.db.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete share link: %w", err)
	}

	return nil
}

func (r *ShareRepository) List(ctx context.Context) (*[]share.Data, error) {
	log.Debugf("List share")
	query := `SELECT id, enable, name, token, gen, access_count, expires, max_access_count
	          FROM share`

	rows, err := r.db.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list share links: %w", err)
	}
	defer rows.Close()

	var shareDatas []share.Data
	for rows.Next() {
		var shareData share.Data
		err := rows.Scan(
			&shareData.ID,
			&shareData.Enable,
			&shareData.Name,
			&shareData.Token,
			&shareData.Gen,
			&shareData.AccessCount,
			&shareData.Expires,
			&shareData.MaxAccessCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan share link: %w", err)
		}
		shareDatas = append(shareDatas, shareData)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate share links: %w", err)
	}

	return &shareDatas, nil
}
func (r *ShareRepository) GetGenByToken(ctx context.Context, token string) (string, error) {
	log.Debugf("Get config by token")
	query := `SELECT gen FROM share WHERE token = ?`

	var config string
	err := r.db.db.QueryRowContext(ctx, query, token).Scan(&config)
	if err != nil {
		return "", fmt.Errorf("failed to get config by token: %w", err)
	}
	return config, nil
}
