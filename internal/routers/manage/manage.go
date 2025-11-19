package manage

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/service"
)

func Slack(c *gin.Context) {
	settingModel := new(models.Setting)
	slack, err := settingModel.Slack()
	jsonResp := utils.JsonResponse{}
	var result string
	if err != nil {
		logger.Error(err)
		result = jsonResp.Success(utils.SuccessContent, nil)
	} else {
		result = jsonResp.Success(utils.SuccessContent, slack)
	}
	c.String(http.StatusOK, result)
}

func UpdateSlack(c *gin.Context) {
	var form UpdateSlackForm
	if err := c.ShouldBind(&form); err != nil {
		logger.Errorf("Slack配置表单验证失败: %v", err)
		json := utils.JsonResponse{}
		result := json.CommonFailure("表单验证失败, 请检测输入")
		c.String(http.StatusOK, result)
		return
	}

	settingModel := new(models.Setting)
	err := settingModel.UpdateSlack(form.Url, form.Template)
	result := utils.JsonResponseByErr(err)
	c.String(http.StatusOK, result)
}

func CreateSlackChannel(c *gin.Context) {
	var form CreateSlackChannelForm
	if err := c.ShouldBind(&form); err != nil {
		logger.Errorf("创建Slack频道表单验证失败: %v", err)
		json := utils.JsonResponse{}
		result := json.CommonFailure("表单验证失败, 请检测输入")
		c.String(http.StatusOK, result)
		return
	}

	settingModel := new(models.Setting)
	var result string
	if settingModel.IsChannelExist(form.Channel) {
		jsonResp := utils.JsonResponse{}
		result = jsonResp.CommonFailure("Channel已存在")
	} else {
		_, err := settingModel.CreateChannel(form.Channel)
		result = utils.JsonResponseByErr(err)
	}
	c.String(http.StatusOK, result)
}

func RemoveSlackChannel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	settingModel := new(models.Setting)
	_, err := settingModel.RemoveChannel(id)
	result := utils.JsonResponseByErr(err)
	c.String(http.StatusOK, result)
}

// endregion

// region 邮件
func Mail(c *gin.Context) {
	settingModel := new(models.Setting)
	mail, err := settingModel.Mail()
	jsonResp := utils.JsonResponse{}
	var result string
	if err != nil {
		logger.Error(err)
		result = jsonResp.Success(utils.SuccessContent, nil)
	} else {
		result = jsonResp.Success("", mail)
	}
	c.String(http.StatusOK, result)
}

type MailServerForm struct {
	Host     string `form:"host" json:"host" binding:"required,max=100"`
	Port     int    `form:"port" json:"port" binding:"required,min=1,max=65535"`
	User     string `form:"user" json:"user" binding:"required,max=64"`
	Password string `form:"password" json:"password" binding:"required,max=64"`
	Template string `form:"template" json:"template"`
}

// CreateMailUserForm 创建邮件用户表单
type CreateMailUserForm struct {
	Username string `form:"username" json:"username" binding:"required,max=50"`
	Email    string `form:"email" json:"email" binding:"required,email,max=100"`
}

// UpdateSlackForm 更新Slack配置表单
type UpdateSlackForm struct {
	Url      string `form:"url" json:"url" binding:"required,url,max=200"`
	Template string `form:"template" json:"template" binding:"required"`
}

// UpdateWebHookForm 更新WebHook配置表单
type UpdateWebHookForm struct {
	Url      string `form:"url" json:"url" binding:"required,url,max=200"`
	Template string `form:"template" json:"template" binding:"required"`
}

// CreateSlackChannelForm 创建Slack频道表单
type CreateSlackChannelForm struct {
	Channel string `form:"channel" json:"channel" binding:"required,max=50"`
}

