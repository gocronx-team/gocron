package models

import "github.com/gocronx-team/gocron/internal/modules/logger"

// RepairSettings 修复缺失的 Setting 配置记录
// 用于解决数据库迁移或升级过程中可能出现的配置缺失问题
func RepairSettings() error {
	logger.Info("Starting to check and repair Setting configuration...")

	// 定义所有必需的配置项
	requiredSettings := []struct {
		Code  string
		Key   string
		Value string
	}{
		// Slack 配置
		{SlackCode, SlackUrlKey, ""},
		{SlackCode, SlackTemplateKey, slackTemplate},

		// 邮件配置
		{MailCode, MailServerKey, ""},
		{MailCode, MailTemplateKey, emailTemplate},

		// Webhook 配置
		{WebhookCode, WebhookUrlKey, ""},
		{WebhookCode, WebhookTemplateKey, webhookTemplate},

		// 系统配置
		{SystemCode, LogRetentionDaysKey, "0"},
		{SystemCode, LogCleanupTimeKey, "03:00"},
		{SystemCode, LogFileSizeLimitKey, "0"},
	}

	// 检查并创建缺失的配置
	for _, cfg := range requiredSettings {
		var count int64
		err := Db.Model(&Setting{}).Where("code = ? AND `key` = ?", cfg.Code, cfg.Key).Count(&count).Error
		if err != nil {
			logger.Error("Failed to check configuration:", err)
			return err
		}

		if count == 0 {
			setting := &Setting{
				Code:  cfg.Code,
				Key:   cfg.Key,
				Value: cfg.Value,
			}
			if err := Db.Create(setting).Error; err != nil {
				logger.Error("Failed to create configuration:", err)
				return err
			}
			logger.Infof("Created missing configuration: code=%s, key=%s", cfg.Code, cfg.Key)
		}
	}

	logger.Info("Setting configuration check completed")
	return nil
}
