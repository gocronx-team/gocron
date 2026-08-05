package models

import (
	"testing"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func TestTaskLogChunkAppendListAndDelete(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&TaskLogChunk{}); err != nil {
		t.Fatal(err)
	}
	old := Db
	Db = db
	defer func() { Db = old }()

	model := new(TaskLogChunk)
	if err := model.Append([]TaskLogChunk{
		{TaskLogId: 7, Seq: 1, Content: "one"},
		{TaskLogId: 7, Seq: 2, Content: "two"},
		{TaskLogId: 8, Seq: 1, Content: "other"},
	}); err != nil {
		t.Fatal(err)
	}
	items, err := model.ListAfter(7, 1, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Seq != 2 || items[0].Content != "two" {
		t.Fatalf("unexpected chunks: %+v", items)
	}
	if err := model.DeleteByTaskLogId(7); err != nil {
		t.Fatal(err)
	}
	items, err = model.ListAfter(7, 0, 100)
	if err != nil || len(items) != 0 {
		t.Fatalf("chunks not deleted: %+v, %v", items, err)
	}
}
