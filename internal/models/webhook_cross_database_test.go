package models

import (
	"os"
	"sync"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Use dedicated disposable databases only: each test creates and drops its table.
func TestWebhookCrossDatabase(t *testing.T) {
	for _, engine := range []string{"mysql", "postgres"} {
		t.Run(engine, func(t *testing.T) {
			dsn := os.Getenv("GOCRON_TEST_" + map[string]string{"mysql": "MYSQL", "postgres": "POSTGRES"}[engine] + "_DSN")
			if dsn == "" {
				t.Skip("dedicated cross-database DSN not configured")
			}
			var dialect gorm.Dialector = mysql.Open(dsn)
			if engine == "postgres" {
				dialect = postgres.Open(dsn)
			}
			db, err := gorm.Open(dialect, &gorm.Config{})
			if err != nil {
				t.Fatal(err)
			}
			// Refuse to touch an existing settings table, even with an explicit DSN.
			if db.Migrator().HasTable(&Setting{}) {
				t.Fatal("cross-database test requires an empty disposable database")
			}
			if err := db.AutoMigrate(&Setting{}); err != nil {
				t.Fatal(err)
			}
			original := Db
			Db = db
			t.Cleanup(func() {
				Db = original
				if err := db.Migrator().DropTable(&Setting{}); err != nil {
					t.Error(err)
				}
				sqlDB, err := db.DB()
				if err == nil {
					_ = sqlDB.Close()
				}
			})
			start := make(chan struct{})
			errors := make(chan error, 2)
			var writers sync.WaitGroup
			for range 2 {
				writers.Add(1)
				go func() {
					defer writers.Done()
					<-start
					errors <- new(Setting).UpdateWebHook("concurrent")
				}()
			}
			close(start)
			writers.Wait()
			close(errors)
			for err := range errors {
				if err != nil {
					t.Fatal(err)
				}
			}
			for _, template := range []string{`{"text":"{{.TaskName}}"}`, "changed", "changed", ""} {
				if err := new(Setting).UpdateWebHook(template); err != nil {
					t.Fatal(err)
				}
				got, err := new(Setting).Webhook()
				if err != nil || got.Template != template {
					t.Fatalf("template=%q err=%v", got.Template, err)
				}
			}
			var count int64
			if err := db.Model(&Setting{}).Count(&count).Error; err != nil || count != 1 {
				t.Fatalf("count=%d err=%v", count, err)
			}
		})
	}
}
