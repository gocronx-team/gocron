package task

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-yaml"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func setupImportExportRouter(t *testing.T) (*gin.Engine, func()) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	// gocron 使用单数表名(SingularTable),测试 db 须一致,否则表名不匹配
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Task{}, &models.TaskHost{}, &models.Host{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	original := models.Db
	models.Db = db
	r := gin.New()
	r.GET("/api/task/export", Export)
	r.POST("/api/task/import", Import)
	return r, func() { models.Db = original }
}

type envelope struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

func TestImportCreatesTasksAndHostAssoc(t *testing.T) {
	r, cleanup := setupImportExportRouter(t)
	defer cleanup()

	if _, err := (&models.Host{Name: "node1", Port: 5921}).Create(); err != nil {
		t.Fatalf("seed host: %v", err)
	}

	body := `version: 1
tasks:
  - name: imp-http
    level: 1
    spec: "0 0 0 1 1 *"
    protocol: 1
    command: "http://example.com/ping"
    http_method: 1
  - name: imp-shell
    level: 1
    spec: "0 0 0 1 1 *"
    protocol: 2
    command: "echo hi"
    hosts: ["node1", "ghost-host"]
`
	req := httptest.NewRequest(http.MethodPost, "/api/task/import", strings.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var env envelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v (body=%s)", err, w.Body.String())
	}
	if env.Code != 0 {
		t.Fatalf("expected success, got code %d (%s)", env.Code, w.Body.String())
	}
	var res ImportResult
	_ = json.Unmarshal(env.Data, &res)
	if res.Created != 2 {
		t.Errorf("expected 2 created, got %d (msgs=%v)", res.Created, res.Messages)
	}

	list, _ := (&models.Task{}).List(models.CommonMap{"Page": 1, "PageSize": 100})
	if len(list) != 2 {
		t.Fatalf("expected 2 tasks in db, got %d", len(list))
	}
	// imp-shell 应关联到存在的 node1(ghost-host 被跳过)
	for _, tk := range list {
		if tk.Name == "imp-shell" {
			if len(tk.Hosts) != 1 || tk.Hosts[0].Name != "node1" {
				t.Errorf("imp-shell hosts = %+v, want [node1]", tk.Hosts)
			}
		}
	}
}

func TestImportSkipsDuplicateAndInvalid(t *testing.T) {
	r, cleanup := setupImportExportRouter(t)
	defer cleanup()

	// 预置一个同名任务
	if _, err := (&models.Task{Name: "dup", Level: models.TaskLevelParent, Spec: "0 0 0 1 1 *",
		Protocol: models.TaskHTTP, Command: "http://x", HttpMethod: models.TaskHTTPMethodGet,
		DependencyStatus: models.TaskDependencyStatusWeak, Status: models.Enabled}).Create(); err != nil {
		t.Fatalf("seed task: %v", err)
	}

	body := `version: 1
tasks:
  - name: dup
    level: 1
    spec: "0 0 0 1 1 *"
    protocol: 1
    command: "http://y"
  - name: bad-proto
    level: 1
    protocol: 9
    command: "x"
  - name: no-cmd
    level: 1
    protocol: 1
`
	req := httptest.NewRequest(http.MethodPost, "/api/task/import", strings.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var env envelope
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	var res ImportResult
	_ = json.Unmarshal(env.Data, &res)
	if res.Created != 0 {
		t.Errorf("expected 0 created (all invalid/dup), got %d", res.Created)
	}
	if res.Skipped != 3 {
		t.Errorf("expected 3 skipped, got %d (msgs=%v)", res.Skipped, res.Messages)
	}
}

func TestExportRoundTrip(t *testing.T) {
	r, cleanup := setupImportExportRouter(t)
	defer cleanup()

	if _, err := (&models.Task{Name: "exp-1", Level: models.TaskLevelParent, Spec: "0 0 0 1 1 *",
		Protocol: models.TaskHTTP, Command: "http://z", HttpMethod: models.TaskHTTPMethodGet,
		NotifyStatus: 5, NotifyKeyword: "ERR", NotifyKeywordRegex: 1,
		DependencyStatus: models.TaskDependencyStatusWeak, Status: models.Enabled}).Create(); err != nil {
		t.Fatalf("seed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/task/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}

	var doc yamlDoc
	if err := yaml.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatalf("export not valid yaml: %v", err)
	}
	if len(doc.Tasks) != 1 || doc.Tasks[0].Name != "exp-1" {
		t.Fatalf("unexpected export: %+v", doc.Tasks)
	}
	// 关键字段(含新的 notify 位掩码 + 正则)应完整保留
	yt := doc.Tasks[0]
	if yt.NotifyStatus != 5 || yt.NotifyKeyword != "ERR" || yt.NotifyKeywordRegex != 1 {
		t.Errorf("notify fields not preserved: %+v", yt)
	}
}
