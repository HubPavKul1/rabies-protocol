package repository

import (
	"context"
	"database/sql"
	// "time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
	"rabies-protocol/internal/domain"
)

type DB struct {
	conn *sql.DB
}

func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	
	// Инициализация схемы БД
	schema := `
	CREATE TABLE IF NOT EXISTS veterinarians (
		id TEXT PRIMARY KEY,
		full_name TEXT NOT NULL UNIQUE,
		position TEXT NOT NULL DEFAULT 'Ветеринарный врач'
	);
	CREATE TABLE IF NOT EXISTS config (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);
	`
	if _, err := conn.Exec(schema); err != nil {
		return nil, err
	}
	
	return &DB{conn: conn}, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) ListVeterinarians(ctx context.Context) ([]domain.Veterinarian, error) {
	rows, err := d.conn.QueryContext(ctx, "SELECT id, full_name, position FROM veterinarians ORDER BY full_name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Veterinarian
	for rows.Next() {
		var v domain.Veterinarian
		if err := rows.Scan(&v.ID, &v.FullName, &v.Position); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

func (d *DB) CreateVeterinarian(ctx context.Context, v *domain.Veterinarian) error {
	v.ID = uuid.New().String()
	_, err := d.conn.ExecContext(ctx,
		"INSERT INTO veterinarians (id, full_name, position) VALUES (?, ?, ?)",
		v.ID, v.FullName, v.Position)
	return err
}

func (d *DB) GetConfig(ctx context.Context, key string) (string, error) {
	var value string
	err := d.conn.QueryRowContext(ctx, "SELECT value FROM config WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (d *DB) SetConfig(ctx context.Context, key, value string) error {
	_, err := d.conn.ExecContext(ctx, "INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)", key, value)
	return err
}

func (d *DB) GetClinicConfig(ctx context.Context) (*domain.ClinicConfig, error) {
	cfg := &domain.ClinicConfig{}
	cfg.Name, _ = d.GetConfig(ctx, "clinic_name")
	cfg.Address, _ = d.GetConfig(ctx, "clinic_address")
	cfg.Phone, _ = d.GetConfig(ctx, "clinic_phone")
	cfg.Email, _ = d.GetConfig(ctx, "clinic_email")
	return cfg, nil
}

func (d *DB) SaveClinicConfig(ctx context.Context, cfg *domain.ClinicConfig) error {
	_ = d.SetConfig(ctx, "clinic_name", cfg.Name)
	_ = d.SetConfig(ctx, "clinic_address", cfg.Address)
	_ = d.SetConfig(ctx, "clinic_phone", cfg.Phone)
	_ = d.SetConfig(ctx, "clinic_email", cfg.Email)
	return nil
}