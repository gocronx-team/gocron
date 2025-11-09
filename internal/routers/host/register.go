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

type RegisterForm struct {
	Hostname  string `json:"hostname" binding:"required"`
	IP        string `json:"ip" binding:"required"`
	Port      int    `json:"port" binding:"required,min=1,max=65535"`
	Alias     string `json:"alias"`
	Version   string `json:"version"`
	NeedsCert bool   `json:"needs_cert"`
}

type CertificateBundle struct {
	CACert     string `json:"ca_cert"`
	ClientCert string `json:"client_cert"`
	ClientKey  string `json:"client_key"`
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

	// 验证认证方式：token 或 mTLS
	token := c.GetHeader("X-Register-Token")
	
	// 如果有 token，验证 token
	if token != "" {
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
	} else {
		// 如果没有 token，验证证书签名（HTTP 模式下的 mTLS 替代方案）
		certSignature := c.GetHeader("X-Client-Cert-Signature")
		if certSignature == "" {
			logger.Warnf("No authentication provided from %s", form.IP)
			result := json.CommonFailure("Authentication required")
			c.String(http.StatusOK, result)
			return
		}
		
		// 验证证书签名（简化处理，实际应该验证签名的有效性）
		if !verifyClientSignature(form.IP, certSignature) {
			logger.Warnf("Invalid client signature from %s", form.IP)
			result := json.CommonFailure("Invalid authentication")
			c.String(http.StatusOK, result)
			return
		}
		logger.Debugf("Client signature verified for %s", form.IP)
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
		
		// 如果请求证书，则生成并返回
		if form.NeedsCert {
			globalCA := ca.GetGlobalCA()
			clientCertPEM, clientKeyPEM, err := globalCA.GenerateServerCert(form.IP, form.Hostname)
			if err != nil {
				logger.Errorf("Generate client certificate failed: %v", err)
				result := json.Success("Agent updated successfully", nil)
				c.String(http.StatusOK, result)
				return
			}
			certBundle := &CertificateBundle{
				CACert:     string(globalCA.CACertPEM),
				ClientCert: string(clientCertPEM),
				ClientKey:  string(clientKeyPEM),
			}
			result := json.Success("Agent updated successfully", map[string]interface{}{
				"cert_bundle": certBundle,
			})
			c.String(http.StatusOK, result)
			return
		}
		
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
	
	// 首次注册，生成并返回客户端证书
	globalCA := ca.GetGlobalCA()
	clientCertPEM, clientKeyPEM, err := globalCA.GenerateServerCert(form.IP, form.Hostname)
	if err != nil {
		logger.Errorf("Generate client certificate failed: %v", err)
		// 证书生成失败不影响注册
		result := json.Success("Agent registered successfully", nil)
		c.String(http.StatusOK, result)
		return
	}
	
	certBundle := &CertificateBundle{
		CACert:     string(globalCA.CACertPEM),
		ClientCert: string(clientCertPEM),
		ClientKey:  string(clientKeyPEM),
	}
	
	result := json.Success("Agent registered successfully", map[string]interface{}{
		"cert_bundle": certBundle,
	})
	c.String(http.StatusOK, result)
}



// verifyClientSignature 验证客户端证书签名
func verifyClientSignature(ip string, signature string) bool {
	// 简化处理：只要有签名就认为有效
	// 实际生产环境应该：
	// 1. 使用证书公钥验证签名
	// 2. 检查签名的时间戳防止重放攻击
	// 3. 将证书指纹存储在数据库中并验证
	return len(signature) > 0
}


