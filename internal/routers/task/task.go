package task

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/cron"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/httpclient"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers/base"
	"github.com/gocronx-team/gocron/internal/routers/user"
	"github.com/gocronx-team/gocron/internal/service"
)

type TaskForm struct {
	Id               int                         `form:"id" json:"id"`
	Level            models.TaskLevel            `form:"level" json:"level" binding:"required,oneof=1 2"`
	DependencyStatus models.TaskDependencyStatus `form:"dependency_status" json:"dependency_status" binding:"oneof=1 2"`
	DependencyTaskId string                      `form:"dependency_task_id" json:"dependency_task_id"`
	Name             string                      `form:"name" json:"name" binding:"required,max=32"`
	Spec             string                      `form:"spec" json:"spec"`
	Protocol         models.TaskProtocol         `form:"protocol" json:"protocol" binding:"oneof=1 2"`
	Command          string                      `form:"command" json:"command" binding:"required,max=65535"`
	HttpMethod       models.TaskHTTPMethod       `form:"http_method" json:"http_method" binding:"oneof=1 2"`
	HttpBody         string                      `form:"http_body" json:"http_body" binding:"max=65535"`
	HttpHeaders      string                      `form:"http_headers" json:"http_headers" binding:"max=4096"`
	SuccessPattern   string                      `form:"success_pattern" json:"success_pattern" binding:"max=512"`
	Timeout          int                         `form:"timeout" json:"timeout" binding:"min=0,max=86400"`
	Multi            int8                        `form:"multi" json:"multi" binding:"oneof=0 1"`
	RetryTimes       int8                        `form:"retry_times" json:"retry_times"`
	RetryInterval    int16                       `form:"retry_interval" json:"retry_interval"`
	HostId           string                      `form:"host_id" json:"host_id"`
	Tag              string                      `form:"tag" json:"tag"`
	Remark           string                      `form:"remark" json:"remark"`
	NotifyStatus     int8                        `form:"notify_status" json:"notify_status" binding:"oneof=0 1 2 3"`
	NotifyType       int8                        `form:"notify_type" json:"notify_type" binding:"oneof=0 1 2"`
	NotifyReceiverId string                      `form:"notify_receiver_id" json:"notify_receiver_id"`
	NotifyKeyword    string                      `form:"notify_keyword" json:"notify_keyword"`
	LogRetentionDays int                         `form:"log_retention_days" json:"log_retention_days" binding:"min=0,max=3650"`
}

// 首页
func Index(c *gin.Context) {
	taskModel := new(models.Task)
	queryParams := parseQueryParams(c)
	total, err := taskModel.Total(queryParams)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	tasks, err := taskModel.List(queryParams)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	for i, item := range tasks {
		tasks[i].NextRunTime = models.NextRunTime(service.ServiceTask.NextRunTime(item))
	}
	jsonResp := utils.JsonResponse{}
	result := jsonResp.Success(utils.SuccessContent, map[string]interface{}{
		"total": total,
		"data":  tasks,
	})
	c.String(http.StatusOK, result)
}

// Detail 任务详情
func Detail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}
	taskModel := new(models.Task)
	task, err := taskModel.Detail(id)
	jsonResp := utils.JsonResponse{}
	var result string
	if err != nil || task.Id == 0 {
		logger.Errorf("编辑任务#获取任务详情失败#任务ID-%d", id)
		result = jsonResp.Success(utils.SuccessContent, nil)
	} else {
		result = jsonResp.Success(utils.SuccessContent, task)
	}
	c.String(http.StatusOK, result)
}

