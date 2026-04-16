package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type TaskScriptVersion struct {
	Id        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskId    int       `json:"task_id" gorm:"type:int;not null;index;uniqueIndex:idx_task_version"`
	Command   string    `json:"command" gorm:"type:text;not null"`
	Remark    string    `json:"remark" gorm:"type:varchar(200);not null;default:''"`
	Username  string    `json:"username" gorm:"type:varchar(64);not null;default:''"`
	Version   int       `json:"version" gorm:"type:int;not null;uniqueIndex:idx_task_version"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	BaseModel `json:"-" gorm:"-"`
}

func (v *TaskScriptVersion) Create() (int, error) {
	result := Db.Create(v)
	return v.Id, result.Error
}

func (v *TaskScriptVersion) List(taskId int, params CommonMap) ([]TaskScriptVersion, error) {
	v.parsePageAndPageSize(params)
	list := make([]TaskScriptVersion, 0)
	err := Db.Where("task_id = ?", taskId).
		Order("version DESC").
		Limit(v.PageSize).Offset(v.pageLimitOffset()).
		Find(&list).Error
	return list, err
}

func (v *TaskScriptVersion) Total(taskId int) (int64, error) {
	var count int64
	err := Db.Model(&TaskScriptVersion{}).Where("task_id = ?", taskId).Count(&count).Error
	return count, err
}

func (v *TaskScriptVersion) Detail(id int) (TaskScriptVersion, error) {
	var version TaskScriptVersion
	err := Db.Where("id = ?", id).First(&version).Error
	return version, err
}

func (v *TaskScriptVersion) GetLatestVersion(taskId int) (int, error) {
	var version TaskScriptVersion
	err := Db.Where("task_id = ?", taskId).Order("version DESC").First(&version).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return version.Version, nil
}

func (v *TaskScriptVersion) CleanOldVersions(taskId int, keep int) error {
	var count int64
	if err := Db.Model(&TaskScriptVersion{}).Where("task_id = ?", taskId).Count(&count).Error; err != nil {
		return err
	}
	if int(count) <= keep {
		return nil
	}

	var boundary TaskScriptVersion
	err := Db.Where("task_id = ?", taskId).
		Order("version DESC").
		Offset(keep).
		Limit(1).
		First(&boundary).Error
	if err != nil {
		return err
	}

	return Db.Where("task_id = ? AND version <= ?", taskId, boundary.Version).
		Delete(&TaskScriptVersion{}).Error
}
