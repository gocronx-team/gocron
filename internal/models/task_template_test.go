package models

import (
	"testing"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func setupTemplateTestDB(t *testing.T) func() {
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

	if err := db.AutoMigrate(&TaskTemplate{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	Db = db

	return func() {
		Db = originalDb
	}
}

func TestTaskTemplate_Create(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	tmpl := &TaskTemplate{
		Name:        "Test Template",
		Description: "A test template",
		Category:    "custom",
		Protocol:    2,
		Command:     "echo hello",
		Timeout:     300,
		CreatedBy:   "admin",
	}

	id, err := tmpl.Create()
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}
}

func TestTaskTemplate_Detail(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	tmpl := &TaskTemplate{
		Name:        "Detail Test",
		Description: "desc",
		Category:    "monitor",
		Protocol:    2,
		Command:     "curl http://example.com",
		Timeout:     30,
		IsBuiltin:   1,
		CreatedBy:   "system",
	}
	id, _ := tmpl.Create()

	result, err := tmpl.Detail(id)
	if err != nil {
		t.Fatalf("Detail returned error: %v", err)
	}
	if result.Name != "Detail Test" {
		t.Errorf("expected name 'Detail Test', got '%s'", result.Name)
	}
	if result.Category != "monitor" {
		t.Errorf("expected category 'monitor', got '%s'", result.Category)
	}
	if result.IsBuiltin != 1 {
		t.Errorf("expected is_builtin = 1, got %d", result.IsBuiltin)
	}
}

func TestTaskTemplate_Detail_NotFound(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	tmpl := new(TaskTemplate)
	_, err := tmpl.Detail(99999)
	if err == nil {
		t.Error("expected error for non-existent template, got nil")
	}
}

func TestTaskTemplate_UpdateBean(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	tmpl := &TaskTemplate{
		Name:     "Original",
		Category: "backup",
		Protocol: 2,
		Command:  "old command",
		Timeout:  100,
	}
	id, _ := tmpl.Create()

	tmpl.Name = "Updated"
	tmpl.Command = "new command"
	tmpl.Timeout = 200
	rows, err := tmpl.UpdateBean(id)
	if err != nil {
		t.Fatalf("UpdateBean returned error: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}

	result, _ := tmpl.Detail(id)
	if result.Name != "Updated" {
		t.Errorf("expected name 'Updated', got '%s'", result.Name)
	}
	if result.Command != "new command" {
		t.Errorf("expected command 'new command', got '%s'", result.Command)
	}
	if result.Timeout != 200 {
		t.Errorf("expected timeout 200, got %d", result.Timeout)
	}
}

func TestTaskTemplate_Delete(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	tmpl := &TaskTemplate{
		Name:     "ToDelete",
		Category: "custom",
		Protocol: 2,
		Command:  "echo bye",
	}
	id, _ := tmpl.Create()

	rows, err := tmpl.Delete(id)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}

	// 确认已删除
	_, err = tmpl.Detail(id)
	if err == nil {
		t.Error("expected error after deletion, got nil")
	}
}

func TestTaskTemplate_List_Empty(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	tmpl := new(TaskTemplate)
	params := CommonMap{"Page": 1, "PageSize": 10}
	list, err := tmpl.List(params)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}

func TestTaskTemplate_List_Pagination(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		tmpl := &TaskTemplate{
			Name:     "tmpl" + string(rune('A'+i)),
			Category: "custom",
			Protocol: 2,
			Command:  "echo test",
		}
		tmpl.Create()
	}

	tmpl := new(TaskTemplate)
	params := CommonMap{"Page": 1, "PageSize": 3}
	list, err := tmpl.List(params)
	if err != nil {
		t.Fatalf("List page 1 error: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 items on page 1, got %d", len(list))
	}

	params["Page"] = 2
	list2, err := tmpl.List(params)
	if err != nil {
		t.Fatalf("List page 2 error: %v", err)
	}
	if len(list2) != 2 {
		t.Errorf("expected 2 items on page 2, got %d", len(list2))
	}
}

func TestTaskTemplate_List_FilterByCategory(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	templates := []TaskTemplate{
		{Name: "backup1", Category: "backup", Protocol: 2, Command: "cmd1"},
		{Name: "monitor1", Category: "monitor", Protocol: 2, Command: "cmd2"},
		{Name: "backup2", Category: "backup", Protocol: 2, Command: "cmd3"},
	}
	for i := range templates {
		templates[i].Create()
	}

	tmpl := new(TaskTemplate)
	params := CommonMap{"Page": 1, "PageSize": 10, "Category": "backup"}
	list, err := tmpl.List(params)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 backup templates, got %d", len(list))
	}
	for _, item := range list {
		if item.Category != "backup" {
			t.Errorf("expected category 'backup', got '%s'", item.Category)
		}
	}
}

func TestTaskTemplate_List_FilterByName(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	templates := []TaskTemplate{
		{Name: "MySQL Backup", Category: "backup", Protocol: 2, Command: "cmd1"},
		{Name: "PG Backup", Category: "backup", Protocol: 2, Command: "cmd2"},
		{Name: "Health Check", Category: "monitor", Protocol: 2, Command: "cmd3"},
	}
	for i := range templates {
		templates[i].Create()
	}

	tmpl := new(TaskTemplate)
	params := CommonMap{"Page": 1, "PageSize": 10, "Name": "Backup"}
	list, err := tmpl.List(params)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 templates matching 'Backup', got %d", len(list))
	}
}

