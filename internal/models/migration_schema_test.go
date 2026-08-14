package models

import (
	"testing"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func TestMigrationEnsureSchemaIsIdempotent(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	previous := Db
	Db = db
	t.Cleanup(func() { Db = previous })

	migration := new(Migration)
	if err := migration.EnsureSchema(); err != nil {
		t.Fatalf("first EnsureSchema() error = %v", err)
	}
	if err := migration.EnsureSchema(); err != nil {
		t.Fatalf("second EnsureSchema() error = %v", err)
	}
	for _, table := range []interface{}{&User{}, &Task{}, &TaskLog{}, &Setting{}, &Secret{}} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("missing table for %T", table)
		}
	}
}