// 保存任务
func Store(c *gin.Context) {
	var form TaskForm
	if err := c.ShouldBind(&form); err != nil {
		base.RespondValidationError(c, err)
		return
	}

	taskModel := models.Task{}
	var id = form.Id
	nameExists, err := taskModel.NameExist(form.Name, form.Id)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	if nameExists {
		base.RespondError(c, i18n.T(c, "task_name_exists"))
		return
	}

	if form.Protocol == models.TaskRPC && form.HostId == "" {
		base.RespondError(c, i18n.T(c, "select_hostname"))
		return
	}

	taskModel.Name = form.Name
	taskModel.Protocol = form.Protocol
	// 清理命令中的 HTML 实体编码
	originalCmd := strings.TrimSpace(form.Command)
	cleanedCmd := utils.CleanHTMLEntities(originalCmd)
	if originalCmd != cleanedCmd {
		logger.Infof("[HTML Entity Cleaned] Task: %s, Original length: %d, Cleaned length: %d", form.Name, len(originalCmd), len(cleanedCmd))
	}
	taskModel.Command = cleanedCmd
	taskModel.Timeout = form.Timeout
	taskModel.Tag = form.Tag
	taskModel.Remark = form.Remark
	taskModel.Multi = form.Multi
	taskModel.RetryTimes = form.RetryTimes
	taskModel.RetryInterval = form.RetryInterval
	taskModel.NotifyStatus = form.NotifyStatus
	taskModel.NotifyType = form.NotifyType
	taskModel.NotifyReceiverId = form.NotifyReceiverId
	taskModel.NotifyKeyword = form.NotifyKeyword
	taskModel.LogRetentionDays = form.LogRetentionDays
	taskModel.Spec = form.Spec
	taskModel.Level = form.Level
	taskModel.DependencyStatus = form.DependencyStatus
	taskModel.DependencyTaskId = strings.TrimSpace(form.DependencyTaskId)
	if taskModel.NotifyStatus > 0 && taskModel.NotifyType != 2 && taskModel.NotifyReceiverId == "" {
		base.RespondError(c, i18n.T(c, "select_at_least_one_receiver"))
		return
	}
	taskModel.HttpMethod = form.HttpMethod
	// 校验 HttpHeaders（JSON 格式 + 黑名单检查）
	if err := httpclient.ValidateHeaders(form.HttpHeaders); err != nil {
		base.RespondError(c, "http_headers: "+err.Error())
		return
	}
	taskModel.HttpBody = form.HttpBody
	taskModel.HttpHeaders = form.HttpHeaders
	taskModel.SuccessPattern = form.SuccessPattern
	if taskModel.Protocol == models.TaskHTTP {
		command := strings.ToLower(taskModel.Command)
		if !strings.HasPrefix(command, "http://") && !strings.HasPrefix(command, "https://") {
			base.RespondError(c, i18n.T(c, "invalid_url"))
			return
		}
	}

	if taskModel.RetryTimes > 10 || taskModel.RetryTimes < 0 {
		base.RespondError(c, i18n.T(c, "retry_times_range_0_10"))
		return
	}

	if taskModel.RetryInterval > 3600 || taskModel.RetryInterval < 0 {
		base.RespondError(c, i18n.T(c, "retry_interval_range_0_3600"))
		return
	}

	if taskModel.DependencyStatus != models.TaskDependencyStatusStrong &&
		taskModel.DependencyStatus != models.TaskDependencyStatusWeak {
		base.RespondError(c, i18n.T(c, "select_dependency"))
		return
	}

	if taskModel.Level == models.TaskLevelParent {
		err = utils.PanicToError(func() {
			cron.Parse(form.Spec)
		})
		if err != nil {
			base.RespondError(c, i18n.T(c, "crontab_parse_failed"), err)
			return
		}
	} else {
		taskModel.DependencyTaskId = ""
		taskModel.Spec = ""
	}

	if id > 0 && taskModel.DependencyTaskId != "" {
		dependencyTaskIds := strings.Split(taskModel.DependencyTaskId, ",")
		if utils.InStringSlice(dependencyTaskIds, strconv.Itoa(id)) {
			base.RespondError(c, i18n.T(c, "cannot_set_self_as_child"))
			return
		}
	}

	if id == 0 {
		taskModel.Status = models.Running
		logger.Infof("[Task Create] Before Create - Multi: %d", taskModel.Multi)
		id, err = taskModel.Create()
		if err == nil {
			// 立即读取验证
			verifyTask, _ := taskModel.Detail(id)
			logger.Infof("[Task Create] After Create - ID: %d, Multi in DB: %d", id, verifyTask.Multi)
			// 供审计中间件回填 target（create 时 form.id=0，URL 里也没有 :id）
			c.Set("audit_target_id", id)
			c.Set("audit_target_name", taskModel.Name)
		}
	} else {
		// 更新前记录旧值用于审计 diff
		oldTask, _ := taskModel.Detail(id)

		// 保存脚本版本（命令变更时）
		if oldTask.Command != taskModel.Command {
			versionModel := new(models.TaskScriptVersion)
			latestVersion, _ := versionModel.GetLatestVersion(id)
			newVersion := &models.TaskScriptVersion{
				TaskId:   id,
				Command:  oldTask.Command,
				Username: user.Username(c),
				Version:  latestVersion + 1,
			}
			if _, vErr := newVersion.Create(); vErr != nil {
				logger.Warnf("保存脚本版本失败 TaskID-%d: %v", id, vErr)
			}
			if cErr := versionModel.CleanOldVersions(id, 30); cErr != nil {
				logger.Warnf("清理旧版本失败 TaskID-%d: %v", id, cErr)
			}
		}

		logger.Infof("[Task Update] Before Update - ID: %d, Multi: %d", id, taskModel.Multi)
		_, err = taskModel.UpdateBean(id)
		if err == nil {
			// 立即读取验证
			verifyTask, _ := taskModel.Detail(id)
			logger.Infof("[Task Update] After Update - ID: %d, Multi in DB: %d", id, verifyTask.Multi)

			// 生成审计 diff
			if diff := buildTaskDiff(oldTask, verifyTask); diff != "" {
				c.Set("audit_detail", diff)
			}
		}
	}

	if err != nil {
		base.RespondError(c, i18n.T(c, "save_failed"), err)
		return
	}

	taskHostModel := new(models.TaskHost)
	if form.Protocol == models.TaskRPC {
		hostIdStrList := strings.Split(form.HostId, ",")
		hostIds := make([]int, len(hostIdStrList))
		for i, hostIdStr := range hostIdStrList {
			hostIds[i], _ = strconv.Atoi(hostIdStr)
		}
		_ = taskHostModel.Add(id, hostIds)
	} else {
		_ = taskHostModel.Remove(id)
	}

	status, _ := taskModel.GetStatus(id)
	if status == models.Enabled && taskModel.Level == models.TaskLevelParent {
		addTaskToTimer(id)
	}

	base.RespondSuccess(c, i18n.T(c, "save_success"), nil)
}

