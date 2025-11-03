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

const slackTemplate = `Task ID: {{.TaskId}}
Task Name: {{.TaskName}}
Status: {{.Status}}
Result: {{.Result}}
Remark: {{.Remark}}`

const emailTemplate = `Task ID: {{.TaskId}}
Task Name: {{.TaskName}}
Status: {{.Status}}
Result: {{.Result}}
Remark: {{.Remark}}`
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
	SystemCode           = "system"
	LogRetentionDaysKey  = "log_retention_days"
	LogCleanupTimeKey    = "log_cleanup_time"
	LogFileSizeLimitKey  = "log_file_size_limit"
	AgentRegisterTokenKey = "agent_register_token"
)

// 初始化基本字段 邮件、slack等
func (setting *Setting) InitBasicField() {
	// 定义所有需要初始化的配置
	configs := []struct {
		Code  string
		Key   string
		Value string
	}{
		{SlackCode, SlackUrlKey, ""},
		{SlackCode, SlackTemplateKey, slackTemplate},
		{MailCode, MailServerKey, ""},
		{MailCode, MailTemplateKey, emailTemplate},
		{WebhookCode, WebhookUrlKey, ""},
		{WebhookCode, WebhookTemplateKey, webhookTemplate},
		{SystemCode, LogRetentionDaysKey, "0"},
		{SystemCode, LogCleanupTimeKey, "03:00"},
		{SystemCode, LogFileSizeLimitKey, "0"},
		{SystemCode, AgentRegisterTokenKey, ""},
	}

	// 检查并创建不存在的配置
	for _, cfg := range configs {
		var count int64
		Db.Model(&Setting{}).Where("code = ? AND key = ?", cfg.Code, cfg.Key).Count(&count)
		if count == 0 {
			s := &Setting{
				Code:  cfg.Code,
				Key:   cfg.Key,
				Value: cfg.Value,
			}
			Db.Create(s)
		}
	}
}

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
	Db.Model(&Setting{}).Where("code = ? AND key = ?", SlackCode, SlackUrlKey).Update("value", url)

	setting.Value = template
	Db.Model(&Setting{}).Where("code = ? AND key = ?", SlackCode, SlackTemplateKey).Update("value", template)

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
	Db.Model(&Setting{}).Where("code = ? AND key = ? AND value = ?", SlackCode, SlackChannelKey, channel).Count(&count)
	return count > 0
}

// 删除slack渠道
func (setting *Setting) RemoveChannel(id int) (int64, error) {
	result := Db.Where("code = ? AND key = ? AND id = ?", SlackCode, SlackChannelKey, id).Delete(&Setting{})
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
			json.Unmarshal([]byte(v.Value), mail)
		case MailUserKey:
			json.Unmarshal([]byte(v.Value), &mailUser)
			mailUser.Id = v.Id
			mail.MailUsers = append(mail.MailUsers, mailUser)
		case MailTemplateKey:
			mail.Template = v.Value
		}

	}
}

func (setting *Setting) UpdateMail(config, template string) error {
	Db.Model(&Setting{}).Where("code = ? AND key = ?", MailCode, MailServerKey).Update("value", config)
	Db.Model(&Setting{}).Where("code = ? AND key = ?", MailCode, MailTemplateKey).Update("value", template)

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
	result := Db.Where("code = ? AND key = ? AND id = ?", MailCode, MailUserKey, id).Delete(&Setting{})
	return result.RowsAffected, result.Error
}

type WebHook struct {
	Url      string `json:"url"`
	Template string `json:"template"`
}

func (setting *Setting) Webhook() (WebHook, error) {
	list := make([]Setting, 0)
	err := Db.Where("code = ?", WebhookCode).Find(&list).Error
	webHook := WebHook{}
	if err != nil {
		return webHook, err
	}

	setting.formatWebhook(list, &webHook)

	return webHook, err
}

func (setting *Setting) formatWebhook(list []Setting, webHook *WebHook) {
	for _, v := range list {
		switch v.Key {
		case WebhookUrlKey:
			webHook.Url = v.Value
		case WebhookTemplateKey:
			webHook.Template = v.Value
		}

	}
}

