package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/modules/app"
	"github.com/gocronx-team/gocron/internal/routers/manage"
)

func TestIterationRoutesRemainAdminOnly(t *testing.T) {
	old := app.Installed
	app.Installed = true
	t.Cleanup(func() { app.Installed = old })
	gin.SetMode(gin.TestMode)
	for _, path := range []string{"/api/system/metrics", "/api/task/run/1", "/api/task/run-capabilities"} {
		for _, admin := range []int{0, 1} {
			r := gin.New()
			r.Use(func(c *gin.Context) { c.Set("is_admin", admin) }, urlAuth)
			called := false
			r.GET(path, func(c *gin.Context) { called = true; c.Status(http.StatusNoContent) })
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
			if called != (admin == 1) {
				t.Fatalf("path %s admin=%d called=%v", path, admin, called)
			}
		}
	}
}

func TestMetricsBeforeInstallDoesNotExposeProcess(t *testing.T) {
	old := app.Installed
	app.Installed = false
	t.Cleanup(func() { app.Installed = old })
	r := gin.New()
	r.GET("/api/system/metrics", manage.Metrics)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/system/metrics", nil))
	if w.Code != http.StatusForbidden || w.Body.Len() != 0 {
		t.Fatal("pre-install metrics exposed")
	}
}
