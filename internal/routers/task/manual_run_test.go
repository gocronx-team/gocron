package task

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/service"
)

func TestManualRunBodyCopyAndValidation(t *testing.T) {
	original := models.Task{Protocol: models.TaskHTTP, HttpMethod: models.TaskHttpMethodPost, HttpBody: `{"scheduled":true}`}
	for _, body := range []string{`{"date":"2026-10-01"}`, `[]`, `null`, `""`} {
		got, ok := manualRunBody(original, body)
		if !ok || got.HttpBody != body || original.HttpBody != `{"scheduled":true}` {
			t.Fatalf("invalid copy override: %+v, %v", got, ok)
		}
		got.HttpBody = original.HttpBody
		if !reflect.DeepEqual(got, original) {
			t.Fatal("override changed other task fields")
		}
	}
	for _, body := range []string{"", " ", "{", strings.Repeat("x", 65536)} {
		if _, ok := manualRunBody(original, body); ok {
			t.Fatalf("accepted invalid body %q", body[:min(len(body), 20)])
		}
	}
	for _, method := range []models.TaskHTTPMethod{0, models.TaskHTTPMethodGet} {
		task := original
		task.HttpMethod = method
		if _, ok := manualRunBody(task, `{}`); ok {
			t.Fatal("accepted non-POST task")
		}
	}
	task := original
	task.Protocol = models.TaskRPC
	if _, ok := manualRunBody(task, `{}`); ok {
		t.Fatal("accepted RPC task")
	}
}

func TestManualRunRealHTTPLeavesScheduledBodyUnchanged(t *testing.T) {
	_, cleanup := setupTestRouter(t)
	defer cleanup()
	if err := models.Db.AutoMigrate(&models.TaskHost{}, &models.Host{}); err != nil {
		t.Fatal(err)
	}
	var bodies []string
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(io.LimitReader(r.Body, 65536))
		bodies = append(bodies, string(body))
		_, _ = w.Write([]byte("ok"))
	}))
	defer target.Close()
	task := models.Task{Name: "real-http", Protocol: models.TaskHTTP, HttpMethod: models.TaskHttpMethodPost, HttpBody: `{"scheduled":true}`, Command: target.URL, Timeout: 2}
	id, err := task.Create()
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(id)}}
	c.Request = httptest.NewRequest("POST", "/api/task/run/1", strings.NewReader(`{"http_body":"{\"manual\":true}"}`))
	handler := &service.HTTPHandler{}
	runWithBody(c, func(copy models.Task) {
		if _, err := handler.Run(copy, 1, nil); err != nil {
			t.Fatal(err)
		}
	})
	saved, err := new(models.Task).Detail(id)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := handler.Run(saved, 2, nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(bodies, []string{`{"manual":true}`, `{"scheduled":true}`}) {
		t.Fatalf("unexpected real HTTP bodies: %v", bodies)
	}
}

func TestRunWithBodyRejectsInvalidEnvelopeBeforeLookup(t *testing.T) {
	for _, body := range []string{`{}`, `{"http_body":null}`, `{"http_body":3}`, `{"http_body":"{}"} {}`, strings.Repeat(" ", 400<<10) + `{}`} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Request = httptest.NewRequest("POST", "/api/task/run/1", strings.NewReader(body))
		RunWithBody(c)
		var response apiResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || response.Code == 0 {
			t.Fatalf("invalid envelope accepted: %v %s", err, w.Body.String())
		}
	}
}

func TestRunWithBodyDispatchesCopyAndPreservesSavedTask(t *testing.T) {
	_, cleanup := setupTestRouter(t)
	defer cleanup()
	if err := models.Db.AutoMigrate(&models.TaskHost{}, &models.Host{}); err != nil {
		t.Fatal(err)
	}
	original := models.Task{Name: "manual", Protocol: models.TaskHTTP, HttpMethod: models.TaskHttpMethodPost, HttpBody: `{"scheduled":true}`, Spec: "0 * * * *", Command: "https://example.com", RetryTimes: 2, Multi: 0}
	id, err := original.Create()
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(id)}}
	c.Request = httptest.NewRequest("POST", "/api/task/run/1", strings.NewReader(`{"http_body":"{\"manual\":true}"}`))
	calls := 0
	runWithBody(c, func(task models.Task) {
		calls++
		if task.HttpBody != `{"manual":true}` || task.RetryTimes != 2 || task.Multi != 0 {
			t.Fatalf("incorrect dispatch: %+v", task)
		}
	})
	if calls != 1 {
		t.Fatalf("expected one dispatch, got %d: %s", calls, w.Body.String())
	}
	var saved models.Task
	if err := models.Db.First(&saved, id).Error; err != nil || saved.HttpBody != original.HttpBody || saved.Spec != original.Spec {
		t.Fatalf("manual run mutated saved configuration: %+v %v", saved, err)
	}
	if detail, _ := c.Get("audit_detail"); strings.Contains(fmt.Sprint(detail), "manual\":true") {
		t.Fatal("body leaked into audit")
	}
}
