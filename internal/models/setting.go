package models

import (
	"encoding/json"
	"strconv"
)

type Setting struct {
	Id    int    `gorm:"primaryKey;autoIncrement"`
	Code  string `gorm:"type:varchar(32);not null"`
	Key   string `gorm:"type:varchar(64);not null"`
	Value string `gorm:"type:varchar(4096);not null;default:''"`
}

// DefaultNotificationTemplate is shared by email and Slack. Keep .Status in
// storage so previous-version instances can render it during rolling upgrades.
const DefaultNotificationTemplate = `Task ID: {{.TaskId}}
Task Name: {{.TaskName}}
Status: {{.Status}}
Result: {{.Result}}
Remark: {{.Remark}}`

const slackTemplate = DefaultNotificationTemplate
const emailTemplate = DefaultNotificationTemplate
const webhookTemplate = `
{
  "task_id": "{{.TaskId}}",
  "task_name": "{{.TaskName}}",
  "status": "{{.Status}}",
  "result": "{{.Result}}",
  "remark": "{{.Remark}}"
}
`

const (
	SlackCode        = "slack"
	SlackUrlKey      = "url"
	SlackTemplateKey = "template"
	SlackChannelKey  = "channel"
)

const (
	MailCode        = "mail"
	MailTemplateKey = "template"
	MailServerKey   = "server"
	MailUserKey     = "user"
)

const (
	WebhookCode        = "webhook"
	WebhookTemplateKey = "template"
	WebhookUrlKey      = "url"
)

const (
	SystemCode          = "system"
	LogRetentionDaysKey = "log_retention_days"
	LogCleanupTimeKey   = "log_cleanup_time"
	LogFileSizeLimitKey = "log_file_size_limit"
)

const (
	LLMCode       = "llm"
	LLMEnableKey  = "enable"
	LLMBaseURLKey = "base_url"
	LLMApiKeyKey  = "api_key"
	LLMModelKey   = "model"
)

// region slack配置

type Slack struct {
	Url      string    `json:"url"`
	Channels []Channel `json:"channels"`
	Template string    `json:"template"`
}

type Channel struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (setting *Setting) Slack() (Slack, error) {
	list := make([]Setting, 0)
	err := Db.Where("code = ?", SlackCode).Find(&list).Error
	slack := Slack{}
	if err != nil {
		return slack, err
	}

	setting.formatSlack(list, &slack)

	return slack, err
}

func (setting *Setting) formatSlack(list []Setting, slack *Slack) {
	for _, v := range list {
		switch v.Key {
		case SlackUrlKey:
			slack.Url = v.Value
		case SlackTemplateKey:
			slack.Template = v.Value
		default:
			slack.Channels = append(slack.Channels, Channel{
				v.Id, v.Value,
			})
		}
	}
}

func (setting *Setting) UpdateSlack(url, template string) error {
	setting.Value = url
	Db.Model(&Setting{}).Where(map[string]interface{}{"code": SlackCode, "key": SlackUrlKey}).Update("value", url)

	setting.Value = template
	Db.Model(&Setting{}).Where(map[string]interface{}{"code": SlackCode, "key": SlackTemplateKey}).Update("value", template)

	return nil
}

// 创建slack渠道
func (setting *Setting) CreateChannel(channel string) (int64, error) {
	setting.Code = SlackCode
	setting.Key = SlackChannelKey
	setting.Value = channel

	result := Db.Create(setting)
	return result.RowsAffected, result.Error
}

func (setting *Setting) IsChannelExist(channel string) bool {
	var count int64
	Db.Model(&Setting{}).Where(map[string]interface{}{"code": SlackCode, "key": SlackChannelKey, "value": channel}).Count(&count)
	return count > 0
}

// 删除slack渠道
func (setting *Setting) RemoveChannel(id int) (int64, error) {
	result := Db.Where(map[string]interface{}{"code": SlackCode, "key": SlackChannelKey, "id": id}).Delete(&Setting{})
	return result.RowsAffected, result.Error
}

// endregion