func TestTaskTemplate_Total(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	tmpl := new(TaskTemplate)
	total, _ := tmpl.Total(CommonMap{})
	if total != 0 {
		t.Errorf("expected 0, got %d", total)
	}

	for i := 0; i < 3; i++ {
		t2 := &TaskTemplate{Name: "t" + string(rune('0'+i)), Category: "custom", Protocol: 2, Command: "cmd"}
		t2.Create()
	}

	total, err := tmpl.Total(CommonMap{})
	if err != nil {
		t.Fatalf("Total returned error: %v", err)
	}
	if total != 3 {
		t.Errorf("expected 3, got %d", total)
	}
}

func TestTaskTemplate_Total_WithFilter(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	templates := []TaskTemplate{
		{Name: "t1", Category: "backup", Protocol: 2, Command: "cmd"},
		{Name: "t2", Category: "monitor", Protocol: 2, Command: "cmd"},
		{Name: "t3", Category: "backup", Protocol: 2, Command: "cmd"},
	}
	for i := range templates {
		templates[i].Create()
	}

	tmpl := new(TaskTemplate)
	total, _ := tmpl.Total(CommonMap{"Category": "backup"})
	if total != 2 {
		t.Errorf("expected 2 backup templates, got %d", total)
	}
}

func TestTaskTemplate_NameExist(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	tmpl := &TaskTemplate{Name: "Unique Name", Category: "custom", Protocol: 2, Command: "cmd"}
	id, _ := tmpl.Create()

	// 同名应该存在
	exists, err := tmpl.NameExist("Unique Name", 0)
	if err != nil {
		t.Fatalf("NameExist returned error: %v", err)
	}
	if !exists {
		t.Error("expected name to exist")
	}

	// 排除自身ID后不应该存在
	exists, _ = tmpl.NameExist("Unique Name", id)
	if exists {
		t.Error("expected name not to exist when excluding self")
	}

	// 不存在的名字
	exists, _ = tmpl.NameExist("Other Name", 0)
	if exists {
		t.Error("expected name not to exist")
	}
}

