package tasklog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed chan struct{}
}

func (r *flushRecorder) Flush() {
	r.ResponseRecorder.Flush()
	select {
	case r.flushed <- struct{}{}:
	default:
	}
}

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestDb(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	err = db.AutoMigrate(&models.TaskLog{}, &models.TaskLogChunk{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	models.Db = db
}

func TestClearByTaskId_InvalidId(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"non-numeric", "abc"},
		{"negative", "-1"},
		{"zero", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)
			r.POST("/api/task/log/clear/:id", ClearByTaskId)

			c.Request, _ = http.NewRequest("POST", "/api/task/log/clear/"+tt.id, nil)
			r.ServeHTTP(w, c.Request)

			body := w.Body.String()
			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}
			// The response should indicate failure (code != 0)
			if !strings.Contains(body, `"code"`) {
				t.Errorf("expected JSON response with code field, got: %s", body)
			}
			// Should not contain success indicators for invalid input
			if strings.Contains(body, `"code":0`) {
				t.Errorf("expected error response for invalid id %q, got success: %s", tt.id, body)
			}
		})
	}
}

func TestClearByTaskId_ValidId(t *testing.T) {
	setupTestDb(t)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/api/task/log/clear/:id", ClearByTaskId)

	req, _ := http.NewRequest("POST", "/api/task/log/clear/1", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	// Should contain a successful JSON response
	if !strings.Contains(body, `"code":0`) {
		t.Errorf("expected success response for valid id, got: %s", body)
	}
}

func TestStopBindsJSONAndFormRequests(t *testing.T) {
	setupTestDb(t)
	log := models.TaskLog{Id: 42, TaskId: 1, Name: "running", Spec: "*", Command: "sleep 10", Status: models.Running}
	if _, err := log.Create(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		contentType string
		body        string
	}{
		{name: "json", contentType: "application/json", body: `{"id":` + strconv.FormatInt(log.Id, 10) + `,"task_id":2}`},
		{name: "form", contentType: "application/x-www-form-urlencoded", body: "id=" + strconv.FormatInt(log.Id, 10) + "&task_id=2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/api/task/log/stop", Stop)
			req := httptest.NewRequest(http.MethodPost, "/api/task/log/stop", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			req.Header.Set("Accept-Language", "en-US")
			r.ServeHTTP(w, req)

			if !strings.Contains(w.Body.String(), "Invalid task ID") {
				t.Fatalf("request was not bound as %s: %s", tt.contentType, w.Body.String())
			}
		})
	}
}

func TestStreamTerminalLogReturnsSnapshotAndDone(t *testing.T) {
	setupTestDb(t)
	log := models.TaskLog{Id: 1, TaskId: 1, Name: "stream", Spec: "* * * * *", Command: "echo hi", Result: "hello world", Status: models.Finish}
	if _, err := log.Create(); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/api/task/log/:id/stream", Stream)
	req := httptest.NewRequest(http.MethodGet, "/api/task/log/"+strconv.FormatInt(log.Id, 10)+"/stream?seq=0", nil)
	r.ServeHTTP(w, req)

	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("content type = %q", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, `event: log`) || !strings.Contains(body, `"content":"hello world"`) || !strings.Contains(body, `"reset":true`) {
		t.Fatalf("missing incremental log event: %s", body)
	}
	if !strings.Contains(body, `event: done`) {
		t.Fatalf("missing done event: %s", body)
	}
}

func TestStreamRunningLogReturnsChunksAfterSequence(t *testing.T) {
	setupTestDb(t)
	log := models.TaskLog{Id: 1, TaskId: 1, Name: "stream", Spec: "*", Command: "echo", Status: models.Running}
	if _, err := log.Create(); err != nil {
		t.Fatal(err)
	}
	if err := new(models.TaskLogChunk).Append([]models.TaskLogChunk{
		{TaskLogId: log.Id, Seq: 1, Content: "old"},
		{TaskLogId: log.Id, Seq: 2, Content: "new"},
	}); err != nil {
		t.Fatal(err)
	}

	recorder := &flushRecorder{ResponseRecorder: httptest.NewRecorder(), flushed: make(chan struct{}, 1)}
	_, r := gin.CreateTestContext(recorder)
	r.GET("/api/task/log/:id/stream", Stream)
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/api/task/log/1/stream?seq=1", nil).WithContext(ctx)
	done := make(chan struct{})
	go func() {
		r.ServeHTTP(recorder, req)
		close(done)
	}()
	select {
	case <-recorder.flushed:
		cancel()
	case <-time.After(time.Second):
		cancel()
		t.Fatal("stream did not flush initial chunks")
	}
	<-done
	body := recorder.Body.String()
	if strings.Contains(body, `"content":"old"`) || !strings.Contains(body, `"content":"new"`) || !strings.Contains(body, `"seq":2`) {
		t.Fatalf("unexpected stream body: %s", body)
	}
}
