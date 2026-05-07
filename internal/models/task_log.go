package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type LocalTime time.Time

func (t LocalTime) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", time.Time(t).Format(DefaultTimeFormat))
	return []byte(formatted), nil
}

func (t *LocalTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	parsed, err := time.ParseInLocation(`"`+DefaultTimeFormat+`"`, string(data), time.Local)
	if err == nil {
		*t = LocalTime(parsed)
	}
	return err
}

func (t LocalTime) Value() (driver.Value, error) {
	return time.Time(t), nil
}

func (t *LocalTime) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if v, ok := value.(time.Time); ok {
		*t = LocalTime(v)
		return nil
	}
	return fmt.Errorf("cannot scan %T into LocalTime", value)
}

type TaskType int8

// 任务执行日志
type TaskLog struct {
	Id         int64        `json:"id" gorm:"primaryKey;autoIncrement;type:bigint"`
	TaskId     int          `json:"task_id" gorm:"not null;index;default:0"`
	Name       string       `json:"name" gorm:"type:varchar(32);not null"`
	Spec       string       `json:"spec" gorm:"type:varchar(64);not null"`
	Protocol   TaskProtocol `json:"protocol" gorm:"smallint;not null;index"`
	Command    string       `json:"command" gorm:"type:varchar(256);not null"`
	Timeout    int          `json:"timeout" gorm:"type:mediumint;not null;default:0"`
	RetryTimes int8         `json:"retry_times" gorm:"smallint;not null;default:0"`
	Hostname   string       `json:"hostname" gorm:"type:varchar(128);not null;default:''"`
	StartTime  LocalTime    `json:"start_time" gorm:"column:start_time;autoCreateTime"`
	EndTime    LocalTime    `json:"end_time" gorm:"column:end_time;autoUpdateTime"`
	Status     Status       `json:"status" gorm:"smallint;not null;index;default:1"`
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
			endTime := time.Time(item.EndTime)
			if item.Status == Running {
				endTime = time.Now()
			}
			execSeconds := endTime.Sub(time.Time(item.StartTime)).Seconds()
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

// 清空指定任务的日志(批量删除)
func (taskLog *TaskLog) ClearByTaskId(taskId int) (int64, error) {
	if taskId <= 0 {
		return 0, nil
	}
	var totalAffected int64
	batchSize := 1000
	for {
		result := Db.Where("task_id = ?", taskId).Limit(batchSize).Delete(&TaskLog{})
		if result.Error != nil {
			return totalAffected, result.Error
		}
		totalAffected += result.RowsAffected
		if result.RowsAffected < int64(batchSize) {
			break
		}
	}
	return totalAffected, nil
}

// 删除N个月前的日志
func (taskLog *TaskLog) Remove(id int) (int64, error) {
	t := time.Now().AddDate(0, -id, 0)
	result := Db.Where("start_time <= ?", t.Format(DefaultTimeFormat)).Delete(&TaskLog{})
	return result.RowsAffected, result.Error
}

// 删除N天前的日志
func (taskLog *TaskLog) RemoveByDays(days int) (int64, error) {
	if days <= 0 {
		return 0, nil
	}
	t := time.Now().AddDate(0, 0, -days)
	result := Db.Where("start_time < ?", t).Delete(&TaskLog{})
	return result.RowsAffected, result.Error
}

// 删除N天前的日志，排除有自定义保留策略的任务
func (taskLog *TaskLog) RemoveByDaysExcludingCustomRetention(days int) (int64, error) {
	if days <= 0 {
		return 0, nil
	}
	t := time.Now().AddDate(0, 0, -days)
	result := Db.Where("start_time < ? AND task_id NOT IN (SELECT id FROM "+TablePrefix+"task WHERE log_retention_days > 0)", t).Delete(&TaskLog{})
	return result.RowsAffected, result.Error
}

// 删除指定任务N天前的日志（批量删除，每批1000条）
func (taskLog *TaskLog) RemoveByTaskIdAndDays(taskId int, days int) (int64, error) {
	if taskId <= 0 || days <= 0 {
		return 0, nil
	}
	t := time.Now().AddDate(0, 0, -days)
	var totalDeleted int64
	for {
		result := Db.Where("task_id = ? AND start_time < ?", taskId, t).
			Limit(1000).
			Delete(&TaskLog{})
		if result.Error != nil {
			return totalDeleted, result.Error
		}
		totalDeleted += result.RowsAffected
		if result.RowsAffected < 1000 {
			break
		}
	}
	return totalDeleted, nil
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

// 统计相关方法

// DailyStats 每日统计数据
type DailyStats struct {
	Date    string `json:"date"`
	Total   int    `json:"total"`
	Success int    `json:"success"`
	Failed  int    `json:"failed"`
}

// GetLast7DaysTrend 获取最近7天的执行趋势
func (taskLog *TaskLog) GetLast7DaysTrend() ([]DailyStats, error) {
	var stats []DailyStats

	// 使用 Go 计算7天前的日期，兼容所有数据库
	sevenDaysAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	err := Db.Raw(`
		SELECT
			DATE(start_time) as date,
			COUNT(*) as total,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as success,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as failed
		FROM `+TablePrefix+`task_log
		WHERE start_time >= ? AND start_time < ?
		GROUP BY DATE(start_time)
		ORDER BY date DESC
	`, Finish, Failure, sevenDaysAgo, tomorrow).Scan(&stats).Error

	return stats, err
}

// GetTodayStats 获取今日统计数据
func (taskLog *TaskLog) GetTodayStats() (total, success, failed int64, err error) {
	// 使用 Go 计算今天的日期范围
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	// 今日总执行次数
	err = Db.Model(&TaskLog{}).
		Where("start_time >= ? AND start_time < ?", today, tomorrow).
		Count(&total).Error
	if err != nil {
		return
	}

	// 今日成功次数
	err = Db.Model(&TaskLog{}).
		Where("start_time >= ? AND start_time < ? AND status = ?", today, tomorrow, Finish).
		Count(&success).Error
	if err != nil {
		return
	}

	// 今日失败次数
	err = Db.Model(&TaskLog{}).
		Where("start_time >= ? AND start_time < ? AND status = ?", today, tomorrow, Failure).
		Count(&failed).Error

	return
}
