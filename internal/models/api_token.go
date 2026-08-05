package models

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// ApiToken MCP/API 远程访问令牌。
// 与 AgentToken 的区别：本令牌长期有效、绑定用户、可被手动吊销，用于远程 MCP 客户端的
// Bearer 认证；数据库仅保存 token 的 sha256 哈希，明文只在创建时返回一次。
type ApiToken struct {
	Id         int        `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId     int        `json:"user_id" gorm:"index;not null"`
	Name       string     `json:"name" gorm:"type:varchar(64);not null;default:''"`
	TokenHash  string     `json:"-" gorm:"type:varchar(64);uniqueIndex;not null"`
	LastUsedAt *time.Time `json:"last_used_at" gorm:"default:null"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// HashToken 计算 token 明文的存储哈希。
func HashToken(plain string) string {
	h := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(h[:])
}

func (t *ApiToken) Create() error {
	return Db.Create(t).Error
}

// ListByUser 返回指定用户的全部 token，按创建时间倒序。
func (t *ApiToken) ListByUser(userId int) ([]ApiToken, error) {
	list := make([]ApiToken, 0)
	err := Db.Where("user_id = ?", userId).Order("id DESC").Find(&list).Error
	return list, err
}

// FindByHash 按哈希查找 token。
func (t *ApiToken) FindByHash(hash string) error {
	return Db.Where("token_hash = ?", hash).First(t).Error
}

// Delete 删除（吊销）属于该用户的 token，限定 user_id 防止越权删除他人令牌。
func (t *ApiToken) Delete(id, userId int) (int64, error) {
	result := Db.Where("id = ? AND user_id = ?", id, userId).Delete(&ApiToken{})
	return result.RowsAffected, result.Error
}

// TouchLastUsed 更新最近使用时间，best-effort，失败不影响认证流程。
func (t *ApiToken) TouchLastUsed() {
	now := time.Now()
	Db.Model(&ApiToken{}).Where("id = ?", t.Id).UpdateColumn("last_used_at", &now)
}
