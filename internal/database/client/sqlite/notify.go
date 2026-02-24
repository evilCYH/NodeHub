package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/evilCYH/NodeHub/internal/database/interfaces"
	"github.com/evilCYH/NodeHub/internal/models/notify"
	"github.com/evilCYH/NodeHub/internal/utils/log"
)

type NotifyRepository struct {
	db *DB
}

func (db *DB) Notify() interfaces.NotifyRepository {
	return &NotifyRepository{db: db}
}

func (r *NotifyRepository) Create(ctx context.Context, channel *notify.Data) error {
	log.Debugf("Create notify")
	query := `INSERT INTO notify (name, type, config )
	          VALUES (?, ?, ?)`

	result, err := r.db.db.ExecContext(ctx, query,
		channel.Name,
		channel.Type,
		channel.Config,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification channel: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get notification channel id: %w", err)
	}

	channel.ID = uint16(id)

	return nil
}

func (r *NotifyRepository) GetByID(ctx context.Context, id uint16) (*notify.Data, error) {
	log.Debugf("Get notify by id")
	query := `SELECT id, name, type, config
	          FROM notify WHERE id = ?`

	var channel notify.Data
	err := r.db.db.QueryRowContext(ctx, query, id).Scan(
		&channel.ID,
		&channel.Name,
		&channel.Type,
		&channel.Config,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get notification channel by id: %w", err)
	}

	return &channel, nil
}

func (r *NotifyRepository) Update(ctx context.Context, channel *notify.Data) error {
	log.Debugf("Update notify")
	query := `UPDATE notify SET name = ?, type = ?, config = ? WHERE id = ?`

	_, err := r.db.db.ExecContext(ctx, query,
		channel.Name,
		channel.Type,
		channel.Config,
		channel.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update notification channel: %w", err)
	}

	return nil
}

func (r *NotifyRepository) Delete(ctx context.Context, id uint16) error {
	log.Debugf("Delete notify")
	query := `DELETE FROM notify WHERE id = ?`

	_, err := r.db.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification channel: %w", err)
	}

	return nil
}

func (r *NotifyRepository) List(ctx context.Context) (*[]notify.Data, error) {
	log.Debugf("List notify")
	query := `SELECT id, name, type, config
	          FROM notify ORDER BY id DESC`

	rows, err := r.db.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list notification channels: %w", err)
	}
	defer rows.Close()

	var channels []notify.Data
	for rows.Next() {
		var channel notify.Data
		err := rows.Scan(
			&channel.ID,
			&channel.Name,
			&channel.Type,
			&channel.Config,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification channel: %w", err)
		}
		channels = append(channels, channel)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate notification channels: %w", err)
	}

	return &channels, nil
}

type NotifyTemplateRepository struct {
	db *DB
}

func (db *DB) NotifyTemplate() interfaces.NotifyTemplateRepository {
	return &NotifyTemplateRepository{db: db}
}

func (r *NotifyTemplateRepository) Create(ctx context.Context, template *notify.Template) error {
	log.Debugf("Create notify template")
	query := `INSERT INTO notify_template (type, title, content)
	          VALUES (?, ?, ?)`

	_, err := r.db.db.ExecContext(ctx, query,
		template.Type,
		template.Title,
		template.Content,
	)

	if err != nil {
		return fmt.Errorf("failed to create notify template: %w", err)
	}

	return nil
}

func (r *NotifyTemplateRepository) GetByType(ctx context.Context, t string) (*notify.Template, error) {
	log.Debugf("Get notify template by type")
	query := `SELECT type, title, content
	          FROM notify_template WHERE type = ?`

	var template notify.Template
	err := r.db.db.QueryRowContext(ctx, query, t).Scan(
		&template.Type,
		&template.Title,
		&template.Content,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get notify template by type: %w", err)
	}

	return &template, nil
}
func (r *NotifyTemplateRepository) Update(ctx context.Context, template *notify.Template) error {
	log.Debugf("Update Notify Template")
	query := `UPDATE notify_template SET title = ?, content = ? WHERE type = ?`

	_, err := r.db.db.ExecContext(ctx, query,
		template.Title,
		template.Content,
		template.Type,
	)
	if err != nil {
		return fmt.Errorf("failed to update notify template: %w", err)
	}
	return nil
}

func (r *NotifyTemplateRepository) List(ctx context.Context) (*[]notify.Template, error) {
	log.Debugf("List notify template")
	query := `SELECT type, title, content
	          FROM notify_template ORDER BY type DESC`

	rows, err := r.db.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list notify templates: %w", err)
	}
	defer rows.Close()

	var templates []notify.Template
	for rows.Next() {
		var template notify.Template
		err := rows.Scan(
			&template.Type,
			&template.Title,
			&template.Content,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notify template: %w", err)
		}
		templates = append(templates, template)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate notify templates: %w", err)
	}

	return &templates, nil
}
