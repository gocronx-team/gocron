package template

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestLegacyTemplateFormPreservesNotificationOptions(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		t.Fatal(err)
	}
	old := models.Db
	models.Db = db
	defer func() { models.Db = old }()
	if err := db.AutoMigrate(&models.TaskTemplate{}); err != nil {
		t.Fatal(err)
	}
	tpl := models.TaskTemplate{Name: "existing", Category: "custom", Protocol: 2, Command: "echo ok", NotifyKeywordLineMode: 1, NotifySuccessText: "成功", NotifyFailureText: "失败"}
	if err := db.Create(&tpl).Error; err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/template/store", Store)
	submit := func(v url.Values) {
		req := httptest.NewRequest(http.MethodPost, "/template/store", strings.NewReader(v.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var response struct {
			Code int `json:"code"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || response.Code != 0 {
			t.Fatalf("store: %s %v", w.Body.String(), err)
		}
	}
	v := url.Values{"id": {strconv.Itoa(tpl.Id)}, "name": {"existing"}, "category": {"custom"}, "protocol": {"2"}, "command": {"echo ok"}, "http_method": {"1"}}
	submit(v) // N-1 frontend omits the three fields.
	var loaded models.TaskTemplate
	if err := db.First(&loaded, tpl.Id).Error; err != nil {
		t.Fatal(err)
	}
	if loaded.NotifyKeywordLineMode != 1 || loaded.NotifySuccessText != "成功" || loaded.NotifyFailureText != "失败" {
		t.Fatalf("legacy update lost options: %+v", loaded)
	}
	v.Set("notify_keyword_line_mode", "0")
	v.Set("notify_success_text", "")
	v.Set("notify_failure_text", "")
	submit(v) // New frontend explicitly clears values.
	if err := db.First(&loaded, tpl.Id).Error; err != nil {
		t.Fatal(err)
	}
	if loaded.NotifyKeywordLineMode != 0 || loaded.NotifySuccessText != "" || loaded.NotifyFailureText != "" {
		t.Fatalf("new update cannot clear options: %+v", loaded)
	}
}