func TestTaskTemplate_IncrementUsage(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	tmpl := &TaskTemplate{Name: "usage test", Category: "custom", Protocol: 2, Command: "cmd"}
	id, _ := tmpl.Create()

	for i := 0; i < 3; i++ {
		if err := tmpl.IncrementUsage(id); err != nil {
			t.Fatalf("IncrementUsage returned error: %v", err)
		}
	}

	result, _ := tmpl.Detail(id)
	if result.UsageCount != 3 {
		t.Errorf("expected usage_count = 3, got %d", result.UsageCount)
	}
}

func TestTaskTemplate_GetCategories(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	templates := []TaskTemplate{
		{Name: "t1", Category: "backup", Protocol: 2, Command: "cmd"},
		{Name: "t2", Category: "monitor", Protocol: 2, Command: "cmd"},
		{Name: "t3", Category: "backup", Protocol: 2, Command: "cmd"},
		{Name: "t4", Category: "deploy", Protocol: 2, Command: "cmd"},
	}
	for i := range templates {
		templates[i].Create()
	}

	tmpl := new(TaskTemplate)
	categories, err := tmpl.GetCategories()
	if err != nil {
		t.Fatalf("GetCategories returned error: %v", err)
	}
	if len(categories) != 3 {
		t.Errorf("expected 3 distinct categories, got %d: %v", len(categories), categories)
	}

	// 验证按字母排序
	expected := []string{"backup", "deploy", "monitor"}
	for i, cat := range categories {
		if cat != expected[i] {
			t.Errorf("expected category[%d] = '%s', got '%s'", i, expected[i], cat)
		}
	}
}

func TestTaskTemplate_GetCategories_Empty(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	tmpl := new(TaskTemplate)
	categories, err := tmpl.GetCategories()
	if err != nil {
		t.Fatalf("GetCategories returned error: %v", err)
	}
	if len(categories) != 0 {
		t.Errorf("expected empty categories, got %v", categories)
	}
}

func TestSeedBuiltinTemplates(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	seedBuiltinTemplates(Db)

	tmpl := new(TaskTemplate)
	total, err := tmpl.Total(CommonMap{})
	if err != nil {
		t.Fatalf("Total returned error: %v", err)
	}
	if total != 9 {
		t.Errorf("expected 9 builtin templates, got %d", total)
	}

	// 验证全部标记为内置
	var list []TaskTemplate
	Db.Where("is_builtin = ?", 1).Find(&list)
	if len(list) != 9 {
		t.Errorf("expected all 9 templates to be builtin, got %d", len(list))
	}

	// 验证分类覆盖
	categories, _ := tmpl.GetCategories()
	if len(categories) < 4 {
		t.Errorf("expected at least 4 categories from builtin templates, got %d: %v", len(categories), categories)
	}
}

func TestSeedBuiltinTemplates_Idempotent(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	seedBuiltinTemplates(Db)
	seedBuiltinTemplates(Db)

	tmpl := new(TaskTemplate)
	total, _ := tmpl.Total(CommonMap{})
	// seedBuiltinTemplates 按 name 去重，两次调用仍然只有 9 条
	if total != 9 {
		t.Errorf("expected 9 templates (idempotent), got %d", total)
	}
}

func TestTaskTemplate_List_BuiltinFirst(t *testing.T) {
	cleanup := setupTemplateTestDB(t)
	defer cleanup()

	// 先创建自定义，再创建内置
	custom := &TaskTemplate{Name: "custom1", Category: "custom", Protocol: 2, Command: "cmd", IsBuiltin: 0}
	custom.Create()
	builtin := &TaskTemplate{Name: "builtin1", Category: "backup", Protocol: 2, Command: "cmd", IsBuiltin: 1}
	builtin.Create()

	tmpl := new(TaskTemplate)
	params := CommonMap{"Page": 1, "PageSize": 10}
	list, _ := tmpl.List(params)

	if len(list) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(list))
	}
	// 内置模板应排在前面
	if list[0].IsBuiltin != 1 {
		t.Errorf("expected builtin template first, got is_builtin=%d", list[0].IsBuiltin)
	}
}
