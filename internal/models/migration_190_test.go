package models

import (
	"testing"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func TestUpgradeFor190AddsColumns(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	original := Db
	Db = db
	defer func() { Db = original }()

	if err := db.AutoMigrate(&Task{}, &TaskTemplate{}); err != nil {
		t.Fatalf("pre-migrate tables: %v", err)
	}
	// 模拟升级前的旧库:去掉这几列,让 upgradeFor190 真正走 AddColumn 分支。
	// v1.9.0 一次性引入 secret_names(机密白名单)与 notify_keyword_exclude(通知排除关键字)。
	if err := db.Migrator().DropColumn(&Task{}, "secret_names"); err != nil {
		t.Fatalf("drop task.secret_names: %v", err)
	}
	if err := db.Migrator().DropColumn(&Task{}, "notify_keyword_exclude"); err != nil {
		t.Fatalf("drop task.notify_keyword_exclude: %v", err)
	}
	if err := db.Migrator().DropColumn(&TaskTemplate{}, "notify_keyword_exclude"); err != nil {
		t.Fatalf("drop task_template.notify_keyword_exclude: %v", err)
	}

	m := &Migration{}
	if err := m.upgradeFor190(db); err != nil {
		t.Fatalf("upgradeFor190 error: %v", err)
	}
	if !db.Migrator().HasColumn(&Task{}, "secret_names") {
		t.Error("task.secret_names not added by upgradeFor190")
	}
	if !db.Migrator().HasColumn(&Task{}, "notify_keyword_exclude") {
		t.Error("task.notify_keyword_exclude not added by upgradeFor190")
	}
	if !db.Migrator().HasColumn(&TaskTemplate{}, "notify_keyword_exclude") {
		t.Error("task_template.notify_keyword_exclude not added by upgradeFor190")
	}

	// 幂等:列已存在时再次执行不应报错
	if err := m.upgradeFor190(db); err != nil {
		t.Fatalf("upgradeFor190 not idempotent: %v", err)
	}

	// 升级后的存量任务:secret_names 为空串 → 白名单解析为 nil(注入全部);
	// notify_keyword_exclude 为空串 → 不排除。均兼容旧行为。
	task := &Task{Name: "legacy", Protocol: TaskRPC, Command: "echo ok", Spec: "@daily"}
	if _, err := task.Create(); err != nil {
		t.Fatalf("create task: %v", err)
	}
	var loaded Task
	if err := db.First(&loaded, task.Id).Error; err != nil {
		t.Fatalf("load task: %v", err)
	}
	if loaded.SecretNames != "" {
		t.Errorf("expected empty secret_names for legacy task, got %q", loaded.SecretNames)
	}
	if loaded.SecretNameList() != nil {
		t.Errorf("empty secret_names must parse to nil whitelist, got %v", loaded.SecretNameList())
	}
	if loaded.NotifyKeywordExclude != "" {
		t.Errorf("expected empty notify_keyword_exclude for legacy task, got %q", loaded.NotifyKeywordExclude)
	}
}

func TestSecretNameList(t *testing.T) {
	tests := []struct {
		raw  string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{",,", nil},
		{"FOO", []string{"FOO"}},
		{"FOO,BAR", []string{"FOO", "BAR"}},
		{" FOO , BAR ,", []string{"FOO", "BAR"}},
	}
	for _, tt := range tests {
		task := &Task{SecretNames: tt.raw}
		got := task.SecretNameList()
		if len(got) != len(tt.want) {
			t.Errorf("SecretNameList(%q) = %v, want %v", tt.raw, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("SecretNameList(%q) = %v, want %v", tt.raw, got, tt.want)
				break
			}
		}
	}
}
