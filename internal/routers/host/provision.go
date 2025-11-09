package host

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/ca"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
)

type ProvisionForm struct {
	Hostname string `json:"hostname" binding:"required"`
	IP       string `json:"ip" binding:"required"`
}

type ServerCertificateBundle struct {
	CACert     string `json:"ca_cert"`
	ServerCert string `json:"server_cert"`
	ServerKey  string `json:"server_key"`
}

// Provision 节点配置接口 - 获取证书并注册节点
func Provision(c *gin.Context) {
	var form ProvisionForm
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
	
	if savedToken == "" || token != savedToken {
		logger.Warnf("Invalid provision token from %s", form.IP)
		result := json.CommonFailure("Invalid token")
		c.String(http.StatusOK, result)
		return
	}

	// 使用统一的 CA 生成服务端证书
	globalCA := ca.GetGlobalCA()
	serverCertPEM, serverKeyPEM, err := globalCA.GenerateServerCert(form.IP, form.Hostname)
	if err != nil {
		logger.Errorf("Generate server certificate failed: %v", err)
		result := json.CommonFailure("Failed to generate certificates")
		c.String(http.StatusOK, result)
		return
	}

	certBundle := &ServerCertificateBundle{
		CACert:     string(globalCA.CACertPEM),
		ServerCert: string(serverCertPEM),
		ServerKey:  string(serverKeyPEM),
	}

	// 检查主机是否已注册
	hostModel := new(models.Host)
	exists, err := hostModel.NameExists(strings.TrimSpace(form.IP), 0)
	if err != nil {
		logger.Errorf("Check host existence failed: %v", err)
		result := json.CommonFailure("Failed to check host")
		c.String(http.StatusOK, result)
		return
	}
	
	if exists {
		logger.Warnf("Host already registered: %s", form.IP)
		result := json.CommonFailure("Host already registered")
		c.String(http.StatusOK, result)
		return
	}

	// 注册节点
	hostModel.Name = strings.TrimSpace(form.IP)
	hostModel.Alias = strings.TrimSpace(form.Hostname)
	hostModel.Port = 5921
	hostModel.Remark = "Auto-provisioned"

	_, err = hostModel.Create()
	if err != nil {
		logger.Errorf("Create host failed: %v", err)
		result := json.CommonFailure("Failed to register host")
		c.String(http.StatusOK, result)
		return
	}

	logger.Infof("Agent provisioned: %s:%d (alias: %s)", form.IP, 5921, form.Hostname)
	
	result := json.Success("Provisioned successfully", map[string]interface{}{
		"cert_bundle": certBundle,
	})
	c.String(http.StatusOK, result)
}
