package models

import "time"

import "gorm.io/gorm/clause"

// TaskLogChunk stores bounded, already-redacted output while a task is running.
// The terminal output remains in TaskLog.Result for API and upgrade compatibility.
type TaskLogChunk struct {
	Id        int64     `json:"id" gorm:"primaryKey;autoIncrement;type:bigint"`
	TaskLogId int64     `json:"task_log_id" gorm:"not null;uniqueIndex:idx_task_log_chunk_log_seq,priority:1"`
	Seq       int64     `json:"seq" gorm:"not null;uniqueIndex:idx_task_log_chunk_log_seq,priority:2"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (TaskLogChunk) TableName() string {
	return TablePrefix + "task_log_chunk"
}

func (chunk *TaskLogChunk) Append(items []TaskLogChunk) error {
	if len(items) == 0 {
		return nil
	}
	return Db.Clauses(clause.OnConflict{DoNothing: true}).Create(&items).Error
}

func (chunk *TaskLogChunk) ListAfter(taskLogId, seq int64, limit int) ([]TaskLogChunk, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	items := make([]TaskLogChunk, 0)
	err := Db.Where("task_log_id = ? AND seq > ?", taskLogId, seq).
		Order("seq ASC").Limit(limit).Find(&items).Error
	return items, err
}

func (chunk *TaskLogChunk) DeleteByTaskLogId(taskLogId int64) error {
	return Db.Where("task_log_id = ?", taskLogId).Delete(&TaskLogChunk{}).Error
}
