package models

import (
	"time"
)

// AgentToken agent注册token
type AgentToken struct {
	Id        int        `json:"id" gorm:"primaryKey;autoIncrement"`
	Token     string     `json:"token" gorm:"type:varchar(64);uniqueIndex;not null"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	Used      bool       `json:"used" gorm:"default:false"`
	UsedAt    *time.Time `json:"used_at" gorm:"default:null"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

func (t *AgentToken) Create() error {
	return Db.Create(t).Error
}

func (t *AgentToken) FindByToken(token string) error {
	return Db.Where("token = ?", token).First(t).Error
}

func (t *AgentToken) MarkAsUsed() error {
	if !t.Used {
		t.Used = true
		now := time.Now()
		t.UsedAt = &now
		return Db.Save(t).Error
	}
	return nil
}

func (t *AgentToken) IsValid() bool {
	return time.Now().Before(t.ExpiresAt)
}

// CleanExpired 清理过期token
func (t *AgentToken) CleanExpired() error {
	return Db.Where("expires_at < ?", time.Now()).Delete(&AgentToken{}).Error
}
