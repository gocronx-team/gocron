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
	resetSecretCache()
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
	resetSecretCache()
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

func TestLoadSecretEnvCaching(t *testing.T) {
	resetSecretCache()
	_ = os.Setenv(crypto.SecretKeyEnv, "svc-cache-test")
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

	enc := func(plain string) string {
		ct, err := crypto.Encrypt(plain)
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}
		return ct
	}

	if _, err := (&models.Secret{Name: "FOO", Value: enc("v1")}).Create(); err != nil {
		t.Fatalf("create FOO: %v", err)
	}
	if got := loadSecretEnv(models.Task{})["FOO"]; got != "v1" {
		t.Fatalf("expected FOO=v1, got %q", got)
	}

	// UpdateColumn 不触发 updated_at、条数不变 => 签名不变 => 应命中缓存,仍返回旧值 v1
	if err := models.Db.Model(&models.Secret{}).Where("name = ?", "FOO").
		UpdateColumn("value", enc("v2")).Error; err != nil {
		t.Fatalf("sneaky update: %v", err)
	}
	if got := loadSecretEnv(models.Task{})["FOO"]; got != "v1" {
		t.Fatalf("expected cached v1 while signature unchanged, got %q", got)
	}

	// 新增一条机密 => 条数变化 => 签名变化 => 缓存失效并重新解密:
	// 此时 FOO 应读到最新的 v2,新机密 BAR 也应出现。
	if _, err := (&models.Secret{Name: "BAR", Value: enc("b1")}).Create(); err != nil {
		t.Fatalf("create BAR: %v", err)
	}
	env := loadSecretEnv(models.Task{})
	if env["FOO"] != "v2" {
		t.Errorf("after invalidation FOO should refresh to v2, got %q", env["FOO"])
	}
	if env["BAR"] != "b1" {
		t.Errorf("new secret BAR should appear after invalidation, got %q", env["BAR"])
	}

	// 删除 BAR => 条数变化 => 缓存失效 => BAR 消失。
	if _, err := (&models.Secret{}).Delete(env2Id(t)); err != nil {
		t.Fatalf("delete BAR: %v", err)
	}
	if _, ok := loadSecretEnv(models.Task{})["BAR"]; ok {
		t.Error("BAR should disappear after deletion invalidates the cache")
	}
}

// env2Id 返回 BAR 的主键 id(测试辅助)。
func env2Id(t *testing.T) int {
	t.Helper()
	s := &models.Secret{}
	if err := models.Db.Where("name = ?", "BAR").First(s).Error; err != nil {
		t.Fatalf("lookup BAR id: %v", err)
	}
	return s.Id
}

func TestLoadSecretEnvWhitelistScoping(t *testing.T) {
	resetSecretCache()
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
