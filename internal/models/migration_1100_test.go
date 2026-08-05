package models

import (
	"testing"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func TestUpgradeFor1100CreatesTaskLogChunkTable(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	oldDb, oldPrefix := Db, TablePrefix
	Db, TablePrefix = db, ""
	defer func() { Db, TablePrefix = oldDb, oldPrefix }()

	m := new(Migration)
	if err := m.upgradeFor1100(db); err != nil {
		t.Fatalf("upgradeFor1100: %v", err)
	}
	if !db.Migrator().HasTable(&TaskLogChunk{}) {
		t.Fatal("task_log_chunk table was not created")
	}
	if !db.Migrator().HasIndex(&TaskLogChunk{}, "idx_task_log_chunk_log_seq") {
		t.Fatal("task_log_chunk sequence index was not created")
	}
	if err := m.upgradeFor1100(db); err != nil {
		t.Fatalf("upgradeFor1100 is not idempotent: %v", err)
	}
}
