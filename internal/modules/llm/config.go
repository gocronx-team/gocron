package llm

import "github.com/gocronx-team/gocron/internal/models"

// FromSettings 依据数据库中的 LLM 配置构建客户端；
// 未启用或配置不完整时返回 ErrNotConfigured，调用方据此优雅降级、不阻塞主流程。
func FromSettings() (*Client, error) {
	cfg, err := new(models.Setting).LLM()
	if err != nil {
		return nil, err
	}
	if !cfg.Enable || cfg.BaseURL == "" || cfg.ApiKey == "" || cfg.Model == "" {
		return nil, ErrNotConfigured
	}
	if err := ValidateBaseURL(cfg.BaseURL); err != nil {
		return nil, err
	}
	return New(cfg.BaseURL, cfg.ApiKey, cfg.Model), nil
}
