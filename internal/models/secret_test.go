package models

import (
	"testing"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func setupSecretTestDb(t *testing.T) func() {
	t.Helper()
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Secret{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	originalDb := Db
	Db = db
	return func() { Db = originalDb }
}

func TestUpgradeFor166CreatesSecretTable(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	original := Db
	Db = db
	defer func() { Db = original }()

	m := &Migration{}
	if err := m.upgradeFor166(db); err != nil {
		t.Fatalf("upgradeFor166 error: %v", err)
	}
	if !db.Migrator().HasTable(&Secret{}) {
		t.Fatal("secret table not created by upgradeFor166")
	}
	// 幂等:表已存在时再次执行不应报错
	if err := m.upgradeFor166(db); err != nil {
		t.Fatalf("upgradeFor166 not idempotent: %v", err)
	}
}

func TestSecretCreateAndFind(t *testing.T) {
	defer setupSecretTestDb(t)()

	s := &Secret{Name: "API_KEY", Value: "ciphertext", Remark: "demo"}
	id, err := s.Create()
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}

	found := &Secret{}
	if err := found.Find(id); err != nil {
		t.Fatalf("Find error: %v", err)
	}
	if found.Name != "API_KEY" || found.Value != "ciphertext" {
		t.Errorf("unexpected record: %+v", found)
	}
}

func TestSecretListExcludesValue(t *testing.T) {
	defer setupSecretTestDb(t)()

	s := &Secret{Name: "TOKEN", Value: "supersecret"}
	if _, err := s.Create(); err != nil {
		t.Fatalf("Create error: %v", err)
	}

	list, err := s.List()
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(list))
	}
	if list[0].Value != "" {
		t.Errorf("List must not expose Value, got %q", list[0].Value)
	}
	if list[0].Name != "TOKEN" {
		t.Errorf("expected name TOKEN, got %q", list[0].Name)
	}

	// All() 仍应携带密文供注入
	all, err := s.All()
	if err != nil {
		t.Fatalf("All error: %v", err)
	}
	if all[0].Value != "supersecret" {
		t.Errorf("All must include Value, got %q", all[0].Value)
	}
}

func TestSecretNameExists(t *testing.T) {
	defer setupSecretTestDb(t)()

	s := &Secret{Name: "DUP", Value: "v"}
	id, _ := s.Create()

	count, err := s.NameExists("DUP", 0)
	if err != nil {
		t.Fatalf("NameExists error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
	// 排除自身后应为 0(更新场景)
	count, _ = s.NameExists("DUP", id)
	if count != 0 {
		t.Errorf("expected 0 when excluding self, got %d", count)
	}
	// 不存在的名字
	count, _ = s.NameExists("NOPE", 0)
	if count != 0 {
		t.Errorf("expected 0 for missing name, got %d", count)
	}
}

func TestSecretUpdateAndDelete(t *testing.T) {
	defer setupSecretTestDb(t)()

	s := &Secret{Name: "K", Value: "v1"}
	id, _ := s.Create()

	affected, err := s.Update(id, CommonMap{"value": "v2", "remark": "updated"})
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 row updated, got %d", affected)
	}
	found := &Secret{}
	_ = found.Find(id)
	if found.Value != "v2" || found.Remark != "updated" {
		t.Errorf("update not applied: %+v", found)
	}

	affected, err = s.Delete(id)
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 row deleted, got %d", affected)
	}
	if err := found.Find(id); err == nil {
		t.Error("expected record to be gone after delete")
	}
}