type Mail struct {
	Host      string     `json:"host"`
	Port      int        `json:"port"`
	User      string     `json:"user"`
	Password  string     `json:"password"`
	MailUsers []MailUser `json:"mail_users"`
	Template  string     `json:"template"`
}

type MailUser struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// region 邮件配置
func (setting *Setting) Mail() (Mail, error) {
	list := make([]Setting, 0)
	err := Db.Where("code = ?", MailCode).Find(&list).Error
	mail := Mail{MailUsers: make([]MailUser, 0)}
	if err != nil {
		return mail, err
	}

	setting.formatMail(list, &mail)

	return mail, err
}

func (setting *Setting) formatMail(list []Setting, mail *Mail) {
	mailUser := MailUser{}
	for _, v := range list {
		switch v.Key {
		case MailServerKey:
			if v.Value != "" {
				_ = json.Unmarshal([]byte(v.Value), mail)
			}
		case MailUserKey:
			if v.Value != "" {
				_ = json.Unmarshal([]byte(v.Value), &mailUser)
				mailUser.Id = v.Id
				mail.MailUsers = append(mail.MailUsers, mailUser)
			}
		case MailTemplateKey:
			mail.Template = v.Value
		}

	}
}

func (setting *Setting) UpdateMail(config, template string) error {
	Db.Model(&Setting{}).Where(map[string]interface{}{"code": MailCode, "key": MailServerKey}).Update("value", config)
	Db.Model(&Setting{}).Where(map[string]interface{}{"code": MailCode, "key": MailTemplateKey}).Update("value", template)

	return nil
}

func (setting *Setting) CreateMailUser(username, email string) (int64, error) {
	setting.Code = MailCode
	setting.Key = MailUserKey
	mailUser := MailUser{0, username, email}
	jsonByte, err := json.Marshal(mailUser)
	if err != nil {
		return 0, err
	}
	setting.Value = string(jsonByte)

	result := Db.Create(setting)
	return result.RowsAffected, result.Error
}

func (setting *Setting) RemoveMailUser(id int) (int64, error) {
	result := Db.Where(map[string]interface{}{"code": MailCode, "key": MailUserKey, "id": id}).Delete(&Setting{})
	return result.RowsAffected, result.Error
}

type WebHook struct {
	WebhookUrls []WebhookUrl `json:"webhook_urls"`
	Template    string       `json:"template"`
}

type WebhookUrl struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

func (setting *Setting) Webhook() (WebHook, error) {
	list := make([]Setting, 0)
	err := Db.Where("code = ?", WebhookCode).Find(&list).Error
	webHook := WebHook{WebhookUrls: make([]WebhookUrl, 0)}
	if err != nil {
		return webHook, err
	}

	setting.formatWebhook(list, &webHook)

	return webHook, err
}

func (setting *Setting) formatWebhook(list []Setting, webHook *WebHook) {
	webhookUrl := WebhookUrl{}
	for _, v := range list {
		switch v.Key {
		case WebhookUrlKey:
			if v.Value != "" {
				_ = json.Unmarshal([]byte(v.Value), &webhookUrl)
				webhookUrl.Id = v.Id
				webHook.WebhookUrls = append(webHook.WebhookUrls, webhookUrl)
			}
		case WebhookTemplateKey:
			webHook.Template = v.Value
		}
	}
}

func (setting *Setting) UpdateWebHook(template string) error {
	Db.Model(&Setting{}).Where(map[string]interface{}{"code": WebhookCode, "key": WebhookTemplateKey}).Update("value", template)
	return nil
}

func (setting *Setting) CreateWebhookUrl(name, url string) (int64, error) {
	webhookUrl := WebhookUrl{0, name, url}
	jsonByte, err := json.Marshal(webhookUrl)
	if err != nil {
		return 0, err
	}

	newSetting := Setting{
		Code:  WebhookCode,
		Key:   WebhookUrlKey,
		Value: string(jsonByte),
	}

	result := Db.Create(&newSetting)
	return result.RowsAffected, result.Error
}

