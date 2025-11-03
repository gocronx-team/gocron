package host

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
)

// GetRegisterToken 获取注册 token
func GetRegisterToken(c *gin.Context) {
	settingModel := new(models.Setting)
	token := settingModel.GetAgentRegisterToken()
	
	json := utils.JsonResponse{}
	result := json.Success(i18n.T(c, "operation_success"), map[string]string{
		"token": token,
	})
	c.String(http.StatusOK, result)
}

// GenerateRegisterToken 生成新的注册 token
func GenerateRegisterToken(c *gin.Context) {
	token := generateRandomToken(32)
	
	settingModel := new(models.Setting)
	err := settingModel.UpdateAgentRegisterToken(token)
	
	json := utils.JsonResponse{}
	if err != nil {
		logger.Error("Generate token failed:", err)
		result := json.CommonFailure(i18n.T(c, "operation_failed"))
		c.String(http.StatusOK, result)
		return
	}
	
	result := json.Success(i18n.T(c, "operation_success"), map[string]string{
		"token": token,
	})
	c.String(http.StatusOK, result)
}

func generateRandomToken(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
