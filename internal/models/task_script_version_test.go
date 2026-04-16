package models

import (
	"testing"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func setupVersionTestDB(t *testing.T) func() {
	t.Helper()
	originalDb := Db

	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&TaskScriptVersion{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	Db = db

	return func() {
		Db = originalDb
	}
}

func TestTaskScriptVersion_Create(t *testing.T) {
	cleanup := setupVersionTestDB(t)
	defer cleanup()

	v := &TaskScriptVersion{
		TaskId:   1,
		Command:  "echo hello",
		Remark:   "initial version",
		Username: "admin",
		Version:  1,
	}

	id, err := v.Create()
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}
}

func TestTaskScriptVersion_List_Empty(t *testing.T) {
	cleanup := setupVersionTestDB(t)
	defer cleanup()

	v := new(TaskScriptVersion)
	params := CommonMap{"Page": 1, "PageSize": 10}
	list, err := v.List(999, params)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}

func TestTaskScriptVersion_List_OrderByVersionDesc(t *testing.T) {
	cleanup := setupVersionTestDB(t)
	defer cleanup()

	taskId := 1
	for i := 1; i <= 5; i++ {
		v := &TaskScriptVersion{
			TaskId:   taskId,
			Command:  "echo v" + string(rune('0'+i)),
			Username: "admin",
			Version:  i,
		}
		if _, err := v.Create(); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	v := new(TaskScriptVersion)
	params := CommonMap{"Page": 1, "PageSize": 10}
	list, err := v.List(taskId, params)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(list) != 5 {
		t.Fatalf("expected 5 items, got %d", len(list))
	}
	// 验证降序
	for i := 1; i < len(list); i++ {
		if list[i].Version > list[i-1].Version {
			t.Errorf("expected descending order, but version[%d]=%d > version[%d]=%d",
				i, list[i].Version, i-1, list[i-1].Version)
		}
	}
}

func TestTaskScriptVersion_List_Pagination(t *testing.T) {
	cleanup := setupVersionTestDB(t)
	defer cleanup()

	taskId := 1
	for i := 1; i <= 5; i++ {
		v := &TaskScriptVersion{
			TaskId:  taskId,
			Command: "echo test",
			Version: i,
		}
		if _, err := v.Create(); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	v := new(TaskScriptVersion)
	params := CommonMap{"Page": 1, "PageSize": 3}
	list, err := v.List(taskId, params)
	if err != nil {
		t.Fatalf("List page 1 error: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 items on page 1, got %d", len(list))
	}

	params["Page"] = 2
	list2, err := v.List(taskId, params)
	if err != nil {
		t.Fatalf("List page 2 error: %v", err)
	}
	if len(list2) != 2 {
		t.Errorf("expected 2 items on page 2, got %d", len(list2))
	}
}

func TestTaskScriptVersion_List_IsolatedByTaskId(t *testing.T) {
	cleanup := setupVersionTestDB(t)
	defer cleanup()

	// 创建两个不同任务的版本
	for i := 1; i <= 3; i++ {
		v := &TaskScriptVersion{TaskId: 1, Command: "task1 cmd", Version: i}
		v.Create()
	}
	for i := 1; i <= 2; i++ {
		v := &TaskScriptVersion{TaskId: 2, Command: "task2 cmd", Version: i}
		v.Create()
	}

	v := new(TaskScriptVersion)
	params := CommonMap{"Page": 1, "PageSize": 10}

	list1, _ := v.List(1, params)
	if len(list1) != 3 {
		t.Errorf("expected 3 versions for task 1, got %d", len(list1))
	}

	list2, _ := v.List(2, params)
	if len(list2) != 2 {
		t.Errorf("expected 2 versions for task 2, got %d", len(list2))
	}
}

func TestTaskScriptVersion_Total(t *testing.T) {
	cleanup := setupVersionTestDB(t)
	defer cleanup()

	v := new(TaskScriptVersion)
	total, err := v.Total(1)
	if err != nil {
		t.Fatalf("Total returned error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected 0, got %d", total)
	}

	for i := 1; i <= 3; i++ {
		ver := &TaskScriptVersion{TaskId: 1, Command: "cmd", Version: i}
		ver.Create()
	}

	total, err = v.Total(1)
	if err != nil {
		t.Fatalf("Total returned error: %v", err)
	}
	if total != 3 {
		t.Errorf("expected 3, got %d", total)
	}
}

func TestTaskScriptVersion_Detail(t *testing.T) {
	cleanup := setupVersionTestDB(t)
	defer cleanup()

	v := &TaskScriptVersion{
		TaskId:   1,
		Command:  "echo detail test",
		Remark:   "test remark",
		Username: "alice",
		Version:  1,
	}
	id, _ := v.Create()

	result, err := v.Detail(id)
	if err != nil {
		t.Fatalf("Detail returned error: %v", err)
	}
	if result.Command != "echo detail test" {
		t.Errorf("expected command 'echo detail test', got '%s'", result.Command)
	}
	if result.Remark != "test remark" {
		t.Errorf("expected remark 'test remark', got '%s'", result.Remark)
	}
	if result.Username != "alice" {
		t.Errorf("expected username 'alice', got '%s'", result.Username)
	}
}

func TestTaskScriptVersion_Detail_NotFound(t *testing.T) {
	cleanup := setupVersionTestDB(t)
	defer cleanup()

	v := new(TaskScriptVersion)
	_, err := v.Detail(99999)
	if err == nil {
		t.Error("expected error for non-existent version, got nil")
	}
}

func TestTaskScriptVersion_GetLatestVersion(t *testing.T) {
	cleanup := setupVersionTestDB(t)
	defer cleanup()

	v := new(TaskScriptVersion)

	// 无版本时返回0
	latest, err := v.GetLatestVersion(1)
	if err != nil {
		t.Fatalf("GetLatestVersion returned error: %v", err)
	}
	if latest != 0 {
		t.Errorf("expected 0 for no versions, got %d", latest)
	}

	// 添加几个版本
	for i := 1; i <= 5; i++ {
		ver := &TaskScriptVersion{TaskId: 1, Command: "cmd", Version: i}
		ver.Create()
	}

	latest, err = v.GetLatestVersion(1)
	if err != nil {
		t.Fatalf("GetLatestVersion returned error: %v", err)
	}
	if latest != 5 {
		t.Errorf("expected latest version 5, got %d", latest)
	}
}

func TestTaskScriptVersion_GetLatestVersion_IsolatedByTask(t *testing.T) {
	cleanup := setupVersionTestDB(t)
	defer cleanup()

	// task 1 有 3 个版本，task 2 有 7 个版本
	for i := 1; i <= 3; i++ {
		v := &TaskScriptVersion{TaskId: 1, Command: "cmd", Version: i}
		v.Create()
	}
	for i := 1; i <= 7; i++ {
		v := &TaskScriptVersion{TaskId: 2, Command: "cmd", Version: i}
		v.Create()
	}

	v := new(TaskScriptVersion)
	latest1, _ := v.GetLatestVersion(1)
	latest2, _ := v.GetLatestVersion(2)

	if latest1 != 3 {
		t.Errorf("expected task 1 latest = 3, got %d", latest1)
	}
	if latest2 != 7 {
		t.Errorf("expected task 2 latest = 7, got %d", latest2)
	}
}

func TestTaskScriptVersion_CleanOldVersions(t *testing.T) {
	cleanup := setupVersionTestDB(t)
	defer cleanup()

	taskId := 1
	for i := 1; i <= 10; i++ {
		v := &TaskScriptVersion{TaskId: taskId, Command: "cmd v" + string(rune('0'+i)), Version: i}
		v.Create()
	}

	v := new(TaskScriptVersion)

	// 保留最新 5 个
	err := v.CleanOldVersions(taskId, 5)
	if err != nil {
		t.Fatalf("CleanOldVersions returned error: %v", err)
	}

	total, _ := v.Total(taskId)
	if total != 5 {
		t.Errorf("expected 5 versions after cleanup, got %d", total)
	}

	// 验证保留的是最新的 5 个 (version 6-10)
	params := CommonMap{"Page": 1, "PageSize": 10}
	list, _ := v.List(taskId, params)
	for _, item := range list {
		if item.Version < 6 {
			t.Errorf("expected only versions >= 6, but found version %d", item.Version)
		}
	}
}

func TestTaskScriptVersion_CleanOldVersions_NoOpWhenUnderLimit(t *testing.T) {
	cleanup := setupVersionTestDB(t)
	defer cleanup()

	taskId := 1
	for i := 1; i <= 3; i++ {
		v := &TaskScriptVersion{TaskId: taskId, Command: "cmd", Version: i}
		v.Create()
	}

	v := new(TaskScriptVersion)
	err := v.CleanOldVersions(taskId, 5)
	if err != nil {
		t.Fatalf("CleanOldVersions returned error: %v", err)
	}

	total, _ := v.Total(taskId)
	if total != 3 {
		t.Errorf("expected 3 versions (no cleanup needed), got %d", total)
	}
}

func TestTaskScriptVersion_CleanOldVersions_IsolatedByTask(t *testing.T) {
	cleanup := setupVersionTestDB(t)
	defer cleanup()

	// task 1: 10 个版本
	for i := 1; i <= 10; i++ {
		v := &TaskScriptVersion{TaskId: 1, Command: "cmd", Version: i}
		v.Create()
	}
	// task 2: 3 个版本
	for i := 1; i <= 3; i++ {
		v := &TaskScriptVersion{TaskId: 2, Command: "cmd", Version: i}
		v.Create()
	}

	v := new(TaskScriptVersion)
	v.CleanOldVersions(1, 2)

	total1, _ := v.Total(1)
	total2, _ := v.Total(2)

	if total1 != 2 {
		t.Errorf("expected 2 versions for task 1 after cleanup, got %d", total1)
	}
	if total2 != 3 {
		t.Errorf("expected 3 versions for task 2 (untouched), got %d", total2)
	}
}