// 删除任务
func Remove(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}
	taskModel := new(models.Task)
	_, err = taskModel.Delete(id)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
	} else {
		taskHostModel := new(models.TaskHost)
		_ = taskHostModel.Remove(id)
		service.ServiceTask.Remove(id)
		base.RespondSuccessWithDefaultMsg(c, nil)
	}
}

// 激活任务
func Enable(c *gin.Context) {
	changeStatus(c, models.Enabled)
}

// 暂停任务
func Disable(c *gin.Context) {
	changeStatus(c, models.Disabled)
}

// 手动运行任务
func Run(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}
	taskModel := new(models.Task)
	task, err := taskModel.Detail(id)
	if err != nil || task.Id <= 0 {
		base.RespondError(c, i18n.T(c, "get_task_detail_failed"), err)
	} else {
		task.Spec = i18n.T(c, "manual_run")
		service.ServiceTask.Run(task)
		base.RespondSuccess(c, i18n.T(c, "task_started_check_log"), nil)
	}
}

// 批量启用任务
func BatchEnable(c *gin.Context) {
	batchChangeStatus(c, models.Enabled)
}

// 批量禁用任务
func BatchDisable(c *gin.Context) {
	batchChangeStatus(c, models.Disabled)
}

// 批量改变任务状态
func batchChangeStatus(c *gin.Context, status models.Status) {
	var form struct {
		Ids []int `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}

	taskModel := new(models.Task)
	successCount := 0
	for _, id := range form.Ids {
		_, err := taskModel.Update(id, models.CommonMap{
			"status": status,
		})
		if err == nil {
			successCount++
			if status == models.Enabled {
				addTaskToTimer(id)
			} else {
				service.ServiceTask.Remove(id)
			}
		}
	}

	base.RespondSuccess(c, i18n.T(c, "operation_success"), map[string]interface{}{
		"success_count": successCount,
		"total_count":   len(form.Ids),
	})
}

// 批量删除任务
func BatchRemove(c *gin.Context) {
	var form struct {
		Ids []int `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}

	taskModel := new(models.Task)
	taskHostModel := new(models.TaskHost)
	successCount := 0
	for _, id := range form.Ids {
		_, err := taskModel.Delete(id)
		if err == nil {
			successCount++
			_ = taskHostModel.Remove(id)
			service.ServiceTask.Remove(id)
		}
	}

	base.RespondSuccess(c, i18n.T(c, "operation_success"), map[string]interface{}{
		"success_count": successCount,
		"total_count":   len(form.Ids),
	})
}

