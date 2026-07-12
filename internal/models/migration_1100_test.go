package models

import (
	"testing"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func newMigrationTestDb(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	original := Db
	Db = db
	t.Cleanup(func() { Db = original })

	if err := db.AutoMigrate(&Task{}, &TaskTemplate{}); err != nil {
		t.Fatalf("pre-migrate tables: %v", err)
	}
	return db
}

func TestUpgradeFor1100AddsNotifyKeywordExcludeColumn(t *testing.T) {
	db := newMigrationTestDb(t)

	// 模拟升级前的旧库:去掉列,让 upgradeFor1100 真正走 AddColumn 分支
	if err := db.Migrator().DropColumn(&Task{}, "notify_keyword_exclude"); err != nil {
		t.Fatalf("drop task column: %v", err)
	}
	if err := db.Migrator().DropColumn(&TaskTemplate{}, "notify_keyword_exclude"); err != nil {
		t.Fatalf("drop template column: %v", err)
	}

	m := &Migration{}
	if err := m.upgradeFor1100(db); err != nil {
		t.Fatalf("upgradeFor1100 error: %v", err)
	}
	if !db.Migrator().HasColumn(&Task{}, "notify_keyword_exclude") {
		t.Error("task.notify_keyword_exclude not added by upgradeFor1100")
	}
	if !db.Migrator().HasColumn(&TaskTemplate{}, "notify_keyword_exclude") {
		t.Error("task_template.notify_keyword_exclude not added by upgradeFor1100")
	}

	// 幂等:列已存在时再次执行不应报错
	if err := m.upgradeFor1100(db); err != nil {
		t.Fatalf("upgradeFor1100 not idempotent: %v", err)
	}

	// 存量任务排除关键字为空串 → 不排除,旧行为不变
	task := &Task{Name: "legacy", Protocol: TaskRPC, Command: "echo ok", Spec: "@daily"}
	if _, err := task.Create(); err != nil {
		t.Fatalf("create task: %v", err)
	}
	var loaded Task
	if err := db.First(&loaded, task.Id).Error; err != nil {
		t.Fatalf("load task: %v", err)
	}
	if loaded.NotifyKeywordExclude != "" {
		t.Errorf("expected empty notify_keyword_exclude for legacy task, got %q", loaded.NotifyKeywordExclude)
	}
}

// upgradeFor170 的 notify_status 单选值→位掩码重映射不可重入;它必须只在
// notify_keyword_regex 列首次创建时执行,列已存在时重复调用不得再改写数据。
func TestUpgradeFor170RemapNotRerunWhenColumnExists(t *testing.T) {
	db := newMigrationTestDb(t)

	// 已完成 170 迁移的库:列存在,notify_status 已是位掩码语义
	task := &Task{Name: "bitmask", Protocol: TaskRPC, Command: "echo ok", Spec: "@daily"}
	if _, err := task.Create(); err != nil {
		t.Fatalf("create task: %v", err)
	}
	// 3 = 失败|成功,4 = 关键字 —— 都是重映射后的合法组合值
	if err := db.Model(&Task{}).Where("id = ?", task.Id).Update("notify_status", 3).Error; err != nil {
		t.Fatalf("set notify_status: %v", err)
	}

	m := &Migration{}
	if err := m.upgradeFor170(db); err != nil {
		t.Fatalf("upgradeFor170 error: %v", err)
	}

	var loaded Task
	if err := db.First(&loaded, task.Id).Error; err != nil {
		t.Fatalf("load task: %v", err)
	}
	if loaded.NotifyStatus != 3 {
		t.Errorf("notify_status remap re-ran on migrated db: got %d, want 3", loaded.NotifyStatus)
	}
}

func TestUpgradeFor170RemapRunsOnLegacyDb(t *testing.T) {
	db := newMigrationTestDb(t)

	// 先建任务再删列:Task.Create 显式写入新列,须在列还存在时执行
	mk := func(name string, status int8) int {
		task := &Task{Name: name, Protocol: TaskRPC, Command: "echo ok", Spec: "@daily"}
		id, err := task.Create()
		if err != nil {
			t.Fatalf("create task: %v", err)
		}
		if err := db.Model(&Task{}).Where("id = ?", id).Update("notify_status", status).Error; err != nil {
			t.Fatalf("set notify_status: %v", err)
		}
		return id
	}
	failOnly := mk("legacy-fail", 1) // 旧 1=仅失败 → bit0,不变
	always := mk("legacy-always", 2) // 旧 2=总是 → 3(失败|成功)
	keyword := mk("legacy-kw", 3)    // 旧 3=关键字 → 4

	// 模拟 <1.7.0 的旧库:无 notify_keyword_regex 列,notify_status 是旧单选值
	if err := db.Migrator().DropColumn(&Task{}, "notify_keyword_regex"); err != nil {
		t.Fatalf("drop task column: %v", err)
	}
	if err := db.Migrator().DropColumn(&TaskTemplate{}, "notify_keyword_regex"); err != nil {
		t.Fatalf("drop template column: %v", err)
	}

	m := &Migration{}
	if err := m.upgradeFor170(db); err != nil {
		t.Fatalf("upgradeFor170 error: %v", err)
	}

	want := map[int]int8{failOnly: 1, always: 3, keyword: 4}
	for id, expected := range want {
		var loaded Task
		if err := db.First(&loaded, id).Error; err != nil {
			t.Fatalf("load task %d: %v", id, err)
		}
		if loaded.NotifyStatus != expected {
			t.Errorf("task %d: notify_status = %d, want %d", id, loaded.NotifyStatus, expected)
		}
	}
}

func TestUpgradeStartIndex(t *testing.T) {
	versionIds := []int{110, 122, 130, 140, 150, 151, 152, 153, 154, 155, 156, 157, 158, 159, 1510, 160, 163, 170, 180, 190, 1100}

	tests := []struct {
		name string
		old  int
		want int
	}{
		// 精确匹配:从旧版本号的下一项开始
		{"from 190 runs only 1100", 190, 20},
		{"from 180", 180, 19},
		{"from 170", 170, 18},
		// 1510(v1.5.10)数值大于其后的 160~190,精确匹配保证不会从它重复迁移
		{"from 1510 continues at 160", 1510, 15},
		{"from 159 continues at 1510", 159, 14},
		// 已是最新版本:无需升级
		{"latest version nothing to run", 1100, -1},
		// 未收录的版本号(发过无迁移的补丁版)退回「第一个大于旧版本」扫描
		{"unlisted 161 falls back before 1510 era ends", 161, 14},
		{"unlisted 111 falls back to 122", 111, 1},
		// 高于全部已知迁移的未收录版本:无需升级
		{"unlisted future version nothing to run", 99999, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := upgradeStartIndex(tt.old, versionIds); got != tt.want {
				t.Errorf("upgradeStartIndex(%d) = %d, want %d", tt.old, got, tt.want)
			}
		})
	}
}
