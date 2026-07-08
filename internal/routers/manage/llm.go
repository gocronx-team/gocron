package manage

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/llm"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers/base"
)

// llmTestTimeout 连接测试的超时:只发一句极短对话,不必等太久。
const llmTestTimeout = 30 * time.Second

type updateLLMForm struct {
	Enable  bool   `json:"enable"`
	BaseURL string `json:"base_url"`
	ApiKey  string `json:"api_key"`
	Model   string `json:"model"`
}

// LLM 返回大模型配置。出于安全，绝不回传 api_key 明文，仅返回是否已配置。
func LLM(c *gin.Context) {
	settingModel := new(models.Setting)
	cfg, err := settingModel.LLM()
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	base.RespondSuccess(c, utils.SuccessContent, gin.H{
		"enable":      cfg.Enable,
		"base_url":    cfg.BaseURL,
		"model":       cfg.Model,
		"api_key_set": cfg.ApiKey != "",
	})
}

// UpdateLLM 更新大模型配置。api_key 留空表示不修改，沿用已保存的值。
func UpdateLLM(c *gin.Context) {
	var form updateLLMForm
	if err := c.ShouldBindJSON(&form); err != nil {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}

	settingModel := new(models.Setting)
	apiKey := strings.TrimSpace(form.ApiKey)
	if apiKey == "" {
		existing, err := settingModel.LLM()
		if err != nil {
			base.RespondErrorWithDefaultMsg(c, err)
			return
		}
		apiKey = existing.ApiKey
	}

	if err := settingModel.UpdateLLM(form.Enable, strings.TrimSpace(form.BaseURL), apiKey, strings.TrimSpace(form.Model)); err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	base.RespondSuccessWithDefaultMsg(c, nil)
}

// TestLLM 用当前表单配置发一句极短对话,验证 base_url/api_key/model 是否可用。
// api_key 留空则用已保存的值(与 UpdateLLM 的"留空不修改"一致),方便只改其他字段时测试。
// 失败时透出 provider 的真实错误(api_key 已在 llm client 层脱敏),便于定位配置问题。
func TestLLM(c *gin.Context) {
	var form updateLLMForm
	if err := c.ShouldBindJSON(&form); err != nil {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}
	baseURL := strings.TrimSpace(form.BaseURL)
	model := strings.TrimSpace(form.Model)
	apiKey := strings.TrimSpace(form.ApiKey)
	if apiKey == "" {
		if existing, err := new(models.Setting).LLM(); err == nil {
			apiKey = existing.ApiKey
		}
	}
	if baseURL == "" || model == "" || apiKey == "" {
		base.RespondError(c, i18n.T(c, "llm_test_incomplete"))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), llmTestTimeout)
	defer cancel()
	if _, err := llm.New(baseURL, apiKey, model).Chat(ctx, "Connection test.", "Reply with the single word OK."); err != nil {
		base.RespondError(c, i18n.T(c, "llm_test_failed")+": "+err.Error())
		return
	}
	base.RespondSuccess(c, i18n.T(c, "llm_test_success"), nil)
}