// 改变任务状态
func changeStatus(c *gin.Context, status models.Status) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}
	taskModel := new(models.Task)
	_, err = taskModel.Update(id, models.CommonMap{
		"status": status,
	})
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
	} else {
		if status == models.Enabled {
			addTaskToTimer(id)
		} else {
			service.ServiceTask.Remove(id)
		}
		base.RespondSuccessWithDefaultMsg(c, nil)
	}
}

// 添加任务到定时器
func addTaskToTimer(id int) {
	taskModel := new(models.Task)
	task, err := taskModel.Detail(id)
	if err != nil {
		logger.Error(err)
		return
	}

	service.ServiceTask.RemoveAndAdd(task)
}

// GetAllTags 获取所有已使用的标签列表
func GetAllTags(c *gin.Context) {
	taskModel := new(models.Task)
	tags, err := taskModel.GetAllTags()
	if err != nil {
		logger.Error(err)
		tags = []string{}
	}
	jsonResp := utils.JsonResponse{}
	result := jsonResp.Success(utils.SuccessContent, tags)
	c.String(http.StatusOK, result)
}

// 解析查询参数
func parseQueryParams(c *gin.Context) models.CommonMap {
	var params models.CommonMap = models.CommonMap{}
	id, _ := strconv.Atoi(c.Query("id"))
	hostId, _ := strconv.Atoi(c.Query("host_id"))
	protocol, _ := strconv.Atoi(c.Query("protocol"))
	status, _ := strconv.Atoi(c.Query("status"))
	params["Id"] = id
	params["HostId"] = hostId
	params["Name"] = strings.TrimSpace(c.Query("name"))
	params["Protocol"] = protocol
	params["Tag"] = strings.TrimSpace(c.Query("tag"))
	if status >= 0 {
		status -= 1
	}
	params["Status"] = status
	base.ParsePageAndPageSize(c, params)

	return params
}

// buildTaskDiff 对比任务的旧值和新值，返回可读的变更摘要
func buildTaskDiff(old, new models.Task) string {
	type change struct {
		Field string `json:"field"`
		Old   string `json:"old"`
		New   string `json:"new"`
	}
	var changes []change

	add := func(field, oldVal, newVal string) {
		if oldVal != newVal {
			changes = append(changes, change{field, oldVal, newVal})
		}
	}

	add("name", old.Name, new.Name)
	add("spec", old.Spec, new.Spec)
	add("command", old.Command, new.Command)
	add("tag", old.Tag, new.Tag)
	add("timeout", strconv.Itoa(old.Timeout), strconv.Itoa(new.Timeout))
	add("retry_times", strconv.Itoa(int(old.RetryTimes)), strconv.Itoa(int(new.RetryTimes)))
	add("retry_interval", strconv.Itoa(int(old.RetryInterval)), strconv.Itoa(int(new.RetryInterval)))
	add("remark", old.Remark, new.Remark)
	add("http_method", strconv.Itoa(int(old.HttpMethod)), strconv.Itoa(int(new.HttpMethod)))
	add("http_body", old.HttpBody, new.HttpBody)
	add("http_headers", old.HttpHeaders, new.HttpHeaders)
	add("success_pattern", old.SuccessPattern, new.SuccessPattern)
	add("notify_status", strconv.Itoa(int(old.NotifyStatus)), strconv.Itoa(int(new.NotifyStatus)))
	add("notify_keyword", old.NotifyKeyword, new.NotifyKeyword)
	add("log_retention_days", strconv.Itoa(old.LogRetentionDays), strconv.Itoa(new.LogRetentionDays))

	if len(changes) == 0 {
		return ""
	}

	// 生成简洁的文本格式
	var b strings.Builder
	for i, ch := range changes {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(ch.Field)
		b.WriteString(": ")
		b.WriteString(ch.Old)
		b.WriteString(" → ")
		b.WriteString(ch.New)
	}
	return b.String()
}
