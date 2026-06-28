package service

import (
	"os"
	"testing"

	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/crypto"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func TestLoadSecretEnvDecrypts(t *testing.T) {
	_ = os.Setenv(crypto.SecretKeyEnv, "svc-secret-test")
	crypto.Init()
	if !crypto.Configured() {
		t.Skip("crypto master key not configured")
	}

	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	original := models.Db
	models.Db = db
	defer func() { models.Db = original }()
	if err := db.AutoMigrate(&models.Secret{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	ct, err := crypto.Encrypt("plainval")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if _, err := (&models.Secret{Name: "FOO", Value: ct}).Create(); err != nil {
		t.Fatalf("create secret: %v", err)
	}

	env := loadSecretEnv()
	if env["FOO"] != "plainval" {
		t.Errorf("expected FOO=plainval, got %q", env["FOO"])
	}

	values := secretValues(env)
	if len(values) != 1 || values[0] != "plainval" {
		t.Errorf("unexpected secret values: %v", values)
	}
}

func TestSecretValuesEmpty(t *testing.T) {
	if secretValues(nil) != nil {
		t.Error("expected nil for empty env map")
	}
	if secretValues(map[string]string{}) != nil {
		t.Error("expected nil for zero-length env map")
	}
}
