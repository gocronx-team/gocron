package models

import (
	"time"

	"github.com/gocronx-team/gocron/internal/modules/logger"

	"gorm.io/gorm"
)

type TaskTemplate struct {
	Id          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"type:varchar(64);not null"`
	Description string    `json:"description" gorm:"type:varchar(500);not null;default:''"`
	Category    string    `json:"category" gorm:"type:varchar(32);not null;default:'custom';index"`
	Protocol    int8      `json:"protocol" gorm:"type:tinyint;not null;default:2"`
	Command     string    `json:"command" gorm:"type:text;not null"`
	HttpMethod  int8      `json:"http_method" gorm:"type:tinyint;not null;default:1"`
	HttpBody    string    `json:"http_body" gorm:"type:text"`
	HttpHeaders    string `json:"http_headers" gorm:"type:text"`
	SuccessPattern string `json:"success_pattern" gorm:"type:varchar(512);not null;default:''"`
	Tag            string `json:"tag" gorm:"type:varchar(255);not null;default:''"`
	Spec           string `json:"spec" gorm:"type:varchar(64);not null;default:''"`
	Timeout        int    `json:"timeout" gorm:"type:int;not null;default:0"`
	Multi          int8   `json:"multi" gorm:"type:tinyint;not null;default:1"`
	RetryTimes       int8   `json:"retry_times" gorm:"type:tinyint;not null;default:0"`
	RetryInterval    int16  `json:"retry_interval" gorm:"type:smallint;not null;default:0"`
	Timezone         string `json:"timezone" gorm:"type:varchar(64);not null;default:''"`
	NotifyStatus     int8   `json:"notify_status" gorm:"type:tinyint;not null;default:0"`
	NotifyType       int8   `json:"notify_type" gorm:"type:tinyint;not null;default:0"`
	NotifyKeyword    string `json:"notify_keyword" gorm:"type:varchar(128);not null;default:''"`
	LogRetentionDays int    `json:"log_retention_days" gorm:"type:smallint;not null;default:0"`
	IsBuiltin        int8   `json:"is_builtin" gorm:"type:tinyint;not null;default:0"`
	UsageCount  int       `json:"usage_count" gorm:"type:int;not null;default:0"`
	CreatedBy   string    `json:"created_by" gorm:"type:varchar(64);not null;default:''"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	BaseModel   `json:"-" gorm:"-"`
}

func (t *TaskTemplate) Create() (int, error) {
	result := Db.Create(t)
	return t.Id, result.Error
}

func (t *TaskTemplate) UpdateBean(id int) (int64, error) {
	result := Db.Model(&TaskTemplate{}).Where("id = ?", id).
		Select("name", "description", "category", "protocol", "command",
			"http_method", "http_body", "http_headers", "success_pattern",
			"tag", "spec", "timeout", "multi", "retry_times", "retry_interval",
			"timezone", "notify_status", "notify_type", "notify_keyword", "log_retention_days").
		UpdateColumns(map[string]interface{}{
			"name":               t.Name,
			"description":        t.Description,
			"category":           t.Category,
			"protocol":           t.Protocol,
			"command":            t.Command,
			"http_method":        t.HttpMethod,
			"http_body":          t.HttpBody,
			"http_headers":       t.HttpHeaders,
			"success_pattern":    t.SuccessPattern,
			"tag":                t.Tag,
			"spec":               t.Spec,
			"timeout":            t.Timeout,
			"multi":              t.Multi,
			"retry_times":        t.RetryTimes,
			"retry_interval":     t.RetryInterval,
			"timezone":           t.Timezone,
			"notify_status":      t.NotifyStatus,
			"notify_type":        t.NotifyType,
			"notify_keyword":     t.NotifyKeyword,
			"log_retention_days": t.LogRetentionDays,
		})
	return result.RowsAffected, result.Error
}

func (t *TaskTemplate) Delete(id int) (int64, error) {
	result := Db.Delete(&TaskTemplate{}, id)
	return result.RowsAffected, result.Error
}

func (t *TaskTemplate) Detail(id int) (TaskTemplate, error) {
	var tmpl TaskTemplate
	err := Db.Where("id = ?", id).First(&tmpl).Error
	return tmpl, err
}

func (t *TaskTemplate) List(params CommonMap) ([]TaskTemplate, error) {
	t.parsePageAndPageSize(params)
	list := make([]TaskTemplate, 0)

	query := Db.Model(&TaskTemplate{})
	t.parseWhere(query, params)

	err := query.Order("is_builtin DESC, updated_at DESC, id DESC").
		Limit(t.PageSize).Offset(t.pageLimitOffset()).
		Find(&list).Error
	return list, err
}

func (t *TaskTemplate) Total(params CommonMap) (int64, error) {
	var count int64
	query := Db.Model(&TaskTemplate{})
	t.parseWhere(query, params)
	err := query.Count(&count).Error
	return count, err
}

func (t *TaskTemplate) parseWhere(query *gorm.DB, params CommonMap) {
	category, ok := params["Category"]
	if ok && category.(string) != "" {
		query.Where("category = ?", category)
	}
	name, ok := params["Name"]
	if ok && name.(string) != "" {
		query.Where("name LIKE ?", "%"+name.(string)+"%")
	}
}

