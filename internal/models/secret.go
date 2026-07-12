package models

import (
	"regexp"
	"strings"
	"time"
)

// reservedEnvNames 是不允许用作机密名的系统环境变量:覆盖它们会破坏任务执行
// 环境(如 PATH)或带来安全风险(如 LD_PRELOAD 库劫持)。
var reservedEnvNames = map[string]bool{
	"PATH":                  true,
	"HOME":                  true,
	"IFS":                   true,
	"SHELL":                 true,
	"PWD":                   true,
	"LD_PRELOAD":            true,
	"LD_LIBRARY_PATH":       true,
	"DYLD_INSERT_LIBRARIES": true,
	"DYLD_LIBRARY_PATH":     true,
	"GOCRON_SECRET_KEY":     true,
}

// IsReservedEnvName 判断机密名是否为受保护的系统环境变量(大小写不敏感)。
func IsReservedEnvName(name string) bool {
	return reservedEnvNames[strings.ToUpper(strings.TrimSpace(name))]
}

// envNamePattern 限定机密名为合法的环境变量名(字母/下划线开头,后接字母数字下划线)。
var envNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// IsValidEnvName 判断名称是否为合法的环境变量名,用于机密名及任务机密白名单项校验。
func IsValidEnvName(name string) bool {
	return envNamePattern.MatchString(name)
}

// Secret 机密(类 GitHub Secrets):值以 AES-GCM 密文存储,写入后不可读取,
// 仅在任务调度执行时按需解密注入到执行环境。Value 字段对外永不序列化。
type Secret struct {
	Id        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"type:varchar(64);uniqueIndex;not null"` // 环境变量名
	Value     string    `json:"-" gorm:"type:text;not null"`                       // 密文,永不返回前端
	Remark    string    `json:"remark" gorm:"type:varchar(255);not null;default:''"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Create 写入一条机密,返回新 id。
func (s *Secret) Create() (int, error) {
	if err := Db.Create(s).Error; err != nil {
		return 0, err
	}
	return s.Id, nil
}

// Update 按 id 更新字段。
func (s *Secret) Update(id int, data CommonMap) (int64, error) {
	updateData := make(map[string]interface{}, len(data))
	for k, v := range data {
		updateData[k] = v
	}
	result := Db.Model(&Secret{}).Where("id = ?", id).Updates(updateData)
	return result.RowsAffected, result.Error
}

// Delete 按 id 删除。
func (s *Secret) Delete(id int) (int64, error) {
	result := Db.Where("id = ?", id).Delete(&Secret{})
	return result.RowsAffected, result.Error
}

// List 返回机密列表,显式排除 Value 密文字段,避免任何序列化路径泄露。
func (s *Secret) List() ([]Secret, error) {
	list := make([]Secret, 0)
	err := Db.Select("id", "name", "remark", "created_at", "updated_at").
		Order("name ASC").Find(&list).Error
	return list, err
}

// All 返回包含 Value 密文的全部机密,仅供调度执行时解密注入使用。
func (s *Secret) All() ([]Secret, error) {
	list := make([]Secret, 0)
	err := Db.Order("name ASC").Find(&list).Error
	return list, err
}

// Find 按 id 查找(含 Value)。
func (s *Secret) Find(id int) error {
	return Db.Where("id = ?", id).First(s).Error
}

// NameExists 判断机密名是否已存在,excludeId 用于更新时排除自身。
func (s *Secret) NameExists(name string, excludeId int) (int64, error) {
	var count int64
	query := Db.Model(&Secret{}).Where("name = ?", name)
	if excludeId > 0 {
		query = query.Where("id != ?", excludeId)
	}
	err := query.Count(&count).Error
	return count, err
}
