package models

import (
	"testing"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func TestUpgradeFor180AddsNotifyDiagnosisColumn(t *testing.T) {
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
	// 模拟升级前的旧库:去掉列,让 upgradeFor180 真正走 AddColumn 分支
	if err := db.Migrator().DropColumn(&Task{}, "notify_diagnosis"); err != nil {
		t.Fatalf("drop task column: %v", err)
	}
	if err := db.Migrator().DropColumn(&TaskTemplate{}, "notify_diagnosis"); err != nil {
		t.Fatalf("drop template column: %v", err)
	}

	m := &Migration{}
	if err := m.upgradeFor180(db); err != nil {
		t.Fatalf("upgradeFor180 error: %v", err)
	}
	if !db.Migrator().HasColumn(&Task{}, "notify_diagnosis") {
		t.Error("task.notify_diagnosis not added by upgradeFor180")
	}
	if !db.Migrator().HasColumn(&TaskTemplate{}, "notify_diagnosis") {
		t.Error("task_template.notify_diagnosis not added by upgradeFor180")
	}

	// 幂等:列已存在时再次执行不应报错
	if err := m.upgradeFor180(db); err != nil {
		t.Fatalf("upgradeFor180 not idempotent: %v", err)
	}
}
