package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"zcyp-im/internal/model"
	"zcyp-im/internal/repository"
)

type AppRepository struct {
	db *sql.DB
}

func NewAppRepository(db *sql.DB) *AppRepository {
	return &AppRepository{db: db}
}

func (r *AppRepository) List() ([]model.App, error) {
	const query = `
SELECT id, app_code, name, app_key, app_secret, status, scenario, created_at, updated_at
FROM apps
ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.App, 0)
	for rows.Next() {
		app, err := scanApp(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, app)
	}

	return items, rows.Err()
}

func (r *AppRepository) Create(params repository.CreateAppParams) (model.App, error) {
	scenarioJSON, err := json.Marshal(params.Scenario)
	if err != nil {
		return model.App{}, err
	}

	const query = `
INSERT INTO apps (app_code, name, app_key, app_secret, status, scenario)
VALUES (?, ?, ?, ?, ?, ?)`

	if _, err := r.db.ExecContext(
		context.Background(),
		query,
		params.AppCode,
		params.Name,
		params.AppKey,
		params.AppSecret,
		params.Status,
		string(scenarioJSON),
	); err != nil {
		return model.App{}, err
	}

	return r.GetByCode(params.AppCode)
}

func (r *AppRepository) GetByCode(appCode string) (model.App, error) {
	const query = `
SELECT id, app_code, name, app_key, app_secret, status, scenario, created_at, updated_at
FROM apps
WHERE app_code = ?
LIMIT 1`

	row := r.db.QueryRowContext(context.Background(), query, appCode)
	app, err := scanApp(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.App{}, fmt.Errorf("app %s: %w", appCode, repository.ErrNotFound)
		}
		return model.App{}, err
	}

	return app, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanApp(s scanner) (model.App, error) {
	var app model.App
	var scenarioRaw string
	if err := s.Scan(
		&app.ID,
		&app.AppCode,
		&app.Name,
		&app.AppKey,
		&app.AppSecret,
		&app.Status,
		&scenarioRaw,
		&app.CreatedAt,
		&app.ModifiedAt,
	); err != nil {
		return model.App{}, err
	}

	if scenarioRaw != "" {
		if err := json.Unmarshal([]byte(scenarioRaw), &app.Scenario); err != nil {
			return model.App{}, err
		}
	}

	return app, nil
}
