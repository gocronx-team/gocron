package secret

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/crypto"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func setupTestRouter(t *testing.T) (*gin.Engine, func()) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	// 配置加密主密钥（crypto.Init 读取后清除环境变量；多次调用幂等）。
	_ = os.Setenv(crypto.SecretKeyEnv, "router-test-key")
	crypto.Init()
	if !crypto.Configured() {
		t.Fatal("crypto not configured for test")
	}

	original := models.Db
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Secret{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	models.Db = db

	r := gin.New()
	r.GET("/api/secret", Index)
	r.POST("/api/secret/store", Store)
	r.POST("/api/secret/remove/:id", Remove)

	return r, func() { models.Db = original }
}

type envelope struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

func postForm(r *gin.Engine, path string, form url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeCode(t *testing.T, w *httptest.ResponseRecorder) int {
	t.Helper()
	var env envelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v (body=%s)", err, w.Body.String())
	}
	return env.Code
}

func TestStoreCreateEncryptsValue(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	w := postForm(r, "/api/secret/store", url.Values{
		"name":  {"API_KEY"},
		"value": {"super-secret-123"},
	})
	if code := decodeCode(t, w); code != 0 {
		t.Fatalf("expected success code 0, got %d", code)
	}

	// 落库的应是密文,且能解密回原值。
	all, err := (&models.Secret{}).All()
	if err != nil || len(all) != 1 {
		t.Fatalf("expected 1 stored secret, got %d (err=%v)", len(all), err)
	}
	if all[0].Value == "super-secret-123" {
		t.Error("value stored in plaintext, expected ciphertext")
	}
	plain, err := crypto.Decrypt(all[0].Value)
	if err != nil || plain != "super-secret-123" {
		t.Errorf("decrypt mismatch: %q (err=%v)", plain, err)
	}
}

func TestStoreInvalidName(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	for _, bad := range []string{"123KEY", "has space", "bad-dash", "with.dot"} {
		w := postForm(r, "/api/secret/store", url.Values{"name": {bad}, "value": {"v"}})
		if code := decodeCode(t, w); code == 0 {
			t.Errorf("expected failure for invalid name %q, got success", bad)
		}
	}
}

func TestStoreDuplicateName(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	postForm(r, "/api/secret/store", url.Values{"name": {"DUP"}, "value": {"v"}})
	w := postForm(r, "/api/secret/store", url.Values{"name": {"DUP"}, "value": {"v2"}})
	if code := decodeCode(t, w); code == 0 {
		t.Error("expected failure for duplicate name, got success")
	}
}

func TestStoreValueRequiredOnCreate(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	w := postForm(r, "/api/secret/store", url.Values{"name": {"NOVAL"}})
	if code := decodeCode(t, w); code == 0 {
		t.Error("expected failure when value missing on create, got success")
	}
}

func TestStoreUpdateKeepsValueWhenEmpty(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	postForm(r, "/api/secret/store", url.Values{"name": {"K"}, "value": {"original"}})
	created, _ := (&models.Secret{}).All()
	id := created[0].Id

	// 不传 value,只改备注:值应保持不变。
	w := postForm(r, "/api/secret/store", url.Values{
		"id":     {strconv.Itoa(id)},
		"name":   {"K"},
		"remark": {"updated"},
	})
	if code := decodeCode(t, w); code != 0 {
		t.Fatalf("update failed, code %d", code)
	}
	got := &models.Secret{}
	_ = got.Find(id)
	plain, _ := crypto.Decrypt(got.Value)
	if plain != "original" {
		t.Errorf("value changed unexpectedly: %q", plain)
	}
	if got.Remark != "updated" {
		t.Errorf("remark not updated: %q", got.Remark)
	}

	// 传新 value:应更新。
	postForm(r, "/api/secret/store", url.Values{
		"id":    {strconv.Itoa(id)},
		"name":  {"K"},
		"value": {"changed"},
	})
	_ = got.Find(id)
	plain, _ = crypto.Decrypt(got.Value)
	if plain != "changed" {
		t.Errorf("value not updated: %q", plain)
	}
}

func TestIndexExcludesValue(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	postForm(r, "/api/secret/store", url.Values{"name": {"LISTED"}, "value": {"hidden"}})

	req := httptest.NewRequest(http.MethodGet, "/api/secret", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if strings.Contains(body, "hidden") {
		t.Errorf("Index response leaked secret value: %s", body)
	}
	if !strings.Contains(body, "LISTED") {
		t.Errorf("Index should list the secret name, got: %s", body)
	}
}

func TestRemove(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()

	postForm(r, "/api/secret/store", url.Values{"name": {"DELME"}, "value": {"v"}})
	created, _ := (&models.Secret{}).All()
	w := postForm(r, "/api/secret/remove/"+strconv.Itoa(created[0].Id), url.Values{})
	if code := decodeCode(t, w); code != 0 {
		t.Fatalf("remove failed, code %d", code)
	}
	remaining, _ := (&models.Secret{}).All()
	if len(remaining) != 0 {
		t.Errorf("expected 0 secrets after remove, got %d", len(remaining))
	}
}

func TestStoreRejectsReservedName(t *testing.T) {
	r, cleanup := setupTestRouter(t)
	defer cleanup()
	for _, name := range []string{"PATH", "LD_PRELOAD", "home"} {
		w := postForm(r, "/api/secret/store", url.Values{"name": {name}, "value": {"somevalue"}})
		if code := decodeCode(t, w); code == 0 {
			t.Errorf("expected reserved name %q to be rejected, got success", name)
		}
	}
}