func UpdateMail(c *gin.Context) {
	var form MailServerForm
	if err := c.ShouldBind(&form); err != nil {
		logger.Errorf("邮件配置表单验证失败: %v", err)
		json := utils.JsonResponse{}
		// 提供更具体的错误信息
		errorMsg := "表单验证失败: "
		if strings.Contains(err.Error(), "email") {
			errorMsg += "用户名必须是有效的邮箱地址"
		} else if strings.Contains(err.Error(), "required") {
			errorMsg += "请填写所有必填字段"
		} else if strings.Contains(err.Error(), "max") {
			errorMsg += "输入内容过长"
		} else if strings.Contains(err.Error(), "min") || strings.Contains(err.Error(), "port") {
			errorMsg += "端口号必须在1-65535之间"
		} else {
			errorMsg += "请检查输入格式"
		}
		result := json.CommonFailure(errorMsg)
		c.String(http.StatusOK, result)
		return
	}

	// 从表单中提取template，单独保存
	template := strings.TrimSpace(form.Template)

	// 将服务器配置序列化为JSON（不包含template）
	serverConfig := struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
	}{
		Host:     form.Host,
		Port:     form.Port,
		User:     form.User,
		Password: form.Password,
	}
	jsonByte, _ := json.Marshal(serverConfig)

	settingModel := new(models.Setting)
	err := settingModel.UpdateMail(string(jsonByte), template)
	result := utils.JsonResponseByErr(err)
	c.String(http.StatusOK, result)
}

func CreateMailUser(c *gin.Context) {
	var form CreateMailUserForm
	if err := c.ShouldBind(&form); err != nil {
		logger.Errorf("创建邮件用户表单验证失败: %v", err)
		json := utils.JsonResponse{}
		result := json.CommonFailure("表单验证失败, 请检测输入")
		c.String(http.StatusOK, result)
		return
	}

	settingModel := new(models.Setting)
	_, err := settingModel.CreateMailUser(form.Username, form.Email)
	result := utils.JsonResponseByErr(err)
	c.String(http.StatusOK, result)
}

func RemoveMailUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	settingModel := new(models.Setting)
	_, err := settingModel.RemoveMailUser(id)
	result := utils.JsonResponseByErr(err)
	c.String(http.StatusOK, result)
}

func WebHook(c *gin.Context) {
	settingModel := new(models.Setting)
	webHook, err := settingModel.Webhook()
	jsonResp := utils.JsonResponse{}
	var result string
	if err != nil {
		logger.Error(err)
		result = jsonResp.Success(utils.SuccessContent, nil)
	} else {
		result = jsonResp.Success("", webHook)
	}
	c.String(http.StatusOK, result)
}

func UpdateWebHook(c *gin.Context) {
	var form UpdateWebHookForm
	if err := c.ShouldBind(&form); err != nil {
		logger.Errorf("Webhook配置表单验证失败: %v", err)
		json := utils.JsonResponse{}
		result := json.CommonFailure("表单验证失败, 请检测输入")
		c.String(http.StatusOK, result)
		return
	}

	settingModel := new(models.Setting)
	err := settingModel.UpdateWebHook(form.Url, form.Template)
	result := utils.JsonResponseByErr(err)
	c.String(http.StatusOK, result)
}

// endregion

// region 系统配置
func GetLogRetentionDays(c *gin.Context) {
	settingModel := new(models.Setting)
	days := settingModel.GetLogRetentionDays()
	cleanupTime := settingModel.GetLogCleanupTime()
	fileSizeLimit := settingModel.GetLogFileSizeLimit()
	jsonResp := utils.JsonResponse{}
	result := jsonResp.Success("", map[string]interface{}{
		"days":            days,
		"cleanup_time":    cleanupTime,
		"file_size_limit": fileSizeLimit,
	})
	c.String(http.StatusOK, result)
}

func UpdateLogRetentionDays(c *gin.Context) {
	var form struct {
		Days          int    `json:"days" binding:"min=0,max=3650"`
		CleanupTime   string `json:"cleanup_time" binding:"required"`
		FileSizeLimit int    `json:"file_size_limit" binding:"min=0,max=10240"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		json := utils.JsonResponse{}
		result := json.CommonFailure("表单验证失败, 请检测输入")
		c.String(http.StatusOK, result)
		return
	}

	settingModel := new(models.Setting)
	err := settingModel.UpdateLogRetentionDays(form.Days)
	if err != nil {
		result := utils.JsonResponseByErr(err)
		c.String(http.StatusOK, result)
		return
	}
	err = settingModel.UpdateLogCleanupTime(form.CleanupTime)
	if err != nil {
		result := utils.JsonResponseByErr(err)
		c.String(http.StatusOK, result)
		return
	}
	err = settingModel.UpdateLogFileSizeLimit(form.FileSizeLimit)
	if err != nil {
		result := utils.JsonResponseByErr(err)
		c.String(http.StatusOK, result)
		return
	}
	// 重新加载日志清理任务
	service.ServiceTask.ReloadLogCleanupTask()
	result := utils.JsonResponseByErr(nil)
	c.String(http.StatusOK, result)
}

// endregion
