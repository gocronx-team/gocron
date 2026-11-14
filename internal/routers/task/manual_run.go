package task

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/routers/base"
	"github.com/gocronx-team/gocron/internal/service"
)

func RunCapabilities(c *gin.Context) {
	base.RespondSuccess(c, "", gin.H{"http_post_body_override": true})
}

// manualRunBody returns a copy; persisted and scheduled task configuration is untouched.
func manualRunBody(task models.Task, body string) (models.Task, bool) {
	if task.Protocol != models.TaskHTTP || task.HttpMethod != models.TaskHttpMethodPost ||
		len(body) > 65535 || strings.TrimSpace(body) == "" || !json.Valid([]byte(body)) {
		return task, false
	}
	task.HttpBody = body
	return task, true
}

func RunWithBody(c *gin.Context) {
	runWithBody(c, service.ServiceTask.Run)
}

func runWithBody(c *gin.Context, run func(models.Task)) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 400<<10)
	var form struct {
		HTTPBody *string `json:"http_body" binding:"required"`
	}
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&form); err != nil || form.HTTPBody == nil {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}
	task, err := new(models.Task).Detail(id)
	if err != nil || task.Id <= 0 {
		base.RespondError(c, i18n.T(c, "get_task_detail_failed"), err)
		return
	}
	task, valid := manualRunBody(task, *form.HTTPBody)
	if !valid {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}
	task.Spec = i18n.T(c, "manual_run")
	c.Set("audit_target_name", task.Name)
	c.Set("audit_detail", "manual HTTP run with temporary body (body omitted)")
	run(task)
	c.Set("manual_run_accepted", true)
	base.RespondSuccess(c, i18n.T(c, "task_started_check_log"), nil)
}