func (setting *Setting) UpdateWebHook(url, template string) error {
	Db.Model(&Setting{}).Where("code = ? AND key = ?", WebhookCode, WebhookUrlKey).Update("value", url)
	Db.Model(&Setting{}).Where("code = ? AND key = ?", WebhookCode, WebhookTemplateKey).Update("value", template)

	return nil
}

// endregion

// region 系统配置
func (setting *Setting) GetLogRetentionDays() int {
	var s Setting
	err := Db.Where("code = ? AND key = ?", SystemCode, LogRetentionDaysKey).First(&s).Error
	if err != nil {
		return 0
	}
	var days int
	if s.Value == "" {
		return 0
	}
	if n, err := strconv.Atoi(s.Value); err == nil {
		days = n
	}
	return days
}

func (setting *Setting) UpdateLogRetentionDays(days int) error {
	var s Setting
	err := Db.Where("code = ? AND key = ?", SystemCode, LogRetentionDaysKey).First(&s).Error
	if err != nil {
		// 记录不存在，创建新记录
		s.Code = SystemCode
		s.Key = LogRetentionDaysKey
		s.Value = strconv.Itoa(days)
		result := Db.Create(&s)
		return result.Error
	}
	// 记录存在，更新
	result := Db.Model(&Setting{}).Where("code = ? AND key = ?", SystemCode, LogRetentionDaysKey).Update("value", strconv.Itoa(days))
	return result.Error
}

func (setting *Setting) GetLogCleanupTime() string {
	var s Setting
	err := Db.Where("code = ? AND key = ?", SystemCode, LogCleanupTimeKey).First(&s).Error
	if err != nil {
		return "03:00"
	}
	if s.Value == "" {
		return "03:00"
	}
	return s.Value
}

func (setting *Setting) UpdateLogCleanupTime(cleanupTime string) error {
	var s Setting
	err := Db.Where("code = ? AND key = ?", SystemCode, LogCleanupTimeKey).First(&s).Error
	if err != nil {
		s.Code = SystemCode
		s.Key = LogCleanupTimeKey
		s.Value = cleanupTime
		result := Db.Create(&s)
		return result.Error
	}
	result := Db.Model(&Setting{}).Where("code = ? AND key = ?", SystemCode, LogCleanupTimeKey).Update("value", cleanupTime)
	return result.Error
}

func (setting *Setting) GetLogFileSizeLimit() int {
	var s Setting
	err := Db.Where("code = ? AND key = ?", SystemCode, LogFileSizeLimitKey).First(&s).Error
	if err != nil {
		return 0
	}
	var size int
	if s.Value == "" {
		return 0
	}
	if n, err := strconv.Atoi(s.Value); err == nil {
		size = n
	}
	return size
}

func (setting *Setting) UpdateLogFileSizeLimit(size int) error {
	var s Setting
	err := Db.Where("code = ? AND key = ?", SystemCode, LogFileSizeLimitKey).First(&s).Error
	if err != nil {
		s.Code = SystemCode
		s.Key = LogFileSizeLimitKey
		s.Value = strconv.Itoa(size)
		result := Db.Create(&s)
		return result.Error
	}
	result := Db.Model(&Setting{}).Where("code = ? AND key = ?", SystemCode, LogFileSizeLimitKey).Update("value", strconv.Itoa(size))
	return result.Error
}

func (setting *Setting) GetAgentRegisterToken() string {
	var s Setting
	err := Db.Where("code = ? AND key = ?", SystemCode, AgentRegisterTokenKey).First(&s).Error
	if err != nil {
		return ""
	}
	return s.Value
}

func (setting *Setting) UpdateAgentRegisterToken(token string) error {
	var s Setting
	err := Db.Where("code = ? AND key = ?", SystemCode, AgentRegisterTokenKey).First(&s).Error
	if err != nil {
		s.Code = SystemCode
		s.Key = AgentRegisterTokenKey
		s.Value = token
		result := Db.Create(&s)
		return result.Error
	}
	result := Db.Model(&Setting{}).Where("code = ? AND key = ?", SystemCode, AgentRegisterTokenKey).Update("value", token)
	return result.Error
}
// endregion
