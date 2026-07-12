package user

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/app"
	"github.com/gocronx-team/gocron/internal/modules/setting"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

// setupUserAuthDb 用内存 sqlite 替换全局 Db，并注入一个测试用的 AuthSecret。
func setupUserAuthDb(t *testing.T) func() {
	t.Helper()
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	origDb := models.Db
	origSetting := app.Setting
	models.Db = db
	app.Setting = &setting.Setting{AuthSecret: "test-secret-abc"}
	return func() {
		models.Db = origDb
		app.Setting = origSetting
	}
}

func ctxWithToken(token string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodGet, "/api/task", nil)
	req.Header.Set("Auth-Token", token)
	c.Request = req
	return c
}

func createUser(t *testing.T, name, email string, isAdmin int8) *models.User {
	t.Helper()
	u := &models.User{Name: name, Email: email, Password: "secret123", IsAdmin: isAdmin}
	if _, err := u.Create(); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

// 启用用户携带有效 token：应通过，并把 uid / is_admin 写入 context。
func TestRestoreToken_EnabledUserPasses(t *testing.T) {
	defer setupUserAuthDb(t)()
	u := createUser(t, "alice", "alice@example.com", 1)
	token, err := generateToken(u)
	if err != nil {
		t.Fatalf("generateToken: %v", err)
	}

	c := ctxWithToken(token)
	if _, err := RestoreToken(c); err != nil {
		t.Fatalf("expected enabled user to pass, got: %v", err)
	}
	if Uid(c) != u.Id {
		t.Fatalf("expected uid %d in context, got %d", u.Id, Uid(c))
	}
	if !IsAdmin(c) {
		t.Fatalf("expected is_admin true in context")
	}
}

// 禁用用户的既有 token：应被拒绝（即时吊销）。
func TestRestoreToken_RejectsDisabledUser(t *testing.T) {
	defer setupUserAuthDb(t)()
	u := createUser(t, "bob", "bob@example.com", 0)
	token, err := generateToken(u)
	if err != nil {
		t.Fatalf("generateToken: %v", err)
	}

	if _, err := (&models.User{}).Disable(u.Id); err != nil {
		t.Fatalf("disable: %v", err)
	}

	c := ctxWithToken(token)
	if _, err := RestoreToken(c); err == nil {
		t.Fatal("expected disabled user's token to be rejected")
	}
}

// 已删除用户的既有 token：应被拒绝。
func TestRestoreToken_RejectsDeletedUser(t *testing.T) {
	defer setupUserAuthDb(t)()
	u := createUser(t, "carol", "carol@example.com", 0)
	token, err := generateToken(u)
	if err != nil {
		t.Fatalf("generateToken: %v", err)
	}

	if _, err := (&models.User{}).Delete(u.Id); err != nil {
		t.Fatalf("delete: %v", err)
	}

	c := ctxWithToken(token)
	if _, err := RestoreToken(c); err == nil {
		t.Fatal("expected deleted user's token to be rejected")
	}
}

// 降权：token 里固化 is_admin=1，但库中已改为 0，context 应取库中的最新值 0。
func TestRestoreToken_UsesFreshRoleAfterDemotion(t *testing.T) {
	defer setupUserAuthDb(t)()
	u := createUser(t, "dave", "dave@example.com", 1)
	token, err := generateToken(u) // 此时 token 的 is_admin=1
	if err != nil {
		t.Fatalf("generateToken: %v", err)
	}

	if _, err := (&models.User{}).Update(u.Id, models.CommonMap{"is_admin": 0}); err != nil {
		t.Fatalf("demote: %v", err)
	}

	c := ctxWithToken(token)
	if _, err := RestoreToken(c); err != nil {
		t.Fatalf("expected demoted-but-enabled user to pass, got: %v", err)
	}
	if IsAdmin(c) {
		t.Fatal("expected is_admin false after demotion (fresh value from DB)")
	}
}
