package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gocronx-team/cron"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGuardScheduledJobReadsSharedTaskStatus(t *testing.T) {
	db := setupScheduledStatusDB(t)

	task := &models.Task{
		Name:             "ha-status-guard",
		Level:            models.TaskLevelParent,
		Spec:             "*/10 * * * * *",
		Protocol:         models.TaskHTTP,
		Command:          "http://example.com",
		HttpMethod:       models.TaskHTTPMethodGet,
		DependencyStatus: models.TaskDependencyStatusWeak,
		Status:           models.Enabled,
	}
	if err := db.Create(task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}

	var runs atomic.Int32
	job := guardScheduledJob(task.Id, func() { runs.Add(1) })

	job()
	if got := runs.Load(); got != 1 {
		t.Fatalf("enabled task runs = %d, want 1", got)
	}

	if err := db.Model(&models.Task{}).Where("id = ?", task.Id).Update("status", models.Disabled).Error; err != nil {
		t.Fatalf("disable task: %v", err)
	}
	job()
	if got := runs.Load(); got != 1 {
		t.Fatalf("disabled task runs = %d, want unchanged at 1", got)
	}

	if err := db.Model(&models.Task{}).Where("id = ?", task.Id).Update("status", models.Enabled).Error; err != nil {
		t.Fatalf("re-enable task: %v", err)
	}
	job()
	if got := runs.Load(); got != 2 {
		t.Fatalf("re-enabled task runs = %d, want 2", got)
	}

	if err := db.Delete(&models.Task{}, task.Id).Error; err != nil {
		t.Fatalf("delete task: %v", err)
	}
	job()
	if got := runs.Load(); got != 2 {
		t.Fatalf("deleted task runs = %d, want unchanged at 2", got)
	}
}

func TestGuardScheduledJobSkipsWhenStatusReadFails(t *testing.T) {
	db := setupScheduledStatusDB(t)

	task := &models.Task{
		Name:             "ha-status-read-error",
		Level:            models.TaskLevelParent,
		Spec:             "*/10 * * * * *",
		Protocol:         models.TaskHTTP,
		Command:          "http://example.com",
		HttpMethod:       models.TaskHTTPMethodGet,
		DependencyStatus: models.TaskDependencyStatusWeak,
		Status:           models.Enabled,
	}
	if err := db.Create(task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}
	if err := db.Migrator().DropTable(&models.Task{}); err != nil {
		t.Fatalf("drop task table: %v", err)
	}

	var runs atomic.Int32
	job := guardScheduledJob(task.Id, func() { runs.Add(1) })
	job()

	if got := runs.Load(); got != 0 {
		t.Fatalf("job runs after status read failure = %d, want 0", got)
	}
}

