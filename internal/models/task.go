package models

import (
	"encoding/json"
	"errors"
	"sort"
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

// NextRunTime 自定义时间类型，零值时序列化为空字符串
type NextRunTime time.Time

func (t NextRunTime) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return json.Marshal("")
	}
	return json.Marshal(tt.Format(DefaultTimeFormat))
}

func (t *NextRunTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		*t = NextRunTime(time.Time{})
		return nil
	}
	tt, err := time.Parse(DefaultTimeFormat, s)
	if err != nil {
		return err
	}
	*t = NextRunTime(tt)
	return nil
}

// 任务
type Task struct {
	Id               int                  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name             string               `json:"name" gorm:"type:varchar(32);not null"`
	Level            TaskLevel            `json:"level" gorm:"type:tinyint;not null;index;default:1"`
	DependencyTaskId string               `json:"dependency_task_id" gorm:"type:varchar(64);not null;default:''"`
	DependencyStatus TaskDependencyStatus `json:"dependency_status" gorm:"type:tinyint;not null;default:1"`
	Spec             string               `json:"spec" gorm:"type:varchar(64);not null"`
	Protocol         TaskProtocol         `json:"protocol" gorm:"type:tinyint;not null;index"`
	Command          string               `json:"command" gorm:"type:text;not null"`
	HttpMethod       TaskHTTPMethod       `json:"http_method" gorm:"type:tinyint;not null;default:1"`
	HttpBody         string               `json:"http_body" gorm:"type:text"`
	HttpHeaders      string               `json:"http_headers" gorm:"type:text"`
	SuccessPattern   string               `json:"success_pattern" gorm:"type:varchar(512);not null;default:''"`
	Timeout          int                  `json:"timeout" gorm:"type:mediumint;not null;default:0"`
	Multi            int8                 `json:"multi" gorm:"type:tinyint;not null;default:0"`
	RetryTimes       int8                 `json:"retry_times" gorm:"type:tinyint;not null;default:0"`
	RetryInterval    int16                `json:"retry_interval" gorm:"type:smallint;not null;default:0"`
	NotifyStatus     int8                 `json:"notify_status" gorm:"type:tinyint;not null;default:0"`
	NotifyType       int8                 `json:"notify_type" gorm:"type:tinyint;not null;default:0"`
	NotifyReceiverId string               `json:"notify_receiver_id" gorm:"type:varchar(256);not null;default:''"`
	NotifyKeyword    string               `json:"notify_keyword" gorm:"type:varchar(128);not null;default:''"`
	Tag              string               `json:"tag" gorm:"type:varchar(255);not null;default:''"`
	LogRetentionDays int                  `json:"log_retention_days" gorm:"type:smallint;not null;default:0"`
	Remark           string               `json:"remark" gorm:"type:varchar(100);not null;default:''"`
	Status           Status               `json:"status" gorm:"type:tinyint;not null;index;default:0"`
	CreatedAt        time.Time            `json:"created" gorm:"column:created;autoCreateTime"`
	DeletedAt        *time.Time           `json:"deleted" gorm:"column:deleted;index"`
	BaseModel        `json:"-" gorm:"-"`
	Hosts            []TaskHostDetail `json:"hosts" gorm:"-"`
	NextRunTime      NextRunTime      `json:"next_run_time" gorm:"-"`
}

// 新增
func (task *Task) Create() (insertId int, err error) {
	// 使用 Select 显式列出所有列，确保零值字段（如 Multi=0）也会被写入，
	// 覆盖 gorm 标签中的 default 值，同时 GORM 会将自增主键回填到 task.Id。
	result := Db.Select(
		"name", "level", "dependency_task_id", "dependency_status",
		"spec", "protocol", "command", "http_method", "http_body",
		"http_headers", "success_pattern", "timeout", "multi",
		"retry_times", "retry_interval", "notify_status", "notify_type",
		"notify_receiver_id", "notify_keyword", "tag", "log_retention_days",
		"remark", "status",
	).Create(task)
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
			"dependency_status", "tag", "http_method", "http_body",
			"http_headers", "success_pattern", "notify_keyword",
			"log_retention_days").
		UpdateColumns(map[string]interface{}{
			"name":               task.Name,
			"spec":               task.Spec,
			"protocol":           task.Protocol,
			"command":            task.Command,
			"timeout":            task.Timeout,
			"multi":              task.Multi,
			"retry_times":        task.RetryTimes,
			"retry_interval":     task.RetryInterval,
			"remark":             task.Remark,
			"notify_status":      task.NotifyStatus,
			"notify_type":        task.NotifyType,
			"notify_receiver_id": task.NotifyReceiverId,
			"dependency_task_id": task.DependencyTaskId,
			"dependency_status":  task.DependencyStatus,
			"tag":                task.Tag,
			"http_method":        task.HttpMethod,
			"http_body":          task.HttpBody,
			"http_headers":       task.HttpHeaders,
			"success_pattern":    task.SuccessPattern,
			"notify_keyword":     task.NotifyKeyword,
			"log_retention_days": task.LogRetentionDays,
		})
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
func (task *Task) ActiveListByHostId(hostId int) ([]Task, error) {
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

// 优化：批量查询任务主机信息，避免N+1查询问题
func (task *Task) setHostsForTasks(tasks []Task) ([]Task, error) {
	if len(tasks) == 0 {
		return tasks, nil
	}

	// 收集所有任务ID
	taskIds := make([]int, len(tasks))
	for i, t := range tasks {
		taskIds[i] = t.Id
	}

	// 批量查询所有任务的主机信息
	taskHostModel := new(TaskHost)
	hostsMap, err := taskHostModel.GetHostsByTaskIds(taskIds)
	if err != nil {
		return nil, err
	}

	// 分配主机信息到对应任务
	for i := range tasks {
		if hosts, ok := hostsMap[tasks[i].Id]; ok {
			tasks[i].Hosts = hosts
		} else {
			tasks[i].Hosts = []TaskHostDetail{}
		}
	}

	return tasks, nil
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

// GetAllTags 获取所有任务中使用的标签，去重并排序返回
func (task *Task) GetAllTags() ([]string, error) {
	var tags []string
	err := Db.Model(&Task{}).Where("tag != ''").Distinct("tag").Pluck("tag", &tags).Error
	if err != nil {
		return nil, err
	}

	tagSet := make(map[string]struct{})
	for _, tagStr := range tags {
		parts := strings.Split(tagStr, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				tagSet[trimmed] = struct{}{}
			}
		}
	}

	result := make([]string, 0, len(tagSet))
	for t := range tagSet {
		result = append(result, t)
	}
	sort.Strings(result)

	return result, nil
}
