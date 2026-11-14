package manage

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/modules/app"
)

func TestMetricsBoundedAndNoTaskLabels(t *testing.T) {
	old := app.Installed
	app.Installed = true
	t.Cleanup(func() { app.Installed = old })
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("is_admin", 1)
	Metrics(c)
	if w.Code != 200 || !strings.Contains(w.Header().Get("Content-Type"), "version=0.0.4") {
		t.Fatalf("invalid metrics response: %d %v", w.Code, w.Header())
	}
	body := w.Body.String()
	for _, name := range []string{"gocron_build_info", "gocron_process_uptime_seconds", "gocron_go_goroutines"} {
		if !strings.Contains(body, name) {
			t.Fatalf("missing %s", name)
		}
	}
	if len(body) > 2048 || strings.Contains(body, "task_id") || strings.Contains(body, "hostname") {
		t.Fatal("unexpected metrics cardinality")
	}
}

func BenchmarkMetrics(b *testing.B) {
	old := app.Installed
	app.Installed = true
	b.Cleanup(func() { app.Installed = old })
	gin.SetMode(gin.TestMode)
	b.ReportAllocs()
	for b.Loop() {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("is_admin", 1)
		Metrics(c)
	}
}

func TestMetricsRejectsUnauthenticatedAccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	Metrics(c)
	if w.Code != 403 || w.Body.Len() != 0 {
		t.Fatal("metrics exposed without admin authentication")
	}
}