func (setting *Setting) RemoveWebhookUrl(id int) (int64, error) {
	result := Db.Where(map[string]interface{}{"code": WebhookCode, "key": WebhookUrlKey, "id": id}).Delete(&Setting{})
	return result.RowsAffected, result.Error
}

// endregion

// region 通用配置辅助方法

// getSettingValue 获取配置值的通用方法
func (setting *Setting) getSettingValue(code, key string) (string, error) {
	var s Setting
	err := Db.Where(map[string]interface{}{"code": code, "key": key}).First(&s).Error
	if err != nil {
		return "", err
	}
	return s.Value, nil
}

// updateOrCreateSetting 更新或创建配置的通用方法
func (setting *Setting) updateOrCreateSetting(code, key, value string) error {
	var s Setting
	err := Db.Where(map[string]interface{}{"code": code, "key": key}).First(&s).Error
	if err != nil {
		// 记录不存在，创建新记录
		s.Code = code
		s.Key = key
		s.Value = value
		result := Db.Create(&s)
		return result.Error
	}
	// 记录存在，更新
	result := Db.Model(&Setting{}).Where(map[string]interface{}{"code": code, "key": key}).Update("value", value)
	return result.Error
}

// endregion

// region 系统配置
func (setting *Setting) GetLogRetentionDays() int {
	value, err := setting.getSettingValue(SystemCode, LogRetentionDaysKey)
	if err != nil || value == "" {
		return 0
	}
	days, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return days
}

func (setting *Setting) UpdateLogRetentionDays(days int) error {
	return setting.updateOrCreateSetting(SystemCode, LogRetentionDaysKey, strconv.Itoa(days))
}

func (setting *Setting) GetLogCleanupTime() string {
	value, err := setting.getSettingValue(SystemCode, LogCleanupTimeKey)
	if err != nil || value == "" {
		return "03:00"
	}
	return value
}

func (setting *Setting) UpdateLogCleanupTime(cleanupTime string) error {
	return setting.updateOrCreateSetting(SystemCode, LogCleanupTimeKey, cleanupTime)
}

// endregion

// region LLM配置

// LLM OpenAI 兼容的大模型接入配置。
type LLM struct {
	Enable  bool   `json:"enable"`
	BaseURL string `json:"base_url"`
	ApiKey  string `json:"api_key"`
	Model   string `json:"model"`
}

// LLM 读取大模型配置。
func (setting *Setting) LLM() (LLM, error) {
	list := make([]Setting, 0)
	err := Db.Where("code = ?", LLMCode).Find(&list).Error
	llm := LLM{}
	if err != nil {
		return llm, err
	}
	for _, v := range list {
		switch v.Key {
		case LLMEnableKey:
			llm.Enable = v.Value == "1"
		case LLMBaseURLKey:
			llm.BaseURL = v.Value
		case LLMApiKeyKey:
			llm.ApiKey = v.Value
		case LLMModelKey:
			llm.Model = v.Value
		}
	}
	return llm, nil
}

// UpdateLLM 更新大模型配置。
func (setting *Setting) UpdateLLM(enable bool, baseURL, apiKey, model string) error {
	enableValue := "0"
	if enable {
		enableValue = "1"
	}
	if err := setting.updateOrCreateSetting(LLMCode, LLMEnableKey, enableValue); err != nil {
		return err
	}
	if err := setting.updateOrCreateSetting(LLMCode, LLMBaseURLKey, baseURL); err != nil {
		return err
	}
	if err := setting.updateOrCreateSetting(LLMCode, LLMApiKeyKey, apiKey); err != nil {
		return err
	}
	return setting.updateOrCreateSetting(LLMCode, LLMModelKey, model)
}

func (setting *Setting) GetLogFileSizeLimit() int {
	value, err := setting.getSettingValue(SystemCode, LogFileSizeLimitKey)
	if err != nil || value == "" {
		return 0
	}
	size, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return size
}

func (setting *Setting) UpdateLogFileSizeLimit(size int) error {
	return setting.updateOrCreateSetting(SystemCode, LogFileSizeLimitKey, strconv.Itoa(size))
}

// endregion
