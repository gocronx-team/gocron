package service

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gocronx-team/gocron/internal/models"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func TestLiveLogAccumulatorMasksSecretAcrossChunks(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.TaskLog{}, &models.TaskLogChunk{}); err != nil {
		t.Fatal(err)
	}
	models.Db = db
	log := models.TaskLog{Id: 1, TaskId: 1, Name: "secret", Spec: "*", Command: "echo", Status: models.Running}
	id, err := log.Create()
	if err != nil {
		t.Fatal(err)
	}
	acc := newLiveLogAccumulator(id, map[string]string{"TOKEN": "hunter2"})
	acc.append("value=hun")
	acc.append("ter2\n")
	acc.close()
	chunks, err := new(models.TaskLogChunk).ListAfter(id, 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	for _, chunk := range chunks {
		output.WriteString(chunk.Content)
	}
	if strings.Contains(output.String(), "hunter2") || !strings.Contains(output.String(), "***") {
		t.Fatalf("secret was not masked: %q", output.String())
	}
}

func TestLimitTaskLog(t *testing.T) {
	result := limitTaskLog(strings.Repeat("界", maxTaskLogBytes), false)
	if len(result) > maxTaskLogBytes || !utf8.ValidString(result) || !strings.HasSuffix(result, taskLogTruncatedMark) {
		t.Fatalf("invalid limited output: bytes=%d valid=%v", len(result), utf8.ValidString(result))
	}
}

func TestStreamingSecretMaskerWithholdsPartialSecret(t *testing.T) {
	masker := newStreamingSecretMasker([]string{"hunter2", "token"})
	var output strings.Builder
	output.WriteString(masker.write("prefix hun"))
	if strings.Contains(output.String(), "hun") {
		t.Fatalf("partial secret was released early: %q", output.String())
	}
	output.WriteString(masker.write("ter2 suffix"))
	output.WriteString(masker.flush())
	if got := output.String(); got != "prefix *** suffix" {
		t.Fatalf("masked output = %q", got)
	}
}

func TestLiveLogAccumulatorCapsPersistedChunks(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.TaskLogChunk{}); err != nil {
		t.Fatal(err)
	}
	old := models.Db
	models.Db = db
	defer func() { models.Db = old }()

	acc := newLiveLogAccumulator(9, nil)
	acc.append(strings.Repeat("x", maxTaskLogBytes+100))
	acc.close()
	chunks, err := new(models.TaskLogChunk).ListAfter(9, 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	for _, chunk := range chunks {
		output.WriteString(chunk.Content)
	}
	if output.Len() > maxTaskLogBytes || !strings.HasSuffix(output.String(), taskLogTruncatedMark) {
		t.Fatalf("unexpected capped output: bytes=%d", output.Len())
	}
}

func TestReconcileStoppedLogFinalizesChunksWithoutOverwritingTerminalLog(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.TaskLog{}, &models.TaskLogChunk{}); err != nil {
		t.Fatal(err)
	}
	old := models.Db
	models.Db = db
	defer func() { models.Db = old }()

	log := models.TaskLog{Id: 42, TaskId: 1, Name: "orphan", Spec: "*", Command: "echo", Status: models.Running}
	if _, err := log.Create(); err != nil {
		t.Fatal(err)
	}
	if err := new(models.TaskLogChunk).Append([]models.TaskLogChunk{
		{TaskLogId: 42, Seq: 1, Content: "first\n"},
		{TaskLogId: 42, Seq: 2, Content: "second\n"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := ServiceTask.ReconcileStoppedLog(42); err != nil {
		t.Fatal(err)
	}
	var got models.TaskLog
	if err := got.Find(42); err != nil {
		t.Fatal(err)
	}
	if got.Status != models.Cancel || got.Result != "first\nsecond\n" {
		t.Fatalf("reconciled log = status %d, result %q", got.Status, got.Result)
	}
	chunks, err := new(models.TaskLogChunk).ListAfter(42, 0, 100)
	if err != nil || len(chunks) != 0 {
		t.Fatalf("chunks after reconciliation = %v, %v", chunks, err)
	}

	if _, err := got.Update(42, models.CommonMap{"status": models.Finish, "result": "final"}); err != nil {
		t.Fatal(err)
	}
	if err := ServiceTask.ReconcileStoppedLog(42); err != nil {
		t.Fatal(err)
	}
	if err := got.Find(42); err != nil || got.Status != models.Finish || got.Result != "final" {
		t.Fatalf("terminal log was overwritten: %+v, %v", got, err)
	}
}
