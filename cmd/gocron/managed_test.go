package main

import (
	"strings"
	"testing"

	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/app"
	"github.com/gocronx-team/gocron/internal/modules/setting"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func TestValidateManagedConfig(t *testing.T) {
	t.Setenv("GOCRON_AUTH_SECRET", "shared-secret")
	config := &setting.Setting{}
	config.Db.Engine = "postgres"
	config.Db.Host = "postgresql.default.svc"
	config.Db.Database = "gocron"
	config.Db.MaxOpenConns = 100
	if err := validateManagedConfig(config); err != nil {
		t.Fatalf("validateManagedConfig() error = %v", err)
	}
}

func TestValidateManagedConfigRejectsSQLite(t *testing.T) {
	t.Setenv("GOCRON_AUTH_SECRET", "shared-secret")
	config := &setting.Setting{}
	config.Db.Engine = "sqlite"
	config.Db.Host = "localhost"
	config.Db.Database = "gocron.db"
	config.Db.MaxOpenConns = 100
	err := validateManagedConfig(config)
	if err == nil || !strings.Contains(err.Error(), "MySQL or PostgreSQL") {
		t.Fatalf("expected managed SQLite rejection, got %v", err)
	}
}

func TestValidateManagedConfigRequiresAuthSecret(t *testing.T) {
	t.Setenv("GOCRON_AUTH_SECRET", "")
	config := &setting.Setting{}
	config.Db.Engine = "mysql"
	config.Db.Host = "mysql.default.svc"
	config.Db.Database = "gocron"
	config.Db.MaxOpenConns = 100
	if err := validateManagedConfig(config); err == nil {
		t.Fatal("expected missing auth secret to be rejected")
	}
}

func TestValidateManagedConfigRequiresTwoConnections(t *testing.T) {
	t.Setenv("GOCRON_AUTH_SECRET", "shared-secret")
	config := &setting.Setting{}
	config.Db.Engine = "postgres"
	config.Db.Host = "postgresql.default.svc"
	config.Db.Database = "gocron"
	config.Db.MaxOpenConns = 1
	if err := validateManagedConfig(config); err == nil {
		t.Fatal("expected a one-connection pool to be rejected")
	}
}

func TestSyncManagedSchemaVersionCreatesCurrentVersion(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	previousDB, previousVersion := models.Db, app.VersionId
	models.Db, app.VersionId = db, 1110
	t.Cleanup(func() {
		models.Db, app.VersionId = previousDB, previousVersion
	})

	if err := syncManagedSchemaVersion(db, new(models.Migration)); err != nil {
		t.Fatalf("syncManagedSchemaVersion() error = %v", err)
	}
	var state models.SchemaVersion
	if err := db.First(&state, 1).Error; err != nil {
		t.Fatalf("read schema version: %v", err)
	}
	if state.Version != 1110 {
		t.Fatalf("schema version = %d, want 1110", state.Version)
	}
}
