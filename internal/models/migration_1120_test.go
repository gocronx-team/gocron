package models

import (
	"testing"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func TestUpgradeFor1120PreservesExistingRows(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	old := Db
	Db = db
	defer func() { Db = old }()
	if err := db.AutoMigrate(&Task{}, &TaskTemplate{}, &Setting{}); err != nil {
		t.Fatal(err)
	}
	for _, table := range []interface{}{&Task{}, &TaskTemplate{}} {
		for _, name := range []string{"notify_keyword_line_mode", "notify_success_text", "notify_failure_text"} {
			if err := db.Migrator().DropColumn(table, name); err != nil {
				t.Fatal(err)
			}
		}
	}
	// Seed rows against the old schema, before adding columns.
	if err := db.Omit("NotifyKeywordLineMode", "NotifySuccessText", "NotifyFailureText").Create(&Task{Name: "legacy", Command: "echo ok", Spec: "@daily", NotifyKeyword: "FAIL", NotifyKeywordExclude: "ignored"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Omit("NotifyKeywordLineMode", "NotifySuccessText", "NotifyFailureText").Create(&TaskTemplate{Name: "legacy-template", Command: "echo ok", NotifyKeyword: "FAIL"}).Error; err != nil {
		t.Fatal(err)
	}
	m := new(Migration)
	for i := 0; i < 2; i++ {
		if err := m.upgradeFor1120(db); err != nil {
			t.Fatal(err)
		}
	}
	var task Task
	if err := db.Where("name = ?", "legacy").First(&task).Error; err != nil {
		t.Fatal(err)
	}
	if task.NotifyKeyword != "FAIL" || task.NotifyKeywordExclude != "ignored" || task.NotifyKeywordLineMode != 0 || task.NotifySuccessText != "" || task.NotifyFailureText != "" {
		t.Fatalf("unexpected upgraded task: %+v", task)
	}
	var template TaskTemplate
	if err := db.Where("name = ?", "legacy-template").First(&template).Error; err != nil {
		t.Fatal(err)
	}
	if template.NotifyKeyword != "FAIL" || template.NotifyKeywordLineMode != 0 || template.NotifySuccessText != "" || template.NotifyFailureText != "" {
		t.Fatalf("unexpected upgraded template: %+v", template)
	}
}

func TestUpgradeFor1120TemplateDefaults(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Task{}, &TaskTemplate{}, &Setting{}); err != nil {
		t.Fatal(err)
	}
	oldEmail := "Task ID: {{.TaskId}}\nTask Name: {{.TaskName}}\nStatus: {{.Status}}\nResult: {{.Result}}\nRemark: {{.Remark}}"
	for _, setting := range []Setting{{Code: MailCode, Key: MailTemplateKey, Value: oldEmail}, {Code: SlackCode, Key: SlackTemplateKey, Value: "custom {{.Status}}"}, {Code: WebhookCode, Key: WebhookTemplateKey, Value: webhookTemplate}} {
		if err := db.Create(&setting).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := new(Migration).upgradeFor1120(db); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ code, want string }{{MailCode, oldEmail}, {SlackCode, "custom {{.Status}}"}, {WebhookCode, webhookTemplate}} {
		var setting Setting
		if err := db.Where("code = ? AND key = ?", tc.code, "template").First(&setting).Error; err != nil {
			t.Fatal(err)
		}
		if setting.Value != tc.want {
			t.Errorf("%s: unexpected template %q", tc.code, setting.Value)
		}
	}
}
