package models

import (
	"encoding/json"
	"testing"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// 创建表
	if err := db.AutoMigrate(&Setting{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

// TestWebhookUrl_JSONMarshaling 测试WebhookUrl的JSON序列化
func TestWebhookUrl_JSONMarshaling(t *testing.T) {
	webhookUrl := WebhookUrl{
		Id:   1,
		Name: "Production Alert",
		Url:  "https://example.com/webhook",
	}

	// 序列化
	data, err := json.Marshal(webhookUrl)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// 反序列化
	var decoded WebhookUrl
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// 验证
	if decoded.Id != webhookUrl.Id {
		t.Errorf("expected Id %d, got %d", webhookUrl.Id, decoded.Id)
	}
	if decoded.Name != webhookUrl.Name {
		t.Errorf("expected Name %s, got %s", webhookUrl.Name, decoded.Name)
	}
	if decoded.Url != webhookUrl.Url {
		t.Errorf("expected Url %s, got %s", webhookUrl.Url, decoded.Url)
	}
}

// TestSetting_CreateWebhookUrl 测试创建webhook地址
func TestSetting_CreateWebhookUrl(t *testing.T) {
	db := setupTestDB(t)
	Db = db

	setting := &Setting{}
	name := "Test Webhook"
	url := "https://test.example.com/webhook"

	rows, err := setting.CreateWebhookUrl(name, url)
	if err != nil {
		t.Fatalf("failed to create webhook url: %v", err)
	}

	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}

	// 验证数据已保存
	var saved Setting
	if err := db.Where(map[string]interface{}{"code": WebhookCode, "key": WebhookUrlKey}).First(&saved).Error; err != nil {
		t.Fatalf("failed to find saved webhook url: %v", err)
	}

	var webhookUrl WebhookUrl
	if err := json.Unmarshal([]byte(saved.Value), &webhookUrl); err != nil {
		t.Fatalf("failed to unmarshal saved value: %v", err)
	}

	if webhookUrl.Name != name {
		t.Errorf("expected name %s, got %s", name, webhookUrl.Name)
	}
	if webhookUrl.Url != url {
		t.Errorf("expected url %s, got %s", url, webhookUrl.Url)
	}
}

// TestSetting_RemoveWebhookUrl 测试删除webhook地址
func TestSetting_RemoveWebhookUrl(t *testing.T) {
	db := setupTestDB(t)
	Db = db

	setting := &Setting{}

	// 先创建一个webhook地址
	_, err := setting.CreateWebhookUrl("Test Webhook", "https://test.example.com/webhook")
	if err != nil {
		t.Fatalf("failed to create webhook url: %v", err)
	}

	// 获取创建的ID
	var saved Setting
	if err := db.Where(map[string]interface{}{"code": WebhookCode, "key": WebhookUrlKey}).First(&saved).Error; err != nil {
		t.Fatalf("failed to find saved webhook url: %v", err)
	}

	// 删除
	rows, err := setting.RemoveWebhookUrl(saved.Id)
	if err != nil {
		t.Fatalf("failed to remove webhook url: %v", err)
	}

	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}

	// 验证已删除
	var count int64
	db.Model(&Setting{}).Where("id = ?", saved.Id).Count(&count)
	if count != 0 {
		t.Errorf("expected webhook url to be deleted, but still exists")
	}
}

// TestSetting_Webhook 测试获取webhook配置
func TestUpdateWebHookMissingTemplateAndErrors(t *testing.T) {
	db := setupTestDB(t)
	original := Db
	Db = db
	t.Cleanup(func() { Db = original })
	model := new(Setting)
	for _, template := range []string{`{"msg_type":"text","content":{"text":"{{.TaskName}}"}}`, "updated", "updated", ""} {
		if err := model.UpdateWebHook(template); err != nil {
			t.Fatal(err)
		}
		got, err := model.Webhook()
		if err != nil || got.Template != template {
			t.Fatalf("template not persisted: got %q, want %q, err %v", got.Template, template, err)
		}
	}
	var count int64
	db.Model(&Setting{}).Where("code = ? AND key = ?", WebhookCode, WebhookTemplateKey).Count(&count)
	if count != 1 {
		t.Fatalf("expected one template row, got %d", count)
	}
	if err := db.Migrator().DropTable(&Setting{}); err != nil {
		t.Fatal(err)
	}
	if err := model.UpdateWebHook("must fail"); err == nil {
		t.Fatal("database error was swallowed")
	}
}

// TestSetting_Webhook 测试获取webhook配置
func TestSetting_Webhook(t *testing.T) {
	db := setupTestDB(t)
	Db = db

	setting := &Setting{}

	// 创建模板配置
	db.Create(&Setting{
		Code:  WebhookCode,
		Key:   WebhookTemplateKey,
		Value: `{"task_id": "{{.TaskId}}"}`,
	})

	// 创建多个webhook地址
	_, _ = setting.CreateWebhookUrl("Webhook 1", "https://webhook1.example.com")
	_, _ = setting.CreateWebhookUrl("Webhook 2", "https://webhook2.example.com")

	// 获取配置
	webhook, err := setting.Webhook()
	if err != nil {
		t.Fatalf("failed to get webhook config: %v", err)
	}

	// 验证模板
	if webhook.Template != `{"task_id": "{{.TaskId}}"}` {
		t.Errorf("unexpected template: %s", webhook.Template)
	}

	// 验证webhook地址数量
	if len(webhook.WebhookUrls) != 2 {
		t.Fatalf("expected 2 webhook urls, got %d", len(webhook.WebhookUrls))
	}

	// 验证webhook地址内容
	names := make(map[string]bool)
	urls := make(map[string]bool)
	for _, w := range webhook.WebhookUrls {
		names[w.Name] = true
		urls[w.Url] = true
	}

	if !names["Webhook 1"] || !names["Webhook 2"] {
		t.Error("webhook names not found")
	}
	if !urls["https://webhook1.example.com"] || !urls["https://webhook2.example.com"] {
		t.Error("webhook urls not found")
	}
}

// TestSetting_UpdateWebHook 测试更新webhook模板
func TestSetting_UpdateWebHook(t *testing.T) {
	db := setupTestDB(t)
	Db = db

	setting := &Setting{}

	// 创建初始模板
	db.Create(&Setting{
		Code:  WebhookCode,
		Key:   WebhookTemplateKey,
		Value: "old template",
	})

	// 更新模板
	newTemplate := `{"task": "{{.TaskName}}"}`
	err := setting.UpdateWebHook(newTemplate)
	if err != nil {
		t.Fatalf("failed to update webhook: %v", err)
	}

	// 验证更新
	var saved Setting
	if err := db.Where(map[string]interface{}{"code": WebhookCode, "key": WebhookTemplateKey}).First(&saved).Error; err != nil {
		t.Fatalf("failed to find updated template: %v", err)
	}

	if saved.Value != newTemplate {
		t.Errorf("expected template %s, got %s", newTemplate, saved.Value)
	}
}

// TestSetting_Webhook_EmptyUrls 测试空webhook地址列表
func TestSetting_Webhook_EmptyUrls(t *testing.T) {
	db := setupTestDB(t)
	Db = db

	setting := &Setting{}

	// 只创建模板，不创建webhook地址
	db.Create(&Setting{
		Code:  WebhookCode,
		Key:   WebhookTemplateKey,
		Value: "template",
	})

	webhook, err := setting.Webhook()
	if err != nil {
		t.Fatalf("failed to get webhook config: %v", err)
	}

	if len(webhook.WebhookUrls) != 0 {
		t.Errorf("expected empty webhook urls, got %d", len(webhook.WebhookUrls))
	}
}

// TestSetting_CreateWebhookUrl_InvalidJSON 测试创建webhook时JSON序列化错误处理
func TestSetting_CreateWebhookUrl_DuplicateNames(t *testing.T) {
	db := setupTestDB(t)
	Db = db

	setting := &Setting{}

	// 创建第一个webhook
	_, err := setting.CreateWebhookUrl("Duplicate", "https://url1.example.com")
	if err != nil {
		t.Fatalf("failed to create first webhook: %v", err)
	}

	// 创建同名webhook（应该允许，因为没有唯一性约束）
	_, err = setting.CreateWebhookUrl("Duplicate", "https://url2.example.com")
	if err != nil {
		t.Fatalf("failed to create second webhook: %v", err)
	}

	// 验证两个都存在
	var count int64
	db.Model(&Setting{}).Where(map[string]interface{}{"code": WebhookCode, "key": WebhookUrlKey}).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 webhook urls, got %d", count)
	}
}

// BenchmarkSetting_CreateWebhookUrl 性能测试：创建webhook地址
func BenchmarkSetting_CreateWebhookUrl(b *testing.B) {
	db, _ := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err := db.AutoMigrate(&Setting{}); err != nil {
		b.Fatalf("failed to migrate: %v", err)
	}
	Db = db

	setting := &Setting{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = setting.CreateWebhookUrl("Benchmark Webhook", "https://benchmark.example.com")
	}
}

// BenchmarkSetting_Webhook 性能测试：获取webhook配置
func BenchmarkSetting_Webhook(b *testing.B) {
	db, _ := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err := db.AutoMigrate(&Setting{}); err != nil {
		b.Fatalf("failed to migrate: %v", err)
	}
	Db = db

	setting := &Setting{}

	// 准备测试数据
	db.Create(&Setting{Code: WebhookCode, Key: WebhookTemplateKey, Value: "template"})
	for i := 0; i < 10; i++ {
		_, _ = setting.CreateWebhookUrl("Webhook", "https://example.com")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = setting.Webhook()
	}
}
