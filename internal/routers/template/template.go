package template

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers/base"
	"github.com/gocronx-team/gocron/internal/routers/user"
)

type TemplateForm struct {
	Id          int    `form:"id" json:"id"`
	Name        string `form:"name" json:"name" binding:"required,max=64"`
	Description string `form:"description" json:"description" binding:"max=500"`
	Category    string `form:"category" json:"category" binding:"required,max=32"`
	Protocol    int8   `form:"protocol" json:"protocol" binding:"oneof=1 2"`
	Command     string `form:"command" json:"command" binding:"required,max=65535"`
	HttpMethod  int8   `form:"http_method" json:"http_method" binding:"oneof=1 2"`
	HttpBody       string `form:"http_body" json:"http_body"`
	HttpHeaders    string `form:"http_headers" json:"http_headers"`
	SuccessPattern string `form:"success_pattern" json:"success_pattern" binding:"max=512"`
	Tag            string `form:"tag" json:"tag"`
	Spec           string `form:"spec" json:"spec"`
	Timeout        int    `form:"timeout" json:"timeout" binding:"min=0,max=86400"`
	Multi            int8   `form:"multi" json:"multi" binding:"oneof=0 1"`
	RetryTimes       int8   `form:"retry_times" json:"retry_times"`
	RetryInterval    int16  `form:"retry_interval" json:"retry_interval"`
	Timezone         string `form:"timezone" json:"timezone"`
	NotifyStatus     int8   `form:"notify_status" json:"notify_status"`
	NotifyType       int8   `form:"notify_type" json:"notify_type"`
	NotifyKeyword    string `form:"notify_keyword" json:"notify_keyword"`
	LogRetentionDays int    `form:"log_retention_days" json:"log_retention_days" binding:"min=0,max=3650"`
}

type SaveFromTaskForm struct {
	TaskId      int    `form:"task_id" json:"task_id" binding:"required"`
	Name        string `form:"name" json:"name" binding:"required,max=64"`
	Description string `form:"description" json:"description" binding:"max=500"`
	Category    string `form:"category" json:"category" binding:"required,max=32"`
}

// Index 模板列表
func Index(c *gin.Context) {
	tmplModel := new(models.TaskTemplate)
	params := parseQueryParams(c)

	total, err := tmplModel.Total(params)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}

	list, err := tmplModel.List(params)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}

	jsonResp := utils.JsonResponse{}
	result := jsonResp.Success(utils.SuccessContent, map[string]interface{}{
		"total": total,
		"data":  list,
	})
	c.String(http.StatusOK, result)
}

// Detail 模板详情
func Detail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}

	tmplModel := new(models.TaskTemplate)
	tmpl, err := tmplModel.Detail(id)
	if err != nil || tmpl.Id == 0 {
		base.RespondError(c, i18n.T(c, "template_not_found"))
		return
	}

	jsonResp := utils.JsonResponse{}
	c.String(http.StatusOK, jsonResp.Success(utils.SuccessContent, tmpl))
}

// Store 创建/更新模板
func Store(c *gin.Context) {
	var form TemplateForm
	if err := c.ShouldBind(&form); err != nil {
		base.RespondValidationError(c, err)
		return
	}

	tmplModel := models.TaskTemplate{}
	id := form.Id

	// 内置模板不可修改
	if id > 0 {
		existing, detailErr := tmplModel.Detail(id)
		if detailErr != nil || existing.Id == 0 {
			base.RespondError(c, i18n.T(c, "template_not_found"))
			return
		}
		if existing.IsBuiltin == 1 {
			base.RespondError(c, i18n.T(c, "builtin_template_readonly"))
			return
		}
	}

	nameExists, err := tmplModel.NameExist(form.Name, id)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	if nameExists {
		base.RespondError(c, i18n.T(c, "template_name_exists"))
		return
	}

	tmplModel.Name = form.Name
	tmplModel.Description = form.Description
	tmplModel.Category = form.Category
	tmplModel.Protocol = form.Protocol
	tmplModel.Command = form.Command
	tmplModel.HttpMethod = form.HttpMethod
	tmplModel.HttpBody = form.HttpBody
	tmplModel.HttpHeaders = form.HttpHeaders
	tmplModel.SuccessPattern = form.SuccessPattern
	tmplModel.Tag = form.Tag
	tmplModel.Spec = form.Spec
	tmplModel.Timeout = form.Timeout
	tmplModel.Multi = form.Multi
	tmplModel.RetryTimes = form.RetryTimes
	tmplModel.RetryInterval = form.RetryInterval
	tmplModel.Timezone = form.Timezone
	tmplModel.NotifyStatus = form.NotifyStatus
	tmplModel.NotifyType = form.NotifyType
	tmplModel.NotifyKeyword = form.NotifyKeyword
	tmplModel.LogRetentionDays = form.LogRetentionDays

	if id == 0 {
		tmplModel.CreatedBy = user.Username(c)
		_, err = tmplModel.Create()
	} else {
		_, err = tmplModel.UpdateBean(id)
	}

	if err != nil {
		base.RespondError(c, i18n.T(c, "save_failed"), err)
		return
	}

	base.RespondSuccess(c, i18n.T(c, "save_success"), nil)
}

