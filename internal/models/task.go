package models

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type TaskProtocol int8

const (
	TaskHTTP TaskProtocol = iota + 1 // HTTP协议
	TaskRPC                          // RPC方式执行命令
)

type TaskLevel int8

const (
	TaskLevelParent TaskLevel = 1 // 父任务
	TaskLevelChild  TaskLevel = 2 // 子任务(依赖任务)
)

type TaskDependencyStatus int8

const (
	TaskDependencyStatusStrong TaskDependencyStatus = 1 // 强依赖
	TaskDependencyStatusWeak   TaskDependencyStatus = 2 // 弱依赖
)

type TaskHTTPMethod int8

const (
	TaskHTTPMethodGet  TaskHTTPMethod = 1
	TaskHttpMethodPost TaskHTTPMethod = 2
)

// 任务
type Task struct {
	Id               int                  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name             string               `json:"name" gorm:"type:varchar(32);not null"`
	Level            TaskLevel            `json:"level" gorm:"type:tinyint;not null;index;default:1"`
	DependencyTaskId string               `json:"dependency_task_id" gorm:"type:varchar(64);not null;default:''"`
	DependencyStatus TaskDependencyStatus `json:"dependency_status" gorm:"type:tinyint;not null;default:1"`
	Spec             string               `json:"spec" gorm:"type:varchar(64);not null"`
	Protocol         TaskProtocol         `json:"protocol" gorm:"type:tinyint;not null;index"`
	Command          string               `json:"command" gorm:"type:varchar(256);not null"`
	HttpMethod       TaskHTTPMethod       `json:"http_method" gorm:"type:tinyint;not null;default:1"`
	Timeout          int                  `json:"timeout" gorm:"type:mediumint;not null;default:0"`
	Multi            int8                 `json:"multi" gorm:"type:tinyint;not null;default:1"`
	RetryTimes       int8                 `json:"retry_times" gorm:"type:tinyint;not null;default:0"`
	RetryInterval    int16                `json:"retry_interval" gorm:"type:smallint;not null;default:0"`
	NotifyStatus     int8                 `json:"notify_status" gorm:"type:tinyint;not null;default:1"`
	NotifyType       int8                 `json:"notify_type" gorm:"type:tinyint;not null;default:0"`
	NotifyReceiverId string               `json:"notify_receiver_id" gorm:"type:varchar(256);not null;default:''"`
	NotifyKeyword    string               `json:"notify_keyword" gorm:"type:varchar(128);not null;default:''"`
	Tag              string               `json:"tag" gorm:"type:varchar(32);not null;default:''"`
	Remark           string               `json:"remark" gorm:"type:varchar(100);not null;default:''"`
	Status           Status               `json:"status" gorm:"type:tinyint;not null;index;default:0"`
	CreatedAt        time.Time            `json:"created" gorm:"column:created;autoCreateTime"`
	DeletedAt        *time.Time           `json:"deleted" gorm:"column:deleted;index"`
	BaseModel        `json:"-" gorm:"-"`
	Hosts            []TaskHostDetail `json:"hosts" gorm:"-"`
	NextRunTime      time.Time        `json:"next_run_time" gorm:"-"`
}

// 新增
func (task *Task) Create() (insertId int, err error) {
	result := Db.Create(task)
	if result.Error == nil {
		insertId = task.Id
	}

	return insertId, result.Error
}

func (task *Task) UpdateBean(id int) (int64, error) {
	result := Db.Model(&Task{}).Where("id = ?", id).
		Select("name", "spec", "protocol", "command", "timeout", "multi",
			"retry_times", "retry_interval", "remark", "notify_status",
			"notify_type", "notify_receiver_id", "dependency_task_id",
			"dependency_status", "tag", "http_method", "notify_keyword").
		Updates(task)
	return result.RowsAffected, result.Error
}

// 更新
func (task *Task) Update(id int, data CommonMap) (int64, error) {
	updateData := make(map[string]interface{})
	for k, v := range data {
		updateData[k] = v
	}
	result := Db.Model(&Task{}).Where("id = ?", id).UpdateColumns(updateData)
	return result.RowsAffected, result.Error
}

// 删除
func (task *Task) Delete(id int) (int64, error) {
	result := Db.Delete(&Task{}, id)
	return result.RowsAffected, result.Error
}

// 禁用
func (task *Task) Disable(id int) (int64, error) {
	return task.Update(id, CommonMap{"status": Disabled})
}

// 激活
func (task *Task) Enable(id int) (int64, error) {
	return task.Update(id, CommonMap{"status": Enabled})
}

