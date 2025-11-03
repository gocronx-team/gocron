package host

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
)

type RegisterForm struct {
	Hostname string `json:"hostname" binding:"required"`
	IP       string `json:"ip" binding:"required"`
	Port     int    `json:"port" binding:"required,min=1,max=65535"`
	Alias    string `json:"alias"`
	Version  string `json:"version"`
}

// Register agent自动注册接口
func Register(c *gin.Context) {
	var form RegisterForm
	json := utils.JsonResponse{}
	
	if err := c.ShouldBindJSON(&form); err != nil {
		result := json.CommonFailure("Invalid request parameters")
		c.String(http.StatusOK, result)
		return
	}

	// 验证 token
	token := c.GetHeader("X-Register-Token")
	settingModel := new(models.Setting)
	savedToken := settingModel.GetAgentRegisterToken()
	
	if savedToken == "" {
		result := json.CommonFailure("Registration token not configured")
		c.String(http.StatusOK, result)
		return
	}
	
	if token != savedToken {
		logger.Warnf("Invalid registration token from %s", form.IP)
		result := json.CommonFailure("Invalid registration token")
		c.String(http.StatusOK, result)
		return
	}

	hostModel := new(models.Host)
	
	// 使用 IP 作为唯一标识，检查是否已存在
	params := models.CommonMap{"Name": form.IP}
	hosts, err := hostModel.List(params)
	
	if err != nil {
		logger.Error("Query host failed:", err)
		result := json.CommonFailure("Registration failed")
		c.String(http.StatusOK, result)
		return
	}

	alias := form.Alias
	if alias == "" {
		alias = form.Hostname
	}

	// 如果已存在，更新信息
	if len(hosts) > 0 {
		existHost := hosts[0]
		existHost.Port = form.Port
		existHost.Alias = alias
		existHost.Remark = "Auto-registered, version: " + form.Version
		
		_, err = existHost.UpdateBean(existHost.Id)
		if err != nil {
			logger.Error("Update host failed:", err)
			result := json.CommonFailure("Update failed")
			c.String(http.StatusOK, result)
			return
		}
		
		logger.Infof("Agent updated: %s:%d (alias: %s)", form.IP, form.Port, alias)
		result := json.Success("Agent updated successfully", nil)
		c.String(http.StatusOK, result)
		return
	}

	// 不存在则创建新记录
	hostModel.Name = strings.TrimSpace(form.IP)
	hostModel.Alias = strings.TrimSpace(alias)
	hostModel.Port = form.Port
	hostModel.Remark = "Auto-registered, version: " + form.Version

	_, err = hostModel.Create()
	if err != nil {
		logger.Error("Create host failed:", err)
		result := json.CommonFailure("Registration failed")
		c.String(http.StatusOK, result)
		return
	}

	logger.Infof("Agent registered: %s:%d (alias: %s)", form.IP, form.Port, alias)
	result := json.Success("Agent registered successfully", nil)
	c.String(http.StatusOK, result)
}