// Remove 删除模板
func Remove(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	tmplModel := new(models.TaskTemplate)

	tmpl, err := tmplModel.Detail(id)
	if err != nil || tmpl.Id == 0 {
		base.RespondError(c, i18n.T(c, "template_not_found"))
		return
	}

	if tmpl.IsBuiltin == 1 {
		base.RespondError(c, i18n.T(c, "builtin_template_no_delete"))
		return
	}

	_, err = tmplModel.Delete(id)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}

	base.RespondSuccessWithDefaultMsg(c, nil)
}

// Apply 应用模板（增加使用次数并返回模板数据）
func Apply(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	tmplModel := new(models.TaskTemplate)

	tmpl, err := tmplModel.Detail(id)
	if err != nil || tmpl.Id == 0 {
		base.RespondError(c, i18n.T(c, "template_not_found"))
		return
	}

	if uErr := tmplModel.IncrementUsage(id); uErr != nil {
		logger.Warnf("增加模板使用次数失败 TemplateID-%d: %v", id, uErr)
	}

	jsonResp := utils.JsonResponse{}
	result := jsonResp.Success(utils.SuccessContent, tmpl)
	c.String(http.StatusOK, result)
}

// SaveFromTask 从现有任务保存为模板
func SaveFromTask(c *gin.Context) {
	var form SaveFromTaskForm
	if err := c.ShouldBind(&form); err != nil {
		base.RespondValidationError(c, err)
		return
	}

	taskModel := new(models.Task)
	task, err := taskModel.Detail(form.TaskId)
	if err != nil || task.Id == 0 {
		base.RespondError(c, i18n.T(c, "task_not_found"))
		return
	}

	tmplModel := models.TaskTemplate{}
	nameExists, err := tmplModel.NameExist(form.Name, 0)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	if nameExists {
		base.RespondError(c, i18n.T(c, "template_name_exists"))
		return
	}

	tmplModel.Name = form.Name
	tmplModel.Description = form.Description
	tmplModel.Category = form.Category
	tmplModel.Protocol = int8(task.Protocol)
	tmplModel.Command = task.Command
	tmplModel.HttpMethod = int8(task.HttpMethod)
	tmplModel.HttpBody = task.HttpBody
	tmplModel.HttpHeaders = task.HttpHeaders
	tmplModel.SuccessPattern = task.SuccessPattern
	tmplModel.Tag = task.Tag
	// 从 spec 中解析 timezone（格式: CRON_TZ=Asia/Shanghai 0 0 2 * * *）
	spec := task.Spec
	if strings.HasPrefix(spec, "CRON_TZ=") || strings.HasPrefix(spec, "TZ=") {
		parts := strings.SplitN(spec, " ", 2)
		if len(parts) == 2 {
			tzPart := parts[0]
			spec = parts[1]
			tmplModel.Timezone = strings.SplitN(tzPart, "=", 2)[1]
		}
	}
	tmplModel.Spec = spec
	tmplModel.Timeout = task.Timeout
	tmplModel.Multi = task.Multi
	tmplModel.RetryTimes = task.RetryTimes
	tmplModel.RetryInterval = task.RetryInterval
	tmplModel.NotifyStatus = task.NotifyStatus
	tmplModel.NotifyType = task.NotifyType
	tmplModel.NotifyKeyword = task.NotifyKeyword
	tmplModel.LogRetentionDays = task.LogRetentionDays
	tmplModel.CreatedBy = user.Username(c)

	_, err = tmplModel.Create()
	if err != nil {
		base.RespondError(c, i18n.T(c, "save_failed"), err)
		return
	}

	base.RespondSuccess(c, i18n.T(c, "save_success"), nil)
}

// Categories 获取所有分类
func Categories(c *gin.Context) {
	tmplModel := new(models.TaskTemplate)
	categories, err := tmplModel.GetCategories()
	if err != nil {
		categories = []string{}
	}

	jsonResp := utils.JsonResponse{}
	result := jsonResp.Success(utils.SuccessContent, categories)
	c.String(http.StatusOK, result)
}

func parseQueryParams(c *gin.Context) models.CommonMap {
	params := models.CommonMap{}
	params["Category"] = strings.TrimSpace(c.Query("category"))
	params["Name"] = strings.TrimSpace(c.Query("name"))
	base.ParsePageAndPageSize(c, params)
	return params
}
