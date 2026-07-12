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

func TestUpgradeFor170CreatesSecretTable(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	original := Db
	Db = db
	defer func() { Db = original }()

	// upgradeFor170 会迁移 task/task_template 的 notify 字段,需先建表
	if err := db.AutoMigrate(&Task{}, &TaskTemplate{}); err != nil {
		t.Fatalf("pre-migrate tables: %v", err)
	}

	m := &Migration{}
	if err := m.upgradeFor170(db); err != nil {
		t.Fatalf("upgradeFor170 error: %v", err)
	}
	if !db.Migrator().HasTable(&Secret{}) {
		t.Fatal("secret table not created by upgradeFor170")
	}
	// 幂等:表已存在时再次执行不应报错
	if err := m.upgradeFor170(db); err != nil {
		t.Fatalf("upgradeFor170 not idempotent: %v", err)
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

func TestIsReservedEnvName(t *testing.T) {
	for _, n := range []string{"PATH", "path", "Ld_Preload", "LD_LIBRARY_PATH", "HOME", "GOCRON_SECRET_KEY", "  PATH  "} {
		if !IsReservedEnvName(n) {
			t.Errorf("expected %q to be reserved", n)
		}
	}
	for _, n := range []string{"API_KEY", "MY_TOKEN", "DB_PASSWORD", "FOO"} {
		if IsReservedEnvName(n) {
			t.Errorf("expected %q to be allowed", n)
		}
	}
}

func TestUpgradeFor170MigratesNotifyStatus(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	original := Db
	Db = db
	defer func() { Db = original }()
	if err := db.AutoMigrate(&Task{}, &TaskTemplate{}, &Secret{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// 旧 notify_status 语义:0=禁用 1=仅失败 2=总是 3=关键字
	for _, ns := range []int8{0, 1, 2, 3} {
		task := &Task{Name: "t" + string(rune('0'+ns)), NotifyStatus: ns, Command: "x"}
		if err := db.Create(task).Error; err != nil {
			t.Fatalf("seed task ns=%d: %v", ns, err)
		}
	}

	// 真实的 <1.7.0 旧库没有 notify_keyword_regex 列;重映射只在首次加列时执行
	// (列已存在时跳过,避免重复升级二次改写位掩码值)
	if err := db.Migrator().DropColumn(&Task{}, "notify_keyword_regex"); err != nil {
		t.Fatalf("drop task column: %v", err)
	}
	if err := db.Migrator().DropColumn(&TaskTemplate{}, "notify_keyword_regex"); err != nil {
		t.Fatalf("drop template column: %v", err)
	}

	if err := (&Migration{}).upgradeFor170(db); err != nil {
		t.Fatalf("upgrade: %v", err)
	}

	// 位掩码映射:0->0, 1->1(失败位不变), 2->3(失败|成功), 3->4(关键字位)
	want := map[string]int8{"t0": 0, "t1": 1, "t2": 3, "t3": 4}
	for name, exp := range want {
		var tk Task
		db.Where("name = ?", name).First(&tk)
		if tk.NotifyStatus != exp {
			t.Errorf("%s: notify_status = %d, want %d", name, tk.NotifyStatus, exp)
		}
	}
}