// Use an on-disk WAL database so tests/benchmarks include real SQLite reads.
func setupScheduledStatusDB(t testing.TB) *gorm.DB {
	t.Helper()
	originalDB, originalPrefix := models.Db, models.TablePrefix
	t.Cleanup(func() {
		models.Db, models.TablePrefix = originalDB, originalPrefix
	})
	db, err := gorm.Open(gormlite.Open(filepath.Join(t.TempDir(), "scheduler.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
		t.Fatal(err)
	}
	models.Db, models.TablePrefix = db, ""
	if err := db.AutoMigrate(&models.Task{}); err != nil {
		t.Fatalf("migrate task table: %v", err)
	}
	return db
}

func TestTaskAddGuardsStaleCronEntry(t *testing.T) {
	db := setupScheduledStatusDB(t)
	originalCron := serviceCron
	serviceCron = cron.New()
	t.Cleanup(func() { serviceCron = originalCron })
	task := models.Task{
		Name: "leader-entry", Level: models.TaskLevelParent,
		Spec: "*/10 * * * * *", Protocol: models.TaskHTTP,
		Command: "http://example.com", Status: models.Enabled, Multi: 1,
	}
	if err := db.Create(&task).Error; err != nil {
		t.Fatal(err)
	}
	// Stop at the first execution-side write, before any HTTP/RPC or queue I/O.
	// Counting this callback detects an unguarded job even if log tables do not exist.
	var admissions int
	if err := db.Callback().Create().Before("gorm:create").Register("test:admission", func(tx *gorm.DB) {
		if _, ok := tx.Statement.Dest.(*models.TaskLog); ok {
			admissions++
			_ = tx.AddError(errors.New("test: stop before task execution"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	ServiceTask.Add(task)
	entries := serviceCron.Entries()
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	entry := entries[0]
	entry.Job.Run()
	if admissions != 1 {
		t.Fatalf("enabled admissions = %d, want 1", admissions)
	}

	// A follower (including an older version) changes only the existing status
	// column. Keep the leader's original entry: no new schema/protocol is needed.
	if err := db.Model(&models.Task{}).Where("id = ?", task.Id).Update("status", models.Disabled).Error; err != nil {
		t.Fatal(err)
	}
	entry.Job.Run()
	if admissions != 1 {
		t.Fatalf("disabled admissions = %d, want unchanged at 1", admissions)
	}
	// The manual-run path still uses createJob without the scheduled guard.
	createJob(task)()
	if admissions != 2 {
		t.Fatalf("manual admissions = %d, want 2", admissions)
	}
	if err := db.Model(&models.Task{}).Where("id = ?", task.Id).Update("status", models.Enabled).Error; err != nil {
		t.Fatal(err)
	}
	entry.Job.Run()
	if admissions != 3 {
		t.Fatalf("re-enabled admissions = %d, want 3", admissions)
	}
	if err := db.Delete(&models.Task{}, task.Id).Error; err != nil {
		t.Fatal(err)
	}
	entry.Job.Run()
	if admissions != 3 {
		t.Fatalf("deleted admissions = %d, want unchanged at 3", admissions)
	}
}

func BenchmarkScheduledStatusGuard(b *testing.B) {
	db := setupScheduledStatusDB(b)
	tasks := make([]models.Task, 1000)
	for i := range tasks {
		tasks[i] = models.Task{Name: fmt.Sprintf("task-%d", i), Status: models.Enabled,
			Spec: "* * * * * *", Protocol: models.TaskHTTP,
			Command: "http://example.com", HttpBody: strings.Repeat("x", 4096)}
	}
	if err := db.CreateInBatches(&tasks, 100).Error; err != nil {
		b.Fatal(err)
	}
	var runs int
	job := cron.FuncJob(func() { runs++ })
	guarded := make([]cron.FuncJob, len(tasks))
	for i := range tasks {
		guarded[i] = guardScheduledJob(tasks[i].Id, job)
	}
	b.Run("Before_NoGuard", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			job()
		}
	})
	b.Run("After_StatusOnly", func(b *testing.B) {
		b.ReportAllocs()
		i := 0
		for b.Loop() {
			guarded[i%len(tasks)]()
			i++
		}
	})
	b.Run("FullRowLookup", func(b *testing.B) {
		b.ReportAllocs()
		i := 0
		for b.Loop() {
			status, err := new(models.Task).GetStatus(tasks[i%len(tasks)].Id)
			if err != nil {
				b.Fatal(err)
			}
			if status == models.Enabled {
				job()
			}
			i++
		}
	})
}

func TestScheduledStatusGuardConcurrentSQLiteUpdate(t *testing.T) {
	db := setupScheduledStatusDB(t)
	task := models.Task{Name: "concurrent-status", Status: models.Enabled,
		Spec: "* * * * * *", Protocol: models.TaskHTTP, Command: "http://example.com"}
	if err := db.Create(&task).Error; err != nil {
		t.Fatal(err)
	}
	// Use a second pool to model a status update independent of the scheduler.
	writer, err := gorm.Open(db.Dialector, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	writerSQL, err := writer.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = writerSQL.Close() })
	var queryCount int
	if err := db.Callback().Query().After("gorm:query").Register("test:read-status", func(tx *gorm.DB) {
		queryCount++
		if tx.Error != nil {
			t.Errorf("concurrent status read: %v", tx.Error)
		}
	}); err != nil {
		t.Fatal(err)
	}
	// Guard reads must not introduce writes or long transactions.
	if err := db.Callback().Update().Before("gorm:update").Register("test:no-update", func(_ *gorm.DB) {
		t.Error("status guard must not write")
	}); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		<-start
		for i := 0; i < 100; i++ {
			status := models.Enabled
			if i%2 == 0 {
				status = models.Disabled
			}
			if err := writer.Model(&models.Task{}).Where("id = ?", task.Id).Update("status", status).Error; err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	job := guardScheduledJob(task.Id, func() {})
	close(start)
	for i := 0; i < 2000; i++ {
		job()
	}
	if err := <-done; err != nil {
		t.Fatalf("concurrent status update: %v", err)
	}
	if queryCount != 2000 {
		t.Fatalf("status queries = %d, want one per trigger (2000)", queryCount)
	}
}