func (t *TaskTemplate) IncrementUsage(id int) error {
	return Db.Model(&TaskTemplate{}).Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

func (t *TaskTemplate) NameExist(name string, id int) (bool, error) {
	var count int64
	query := Db.Model(&TaskTemplate{}).Where("name = ?", name)
	if id > 0 {
		query = query.Where("id != ?", id)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (t *TaskTemplate) GetCategories() ([]string, error) {
	var categories []string
	err := Db.Model(&TaskTemplate{}).Distinct("category").Order("category").Pluck("category", &categories).Error
	return categories, err
}

// seedBuiltinTemplates 初始化内置模板
func seedBuiltinTemplates(tx *gorm.DB) {
	templates := []TaskTemplate{
		{
			Name:        "MySQL Database Backup",
			Description: "Backup MySQL database to compressed file",
			Category:    "backup",
			Protocol:    2,
			Command:     `mysqldump -h {{db_host}} -u {{db_user}} -p'{{db_pass}}' {{db_name}} | gzip > /backup/{{db_name}}_$(date +%Y%m%d_%H%M%S).sql.gz`,
			Tag:         "backup,database",
			Spec:        "0 0 2 * * *",
			Timeout:     3600,
			Multi:       0,
			IsBuiltin:   1,
		},
		{
			Name:        "PostgreSQL Database Backup",
			Description: "Backup PostgreSQL database to compressed file",
			Category:    "backup",
			Protocol:    2,
			Command:     `PGPASSWORD='{{db_pass}}' pg_dump -h {{db_host}} -U {{db_user}} {{db_name}} | gzip > /backup/{{db_name}}_$(date +%Y%m%d_%H%M%S).sql.gz`,
			Tag:         "backup,database",
			Spec:        "0 0 2 * * *",
			Timeout:     3600,
			Multi:       0,
			IsBuiltin:   1,
		},
		{
			Name:        "Clean Log Files",
			Description: "Delete log files older than specified days",
			Category:    "cleanup",
			Protocol:    2,
			Command:     `find {{log_dir}} -name "*.log" -mtime +{{retain_days}} -delete && echo "Cleanup completed"`,
			Tag:         "cleanup,logs",
			Spec:        "0 0 3 * * *",
			Timeout:     300,
			Multi:       0,
			IsBuiltin:   1,
		},
		{
			Name:        "Clean Temp Files",
			Description: "Delete temporary files in specified directory",
			Category:    "cleanup",
			Protocol:    2,
			Command:     `find {{temp_dir}} -type f -mtime +{{retain_days}} -delete && echo "Cleaned $(date)"`,
			Tag:         "cleanup",
			Spec:        "0 0 4 * * *",
			Timeout:     300,
			Multi:       0,
			IsBuiltin:   1,
		},
		{
			Name:        "HTTP Health Check",
			Description: "Check if HTTP endpoint is healthy",
			Category:    "monitor",
			Protocol:    2,
			Command:     `curl -sf -o /dev/null -w "%{http_code}" {{check_url}} || exit 1`,
			Tag:         "monitor,health",
			Spec:        "0 */5 * * * *",
			Timeout:     30,
			RetryTimes:  3,
			RetryInterval: 30,
			IsBuiltin:   1,
		},
		{
			Name:        "Disk Usage Alert",
			Description: "Alert when disk usage exceeds threshold",
			Category:    "monitor",
			Protocol:    2,
			Command:     `usage=$(df {{mount_point}} | awk 'NR==2{print $5}' | tr -d '%%') && [ "$usage" -lt {{threshold}} ] && echo "OK: ${usage}%% used" || (echo "WARN: ${usage}%% used, exceeds {{threshold}}%%" && exit 1)`,
			Tag:         "monitor,disk",
			Spec:        "0 */30 * * * *",
			Timeout:     30,
			IsBuiltin:   1,
		},
		{
			Name:        "Docker Container Restart",
			Description: "Restart a Docker container and verify status",
			Category:    "deploy",
			Protocol:    2,
			Command:     `docker restart {{container_name}} && sleep 3 && docker ps | grep {{container_name}}`,
			Tag:         "deploy,docker",
			Timeout:     120,
			Multi:       0,
			IsBuiltin:   1,
		},
		{
			Name:        "HTTP API Call (GET)",
			Description: "Call an HTTP GET API endpoint",
			Category:    "api",
			Protocol:    1,
			Command:     `{{api_url}}`,
			HttpMethod:  1,
			Tag:         "api,http",
			Timeout:     30,
			RetryTimes:  2,
			RetryInterval: 10,
			IsBuiltin:   1,
		},
		{
			Name:        "HTTP API Call (POST)",
			Description: "Call an HTTP POST API with JSON body",
			Category:    "api",
			Protocol:    1,
			Command:     `{{api_url}}`,
			HttpMethod:  2,
			HttpBody:    `{{json_body}}`,
			HttpHeaders: `{"Content-Type": "application/json"}`,
			Tag:         "api,http",
			Timeout:     30,
			RetryTimes:  2,
			RetryInterval: 10,
			IsBuiltin:   1,
		},
	}

	for i := range templates {
		var count int64
		tx.Model(&TaskTemplate{}).Where("name = ?", templates[i].Name).Count(&count)
		if count > 0 {
			continue
		}
		if err := tx.Create(&templates[i]).Error; err != nil {
			logger.Warnf("初始化内置模板 [%s] 失败: %v", templates[i].Name, err)
		}
	}
}