// 获取所有激活任务
func (task *Task) ActiveList(page, pageSize int) ([]Task, error) {
	params := CommonMap{"Page": page, "PageSize": pageSize}
	task.parsePageAndPageSize(params)
	list := make([]Task, 0)
	err := Db.Where("status = ? AND level = ?", Enabled, TaskLevelParent).
		Limit(task.PageSize).Offset(task.pageLimitOffset()).
		Find(&list).Error

	if err != nil {
		return list, err
	}

	return task.setHostsForTasks(list)
}

// 获取某个主机下的所有激活任务
func (task *Task) ActiveListByHostId(hostId int16) ([]Task, error) {
	taskHostModel := new(TaskHost)
	taskIds, err := taskHostModel.GetTaskIdsByHostId(hostId)
	if err != nil {
		return nil, err
	}
	if len(taskIds) == 0 {
		return nil, nil
	}
	list := make([]Task, 0)
	err = Db.Where("status = ? AND level = ?", Enabled, TaskLevelParent).
		Where("id IN ?", taskIds).
		Find(&list).Error
	if err != nil {
		return list, err
	}

	return task.setHostsForTasks(list)
}

func (task *Task) setHostsForTasks(tasks []Task) ([]Task, error) {
	taskHostModel := new(TaskHost)
	var err error
	for i, value := range tasks {
		taskHostDetails, err := taskHostModel.GetHostIdsByTaskId(value.Id)
		if err != nil {
			return nil, err
		}
		tasks[i].Hosts = taskHostDetails
	}

	return tasks, err
}

// 判断任务名称是否存在
func (task *Task) NameExist(name string, id int) (bool, error) {
	var count int64
	query := Db.Model(&Task{}).Where("name = ? AND status = ?", name, Enabled)
	if id > 0 {
		query = query.Where("id != ?", id)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (task *Task) GetStatus(id int) (Status, error) {
	err := Db.First(task, id).Error
	if err != nil {
		return 0, err
	}

	return task.Status, nil
}

func (task *Task) Detail(id int) (Task, error) {
	t := Task{}
	err := Db.Where("id = ?", id).First(&t).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return t, nil
		}
		return t, err
	}

	taskHostModel := new(TaskHost)
	t.Hosts, err = taskHostModel.GetHostIdsByTaskId(id)

	return t, err
}

func (task *Task) List(params CommonMap) ([]Task, error) {
	task.parsePageAndPageSize(params)
	list := make([]Task, 0)

	query := Db.Table(TablePrefix + "task as t").
		Joins("LEFT JOIN " + TablePrefix + "task_host as th ON t.id = th.task_id")

	task.parseWhere(query, params)

	err := query.Group("t.id").
		Order("t.id DESC").
		Select("t.*").
		Limit(task.PageSize).Offset(task.pageLimitOffset()).
		Find(&list).Error

	if err != nil {
		return nil, err
	}

	return task.setHostsForTasks(list)
}

// 获取依赖任务列表
func (task *Task) GetDependencyTaskList(ids string) ([]Task, error) {
	list := make([]Task, 0)
	if ids == "" {
		return list, nil
	}
	idList := strings.Split(ids, ",")

	err := Db.Where("level = ?", TaskLevelChild).
		Where("id IN ?", idList).
		Find(&list).Error

	if err != nil {
		return list, err
	}

	return task.setHostsForTasks(list)
}

func (task *Task) Total(params CommonMap) (int64, error) {
	type Result struct {
		Count int64
	}
	var result Result

	query := Db.Table(TablePrefix + "task as t").
		Joins("LEFT JOIN " + TablePrefix + "task_host as th ON t.id = th.task_id")

	task.parseWhere(query, params)

	err := query.Group("t.id").Count(&result.Count).Error

	return result.Count, err
}

// 解析where
func (task *Task) parseWhere(query *gorm.DB, params CommonMap) {
	if len(params) == 0 {
		return
	}
	id, ok := params["Id"]
	if ok && id.(int) > 0 {
		query.Where("t.id = ?", id)
	}
	hostId, ok := params["HostId"]
	if ok && hostId.(int) > 0 {
		query.Where("th.host_id = ?", hostId)
	}
	name, ok := params["Name"]
	if ok && name.(string) != "" {
		query.Where("t.name LIKE ?", "%"+name.(string)+"%")
	}
	protocol, ok := params["Protocol"]
	if ok && protocol.(int) > 0 {
		query.Where("protocol = ?", protocol)
	}
	status, ok := params["Status"]
	if ok && status.(int) > -1 {
		query.Where("status = ?", status)
	}

	tag, ok := params["Tag"]
	if ok && tag.(string) != "" {
		query.Where("t.tag LIKE ?", "%"+tag.(string)+"%")
	}
}
