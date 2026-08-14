package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/app"
	"github.com/gocronx-team/gocron/internal/modules/setting"
	"gorm.io/gorm"
)

const managedBootstrapLock = "gocron_managed_bootstrap"

func validateManagedConfig(config *setting.Setting) error {
	if config == nil {
		return errors.New("configuration is required")
	}
	engine := strings.ToLower(strings.TrimSpace(config.Db.Engine))
	if engine != "mysql" && engine != "postgres" {
		return errors.New("GOCRON_MANAGED requires MySQL or PostgreSQL")
	}
	if strings.TrimSpace(config.Db.Host) == "" || strings.TrimSpace(config.Db.Database) == "" {
		return errors.New("managed database host and database name are required")
	}
	if config.Db.MaxOpenConns < 2 {
		return errors.New("managed mode requires at least 2 open database connections")
	}
	if strings.TrimSpace(os.Getenv("GOCRON_AUTH_SECRET")) == "" {
		return errors.New("GOCRON_AUTH_SECRET is required in managed mode")
	}
	return nil
}

func bootstrapManagedDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("access database pool: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("reserve bootstrap connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if err := acquireManagedBootstrapLock(ctx, conn, db.Dialector.Name()); err != nil {
		return err
	}
	defer releaseManagedBootstrapLock(conn, db.Dialector.Name())

	migration := new(models.Migration)
	if err := migration.EnsureSchema(); err != nil {
		return fmt.Errorf("ensure managed schema: %w", err)
	}
	if err := syncManagedSchemaVersion(db, migration); err != nil {
		return err
	}
	if err := models.RepairSettings(); err != nil {
		return fmt.Errorf("repair settings: %w", err)
	}

	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return fmt.Errorf("count users: %w", err)
	}
	if userCount > 0 {
		return nil
	}

	username := strings.TrimSpace(os.Getenv("GOCRON_ADMIN_USERNAME"))
	password := os.Getenv("GOCRON_ADMIN_PASSWORD")
	email := strings.TrimSpace(os.Getenv("GOCRON_ADMIN_EMAIL"))
	if len(username) < 3 || len(password) < 6 || email == "" {
		return errors.New("GOCRON_ADMIN_USERNAME, GOCRON_ADMIN_PASSWORD, and GOCRON_ADMIN_EMAIL are required to initialize an empty database")
	}
	user := &models.User{Name: username, Password: password, Email: email, IsAdmin: 1}
	if _, err := user.Create(); err != nil {
		return fmt.Errorf("create managed administrator: %w", err)
	}
	return nil
}

func syncManagedSchemaVersion(db *gorm.DB, migration *models.Migration) error {
	if err := db.AutoMigrate(&models.SchemaVersion{}); err != nil {
		return fmt.Errorf("create managed schema version table: %w", err)
	}
	var state models.SchemaVersion
	result := db.First(&state, 1)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		state = models.SchemaVersion{ID: 1, Version: app.VersionId}
		if err := db.Create(&state).Error; err != nil {
			return fmt.Errorf("record managed schema version: %w", err)
		}
		return nil
	}
	if result.Error != nil {
		return fmt.Errorf("read managed schema version: %w", result.Error)
	}
	if state.Version >= app.VersionId {
		return nil
	}
	migration.Upgrade(state.Version)
	if err := db.Model(&state).Update("version", app.VersionId).Error; err != nil {
		return fmt.Errorf("update managed schema version: %w", err)
	}
	return nil
}

func acquireManagedBootstrapLock(ctx context.Context, conn *sql.Conn, engine string) error {
	switch engine {
	case "postgres":
		if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock(hashtext($1))", managedBootstrapLock); err != nil {
			return fmt.Errorf("acquire PostgreSQL bootstrap lock: %w", err)
		}
	case "mysql":
		var acquired sql.NullInt64
		if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", managedBootstrapLock, 110).Scan(&acquired); err != nil {
			return fmt.Errorf("acquire MySQL bootstrap lock: %w", err)
		}
		if !acquired.Valid || acquired.Int64 != 1 {
			return errors.New("timed out acquiring MySQL bootstrap lock")
		}
	default:
		return fmt.Errorf("managed bootstrap does not support database engine %q", engine)
	}
	return nil
}

func releaseManagedBootstrapLock(conn *sql.Conn, engine string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	switch engine {
	case "postgres":
		_, _ = conn.ExecContext(ctx, "SELECT pg_advisory_unlock(hashtext($1))", managedBootstrapLock)
	case "mysql":
		_, _ = conn.ExecContext(ctx, "SELECT RELEASE_LOCK(?)", managedBootstrapLock)
	}
}
