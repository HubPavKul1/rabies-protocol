package wails

import (
	"context"
	"os"
	"path/filepath"

	"rabies-protocol/internal/domain"
	"rabies-protocol/internal/repository"
)

type App struct {
	ctx context.Context
	db  *repository.DB
}

func NewApp() *App {
	// 1. Определяем имя папки для данных
	dataDir := "data"
	
	// 2. Создаем папку и все необходимые родительские директории, если их нет
	// 0755 - стандартные права доступа (чтение/запись/выполнение для владельца, чтение/выполнение для остальных)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		panic("Не удалось создать папку данных: " + err.Error())
	}

	// 3. Формируем полный путь к файлу базы данных
	dbPath := filepath.Join(dataDir, "rabies.db")

	// 4. Инициализируем базу данных
	db, err := repository.New(dbPath)
	if err != nil {
		panic("Ошибка инициализации базы данных: " + err.Error())
	}
	
	return &App{db: db}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Shutdown(ctx context.Context) {
	if a.db != nil {
		_ = a.db.Close()
	}
}

// --- Методы, доступные из Frontend ---

func (a *App) IsConfigured() (bool, error) {
	cfg, err := a.db.GetClinicConfig(a.ctx)
	if err != nil {
		return false, err
	}
	return cfg.Name != "", nil
}

func (a *App) GetClinicConfig() (*domain.ClinicConfig, error) {
	return a.db.GetClinicConfig(a.ctx)
}

func (a *App) SaveClinicConfig(cfg domain.ClinicConfig) error {
	return a.db.SaveClinicConfig(a.ctx, &cfg)
}

func (a *App) ListVeterinarians() ([]domain.Veterinarian, error) {
	return a.db.ListVeterinarians(a.ctx)
}

func (a *App) CreateVeterinarian(v domain.Veterinarian) error {
	return a.db.CreateVeterinarian(a.ctx, &v)
}