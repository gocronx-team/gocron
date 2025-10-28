package models

import (
	"time"

	"gorm.io/gorm"
)

type TaskType int8

// 任务执行日志
type TaskLog struct {
	Id         int64        `json:"id" gorm:"primaryKey;autoIncrement;type:bigint"`
	TaskId     int          `json:"task_id" gorm:"not null;index;default:0"`
	Name       string       `json:"name" gorm:"type:varchar(32);not null"`
	Spec       string       `json:"spec" gorm:"type:varchar(64);not null"`
	Protocol   TaskProtocol `json:"protocol" gorm:"type:tinyint;not null;index"`
	Command    string       `json:"command" gorm:"type:varchar(256);not null"`
	Timeout    int          `json:"timeout" gorm:"type:mediumint;not null;default:0"`
	RetryTimes int8         `json:"retry_times" gorm:"type:tinyint;not null;default:0"`
	Hostname   string       `json:"hostname" gorm:"type:varchar(128);not null;default:''"`
	StartTime  time.Time    `json:"start_time" gorm:"column:start_time;autoCreateTime"`
	EndTime    time.Time    `json:"end_time" gorm:"column:end_time;autoUpdateTime"`
	Status     Status       `json:"status" gorm:"type:tinyint;not null;index;default:1"`
	Result     string       `json:"result" gorm:"type:mediumtext;not null"`
	TotalTime  int          `json:"total_time" gorm:"-"`
	BaseModel  `json:"-" gorm:"-"`
}

func (taskLog *TaskLog) Create() (insertId int64, err error) {
	result := Db.Create(taskLog)
	if result.Error == nil {
		insertId = taskLog.Id
	}

	return insertId, result.Error
}

// 更新
func (taskLog *TaskLog) Update(id int64, data CommonMap) (int64, error) {
	updateData := make(map[string]interface{})
	for k, v := range data {
		updateData[k] = v
	}
	result := Db.Model(&TaskLog{}).Where("id = ?", id).UpdateColumns(updateData)
	return result.RowsAffected, result.Error
}

func (taskLog *TaskLog) List(params CommonMap) ([]TaskLog, error) {
	taskLog.parsePageAndPageSize(params)
	list := make([]TaskLog, 0)
	query := Db.Order("id DESC")
	taskLog.parseWhere(query, params)
	err := query.Limit(taskLog.PageSize).Offset(taskLog.pageLimitOffset()).Find(&list).Error
	
	if len(list) > 0 {
		for i, item := range list {
			endTime := item.EndTime
			if item.Status == Running {
				endTime = time.Now()
			}
			execSeconds := endTime.Sub(item.StartTime).Seconds()
			list[i].TotalTime = int(execSeconds)
		}
	}

	return list, err
}

// 清空表
func (taskLog *TaskLog) Clear() (int64, error) {
	result := Db.Where("1=1").Delete(&TaskLog{})
	return result.RowsAffected, result.Error
}

// 删除N个月前的日志
func (taskLog *TaskLog) Remove(id int) (int64, error) {
	t := time.Now().AddDate(0, -id, 0)
	result := Db.Where("start_time <= ?", t.Format(DefaultTimeFormat)).Delete(&TaskLog{})
	return result.RowsAffected, result.Error
}

func (taskLog *TaskLog) Total(params CommonMap) (int64, error) {
	var count int64
	query := Db.Model(&TaskLog{})
	taskLog.parseWhere(query, params)
	err := query.Count(&count).Error
	return count, err
}

// 解析where
func (taskLog *TaskLog) parseWhere(query *gorm.DB, params CommonMap) {
	if len(params) == 0 {
		return
	}
	taskId, ok := params["TaskId"]
	if ok && taskId.(int) > 0 {
		query.Where("task_id = ?", taskId)
	}
	protocol, ok := params["Protocol"]
	if ok && protocol.(int) > 0 {
		query.Where("protocol = ?", protocol)
	}
	status, ok := params["Status"]
	if ok && status.(int) > -1 {
		query.Where("status = ?", status)
	}
}
