package tasklog

// 任务日志

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	rpcClient "github.com/gocronx-team/gocron/internal/modules/rpc/client"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers/base"
	"github.com/gocronx-team/gocron/internal/service"
)

type streamEvent struct {
	Content string        `json:"content"`
	Seq     int64         `json:"seq"`
	Status  models.Status `json:"status"`
	Reset   bool          `json:"reset,omitempty"`
}

// Stream incrementally exposes the persisted task output. The database is the
// catch-up source, so reconnecting clients can resume with the last chunk
// sequence without requiring an in-memory broker.
func Stream(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		base.RespondError(c, i18n.T(c, "invalid_log_id"))
		return
	}
	seq, err := strconv.ParseInt(c.DefaultQuery("seq", "0"), 10, 64)
	if err != nil || seq < 0 {
		base.RespondError(c, i18n.T(c, "invalid_log_id"))
		return
	}

	var initial models.TaskLog
	if err := initial.Find(id); err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return
	}

	ticker := time.NewTicker(time.Second)
	heartbeat := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	defer heartbeat.Stop()

	writeSnapshot := func(log models.TaskLog) bool {
		if log.Status != models.Running {
			payload, _ := json.Marshal(streamEvent{Content: log.Result, Status: log.Status, Reset: true})
			fmt.Fprintf(c.Writer, "event: log\ndata: %s\n\n", payload)
			payload, _ = json.Marshal(streamEvent{Status: log.Status})
			fmt.Fprintf(c.Writer, "event: done\ndata: %s\n\n", payload)
			flusher.Flush()
			return false
		}

		chunks, err := new(models.TaskLogChunk).ListAfter(log.Id, seq, 100)
		if err != nil {
			return false
		}
		for _, chunk := range chunks {
			payload, marshalErr := json.Marshal(streamEvent{Content: chunk.Content, Seq: chunk.Seq, Status: log.Status})
			if marshalErr != nil {
				return false
			}
			fmt.Fprintf(c.Writer, "event: log\ndata: %s\n\n", payload)
			seq = chunk.Seq
		}
		if len(chunks) > 0 {
			flusher.Flush()
		}
		return true
	}

	if !writeSnapshot(initial) {
		return
	}
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			var log models.TaskLog
			if err := log.Find(id); err != nil || !writeSnapshot(log) {
				return
			}
		case <-heartbeat.C:
			fmt.Fprint(c.Writer, ": keep-alive\n\n")
			flusher.Flush()
		}
	}
}

func Index(c *gin.Context) {
	logModel := new(models.TaskLog)
	queryParams := parseQueryParams(c)
	total, err := logModel.Total(queryParams)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	logs, err := logModel.List(queryParams)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	base.RespondSuccess(c, utils.SuccessContent, map[string]interface{}{
		"total": total,
		"data":  logs,
	})
}

// 清空日志
func Clear(c *gin.Context) {
	taskLogModel := new(models.TaskLog)
	_, err := taskLogModel.Clear()
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
	} else {
		base.RespondSuccessWithDefaultMsg(c, nil)
	}
}

type stopRequest struct {
	Id     int64 `form:"id" json:"id" binding:"required,min=1"`
	TaskId int   `form:"task_id" json:"task_id" binding:"required,min=1"`
}

// 停止运行中的任务
func Stop(c *gin.Context) {
	var req stopRequest
	if err := c.ShouldBind(&req); err != nil {
		base.RespondError(c, i18n.T(c, "invalid_log_id"))
		return
	}

	var taskLog models.TaskLog
	if err := taskLog.Find(req.Id); err != nil {
		base.RespondError(c, i18n.T(c, "invalid_log_id"))
		return
	}
	if taskLog.TaskId != req.TaskId {
		base.RespondError(c, i18n.T(c, "invalid_task_id"))
		return
	}
	taskModel := new(models.Task)
	task, err := taskModel.Detail(req.TaskId)
	if err != nil {
		base.RespondError(c, i18n.T(c, "get_task_info_failed")+"#"+err.Error(), err)
		return
	}
	if task.Protocol != models.TaskRPC {
		base.RespondError(c, i18n.T(c, "only_shell_task_can_stop"))
		return
	}
	if len(task.Hosts) == 0 {
		base.RespondError(c, i18n.T(c, "task_node_list_empty"))
		return
	}
	stopFailed := false
	notRunning := 0
	for _, host := range task.Hosts {
		if err := service.ServiceTask.Stop(host.Name, host.Port, req.Id); err != nil {
			if errors.Is(err, rpcClient.ErrTaskNotRunning) {
				notRunning++
			} else {
				stopFailed = true
			}
		}
	}
	if stopFailed {
		base.RespondError(c, i18n.T(c, "stop_task_failed"))
		return
	}
	if notRunning == len(task.Hosts) {
		if err := service.ServiceTask.ReconcileStoppedLog(req.Id); err != nil {
			base.RespondError(c, i18n.T(c, "stop_task_failed"))
			return
		}
		base.RespondSuccess(c, i18n.T(c, "stop_task_reconciled"), nil)
		return
	}

	base.RespondSuccess(c, i18n.T(c, "stop_task_sent"), nil)
}

// 删除N个月前的日志
func Remove(c *gin.Context) {
	month, _ := strconv.Atoi(c.Param("id"))
	if month < 1 || month > 12 {
		base.RespondError(c, i18n.T(c, "param_range_1_12"))
		return
	}
	taskLogModel := new(models.TaskLog)
	_, err := taskLogModel.Remove(month)
	if err != nil {
		base.RespondError(c, i18n.T(c, "delete_failed"), err)
	} else {
		base.RespondSuccess(c, i18n.T(c, "delete_success"), nil)
	}
}

// 清空指定任务的日志
func ClearByTaskId(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		base.RespondError(c, i18n.T(c, "invalid_task_id"))
		return
	}
	taskLogModel := new(models.TaskLog)
	affected, err := taskLogModel.ClearByTaskId(id)
	if err != nil {
		base.RespondError(c, i18n.T(c, "delete_failed"), err)
	} else {
		base.RespondSuccess(c, i18n.T(c, "delete_success"), map[string]interface{}{
			"affected": affected,
		})
	}
}

// 解析查询参数
func parseQueryParams(c *gin.Context) models.CommonMap {
	var params models.CommonMap = models.CommonMap{}
	taskId, _ := strconv.Atoi(c.Query("task_id"))
	protocol, _ := strconv.Atoi(c.Query("protocol"))
	params["TaskId"] = taskId
	params["Protocol"] = protocol
	params["Keyword"] = strings.TrimSpace(c.Query("keyword"))
	// 前端直接传状态枚举值(失败=0/运行中=1/成功=2/取消=3),空值表示不过滤。
	params["Status"] = base.ParseStatusFilter(c.Query("status"))
	base.ParsePageAndPageSize(c, params)

	return params
}
