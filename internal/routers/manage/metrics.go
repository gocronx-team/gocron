package manage

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/modules/app"
	"github.com/gocronx-team/gocron/internal/routers/user"
)

var metricsStartedAt = time.Now()

// Metrics provides bounded, process-local Prometheus text without database
// scans or task/host labels. The API's normal administrator auth applies.
func Metrics(c *gin.Context) {
	if !app.Installed || !user.IsAdmin(c) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(fmt.Sprintf(`# HELP gocron_build_info Application build information.
# TYPE gocron_build_info gauge
gocron_build_info{version=%q} 1
# HELP gocron_process_uptime_seconds Time since metrics package initialization.
# TYPE gocron_process_uptime_seconds gauge
gocron_process_uptime_seconds %.3f
# HELP gocron_go_goroutines Number of goroutines in this process.
# TYPE gocron_go_goroutines gauge
gocron_go_goroutines %d
`, app.AppVersion, time.Since(metricsStartedAt).Seconds(), runtime.NumGoroutine())))
}
