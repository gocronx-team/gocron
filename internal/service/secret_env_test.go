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

	env := loadSecretEnv(models.Task{})
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

func TestLoadSecretEnvSkipsReserved(t *testing.T) {
	_ = os.Setenv(crypto.SecretKeyEnv, "svc-reserved-test")
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

	// 直接写入一条保留名机密(模拟绕过 API 的坏数据),应被 loadSecretEnv 跳过
	evil, _ := crypto.Encrypt("evilpath")
	if _, err := (&models.Secret{Name: "PATH", Value: evil}).Create(); err != nil {
		t.Fatal(err)
	}
	good, _ := crypto.Encrypt("goodval")
	if _, err := (&models.Secret{Name: "API_KEY", Value: good}).Create(); err != nil {
		t.Fatal(err)
	}

	env := loadSecretEnv(models.Task{})
	if _, ok := env["PATH"]; ok {
		t.Error("reserved name PATH must be skipped by loadSecretEnv")
	}
	if env["API_KEY"] != "goodval" {
		t.Errorf("API_KEY should be loaded, got %q", env["API_KEY"])
	}
}

func TestLoadSecretEnvWhitelistScoping(t *testing.T) {
	_ = os.Setenv(crypto.SecretKeyEnv, "svc-whitelist-test")
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

	for name, plain := range map[string]string{
		"DB_PASSWORD": "db-pass",
		"API_KEY":     "api-key",
		"SSH_KEY":     "ssh-key",
	} {
		ct, err := crypto.Encrypt(plain)
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}
		if _, err := (&models.Secret{Name: name, Value: ct}).Create(); err != nil {
			t.Fatalf("create secret %s: %v", name, err)
		}
	}

	// 配置了白名单的任务:只注入列出的机密
	scoped := loadSecretEnv(models.Task{Id: 1, Name: "scoped", SecretNames: "API_KEY"})
	if len(scoped) != 1 || scoped["API_KEY"] != "api-key" {
		t.Errorf("whitelist task must only get API_KEY, got %v", scoped)
	}
	if _, leaked := scoped["DB_PASSWORD"]; leaked {
		t.Error("DB_PASSWORD must not be injected into a task whose whitelist excludes it")
	}

	// 白名单多个名称
	multi := loadSecretEnv(models.Task{Id: 2, Name: "multi", SecretNames: "DB_PASSWORD,SSH_KEY"})
	if len(multi) != 2 || multi["DB_PASSWORD"] != "db-pass" || multi["SSH_KEY"] != "ssh-key" {
		t.Errorf("expected DB_PASSWORD+SSH_KEY, got %v", multi)
	}

	// 白名单引用不存在的机密:不报错,仅注入存在的部分
	partial := loadSecretEnv(models.Task{Id: 3, Name: "partial", SecretNames: "API_KEY,NOT_EXIST"})
	if len(partial) != 1 || partial["API_KEY"] != "api-key" {
		t.Errorf("expected only API_KEY for partial whitelist, got %v", partial)
	}

	// 未配置白名单的任务:保持历史行为,注入全部机密
	legacy := loadSecretEnv(models.Task{Id: 4, Name: "legacy"})
	if len(legacy) != 3 {
		t.Errorf("legacy task without whitelist must get all secrets, got %v", legacy)
	}
}
