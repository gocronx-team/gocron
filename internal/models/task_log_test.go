package models

import (
	"testing"
	"time"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func setupTaskLogTestDb(t testing.TB) func() {
	t.Helper()
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	err = db.AutoMigrate(&TaskLog{}, &TaskHost{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	originalDb := Db
	Db = db
	return func() {
		Db = originalDb
	}
}

func BenchmarkList_HostFilter(b *testing.B) {
	cleanup := setupTaskLogTestDb(b)
	defer cleanup()

	const taskCount = 1000
	logs := make([]TaskLog, 10_000)
	for i := range logs {
		logs[i] = TaskLog{
			TaskId:  i%taskCount + 1,
			Name:    "benchmark",
			Spec:    "*",
			Command: "true",
		}
	}
	if err := Db.CreateInBatches(&logs, 500).Error; err != nil {
		b.Fatalf("create logs: %v", err)
	}
	taskHosts := make([]TaskHost, taskCount)
	for i := range taskHosts {
		taskHosts[i] = TaskHost{TaskId: i + 1, HostId: i%10 + 1}
	}
	if err := Db.CreateInBatches(&taskHosts, 500).Error; err != nil {
		b.Fatalf("create task hosts: %v", err)
	}

	params := CommonMap{"HostId": 7, "Page": 1, "PageSize": 20}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := new(TaskLog).List(params); err != nil {
			b.Fatal(err)
		}
	}
}

func TestList_HostAndTimeRangeFilter(t *testing.T) {
	cleanup := setupTaskLogTestDb(t)
	defer cleanup()

	base := time.Date(2026, 8, 7, 12, 0, 0, 0, time.Local)
	logs := []TaskLog{
		{TaskId: 1, Name: "host-7-in-range", Spec: "*", Command: "echo 1", StartTime: LocalTime(base)},
		{TaskId: 1, Name: "host-7-too-old", Spec: "*", Command: "echo 2", StartTime: LocalTime(base.AddDate(0, 0, -10))},
		{TaskId: 2, Name: "host-8-in-range", Spec: "*", Command: "echo 3", StartTime: LocalTime(base)},
	}
	for i := range logs {
		if _, err := logs[i].Create(); err != nil {
			t.Fatalf("create log: %v", err)
		}
	}
	if err := Db.Create(&[]TaskHost{{TaskId: 1, HostId: 7}, {TaskId: 2, HostId: 8}}).Error; err != nil {
		t.Fatalf("create task hosts: %v", err)
	}

	got, err := new(TaskLog).List(CommonMap{
		"HostId":    7,
		"StartTime": base.AddDate(0, 0, -6),
		"EndTime":   base.AddDate(0, 0, 1),
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 || got[0].TaskId != 1 || got[0].Name != "host-7-in-range" {
		t.Fatalf("combined host/date filter returned %+v", got)
	}
}

func TestList_TimeRangeFilter(t *testing.T) {
	cleanup := setupTaskLogTestDb(t)
	defer cleanup()

	base := time.Date(2026, 6, 15, 12, 0, 0, 0, time.Local)
	seed := []TaskLog{
		{TaskId: 1, Name: "t1", Spec: "* * * * *", Command: "echo 1", Result: "ok", StartTime: LocalTime(base.Add(-2 * time.Hour))},
		{TaskId: 2, Name: "t2", Spec: "* * * * *", Command: "echo 2", Result: "ok", StartTime: LocalTime(base.Add(-1 * time.Hour))},
		{TaskId: 3, Name: "t3", Spec: "* * * * *", Command: "echo 3", Result: "ok", StartTime: LocalTime(base)},
	}
	for i := range seed {
		if _, err := seed[i].Create(); err != nil {
			t.Fatalf("create log: %v", err)
		}
	}

	// StartTime 下界：含 90 分钟前之后 → 命中 t2、t3
	after, err := new(TaskLog).List(CommonMap{"StartTime": base.Add(-90 * time.Minute)})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(after) != 2 {
		t.Fatalf("StartTime filter expected 2, got %d", len(after))
	}

	// EndTime 上界：90 分钟前之前(不含) → 仅命中 t1
	before, err := new(TaskLog).List(CommonMap{"EndTime": base.Add(-90 * time.Minute)})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(before) != 1 || before[0].TaskId != 1 {
		t.Fatalf("EndTime filter expected only t1, got %+v", before)
	}

	// 区间 [90min前, 30min前) → 仅命中 t2
	mid, err := new(TaskLog).List(CommonMap{
		"StartTime": base.Add(-90 * time.Minute),
		"EndTime":   base.Add(-30 * time.Minute),
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(mid) != 1 || mid[0].TaskId != 2 {
		t.Fatalf("range filter expected only t2, got %+v", mid)
	}
}

func TestList_KeywordFilter(t *testing.T) {
	cleanup := setupTaskLogTestDb(t)
	defer cleanup()

	seed := []TaskLog{
		{TaskId: 1, Name: "新增用户任务", Spec: "* * * * *", Command: "echo 1", Result: "ok"},
		{TaskId: 2, Name: "清理日志", Spec: "* * * * *", Command: "echo 2", Result: "build goland project"},
		{TaskId: 3, Name: "新增订单同步", Spec: "* * * * *", Command: "echo 3", Result: "ok"},
	}
	for i := range seed {
		if _, err := seed[i].Create(); err != nil {
			t.Fatalf("create log: %v", err)
		}
	}

	// 关键字命中任务名（两条）
	byName, err := new(TaskLog).List(CommonMap{"Keyword": "新增"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(byName) != 2 {
		t.Fatalf("expected 2 logs matching name 新增, got %d", len(byName))
	}

	// 关键字命中执行输出（仅任务2的 result 含 goland，name 不含）
	byResult, err := new(TaskLog).List(CommonMap{"Keyword": "goland"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(byResult) != 1 || byResult[0].TaskId != 2 {
		t.Fatalf("expected 1 log matching result goland (task 2), got %+v", byResult)
	}

	// 空关键字不过滤
	all, err := new(TaskLog).List(CommonMap{"Keyword": ""})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("empty keyword should not filter, got %d", len(all))
	}

	// 无匹配
	none, err := new(TaskLog).List(CommonMap{"Keyword": "不存在的关键字"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(none))
	}

	// 关键字 + 任务ID 组合过滤（验证 OR 条件被正确括组，不会与 task_id 串味）
	combo, err := new(TaskLog).List(CommonMap{"Keyword": "新增", "TaskId": 3})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(combo) != 1 || combo[0].TaskId != 3 {
		t.Fatalf("combined filter failed: %+v", combo)
	}
}

func TestClearByTaskId_Normal(t *testing.T) {
	cleanup := setupTaskLogTestDb(t)
	defer cleanup()

	// Insert logs for task 1 and task 2
	for i := 0; i < 5; i++ {
		log := &TaskLog{TaskId: 1, Name: "task1", Spec: "* * * * *", Command: "echo 1", Result: "ok"}
		if _, err := log.Create(); err != nil {
			t.Fatalf("failed to create log: %v", err)
		}
	}
	for i := 0; i < 3; i++ {
		log := &TaskLog{TaskId: 2, Name: "task2", Spec: "* * * * *", Command: "echo 2", Result: "ok"}
		if _, err := log.Create(); err != nil {
			t.Fatalf("failed to create log: %v", err)
		}
	}

	taskLog := new(TaskLog)
	affected, err := taskLog.ClearByTaskId(1)
	if err != nil {
		t.Fatalf("ClearByTaskId returned error: %v", err)
	}
	if affected != 5 {
		t.Errorf("expected 5 affected rows, got %d", affected)
	}

	// Verify task 1 logs are gone
	var count int64
	Db.Model(&TaskLog{}).Where("task_id = ?", 1).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 remaining logs for task 1, got %d", count)
	}

	// Verify task 2 logs are untouched
	Db.Model(&TaskLog{}).Where("task_id = ?", 2).Count(&count)
	if count != 3 {
		t.Errorf("expected 3 remaining logs for task 2, got %d", count)
	}
}

func TestClearByTaskId_NoLogs(t *testing.T) {
	cleanup := setupTaskLogTestDb(t)
	defer cleanup()

	taskLog := new(TaskLog)
	affected, err := taskLog.ClearByTaskId(999)
	if err != nil {
		t.Fatalf("ClearByTaskId returned error: %v", err)
	}
	if affected != 0 {
		t.Errorf("expected 0 affected rows, got %d", affected)
	}
}

func TestClearByTaskId_ZeroId(t *testing.T) {
	cleanup := setupTaskLogTestDb(t)
	defer cleanup()

	taskLog := new(TaskLog)
	affected, err := taskLog.ClearByTaskId(0)
	if err != nil {
		t.Fatalf("ClearByTaskId returned error: %v", err)
	}
	if affected != 0 {
		t.Errorf("expected 0 affected rows, got %d", affected)
	}
}
