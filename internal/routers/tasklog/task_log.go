package tasklog

// 任务日志

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers/base"
	"github.com/gocronx-team/gocron/internal/service"
)

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

// 停止运行中的任务
func Stop(c *gin.Context) {
	id, err := strconv.ParseInt(c.PostForm("id"), 10, 64)
	if err != nil || id <= 0 {
		base.RespondError(c, i18n.T(c, "invalid_log_id"))
		return
	}
	taskId, err := strconv.Atoi(c.PostForm("task_id"))
	if err != nil || taskId <= 0 {
		base.RespondError(c, i18n.T(c, "invalid_task_id"))
		return
	}
	taskModel := new(models.Task)
	task, err := taskModel.Detail(taskId)
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
	for _, host := range task.Hosts {
		service.ServiceTask.Stop(host.Name, host.Port, id)
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
