package task

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers/base"
	"github.com/gocronx-team/gocron/internal/service"
	"github.com/jakecoffman/cron"
	"github.com/ouqiang/goutil"
)

type TaskForm struct {
	Id               int                         `form:"id" json:"id"`
	Level            models.TaskLevel            `form:"level" json:"level" binding:"required,oneof=1 2"`
	DependencyStatus models.TaskDependencyStatus `form:"dependency_status" json:"dependency_status"`
	DependencyTaskId string                      `form:"dependency_task_id" json:"dependency_task_id"`
	Name             string                      `form:"name" json:"name" binding:"required,max=32"`
	Spec             string                      `form:"spec" json:"spec"`
	Protocol         models.TaskProtocol         `form:"protocol" json:"protocol" binding:"oneof=1 2"`
	Command          string                      `form:"command" json:"command" binding:"required,max=256"`
	HttpMethod       models.TaskHTTPMethod       `form:"http_method" json:"http_method" binding:"oneof=1 2"`
	Timeout          int                         `form:"timeout" json:"timeout" binding:"min=0,max=86400"`
	Multi            int8                        `form:"multi" json:"multi" binding:"oneof=1 2"`
	RetryTimes       int8                        `form:"retry_times" json:"retry_times"`
	RetryInterval    int16                       `form:"retry_interval" json:"retry_interval"`
	HostId           string                      `form:"host_id" json:"host_id"`
	Tag              string                      `form:"tag" json:"tag"`
	Remark           string                      `form:"remark" json:"remark"`
	NotifyStatus     int8                        `form:"notify_status" json:"notify_status" binding:"oneof=1 2 3 4"`
	NotifyType       int8                        `form:"notify_type" json:"notify_type" binding:"oneof=1 2 3 4"`
	NotifyReceiverId string                      `form:"notify_receiver_id" json:"notify_receiver_id"`
	NotifyKeyword    string                      `form:"notify_keyword" json:"notify_keyword"`
}

// 首页
func Index(c *gin.Context) {
	taskModel := new(models.Task)
	queryParams := parseQueryParams(c)
	total, err := taskModel.Total(queryParams)
	if err != nil {
		logger.Error(err)
	}
	tasks, err := taskModel.List(queryParams)
	if err != nil {
		logger.Error(err)
	}
	for i, item := range tasks {
		tasks[i].NextRunTime = service.ServiceTask.NextRunTime(item)
	}
	jsonResp := utils.JsonResponse{}
	result := jsonResp.Success(utils.SuccessContent, map[string]interface{}{
		"total": total,
		"data":  tasks,
	})
	c.String(http.StatusOK, result)
}

// Detail 任务详情
func Detail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	taskModel := new(models.Task)
	task, err := taskModel.Detail(id)
	jsonResp := utils.JsonResponse{}
	var result string
	if err != nil || task.Id == 0 {
		logger.Errorf("编辑任务#获取任务详情失败#任务ID-%d", id)
		result = jsonResp.Success(utils.SuccessContent, nil)
	} else {
		result = jsonResp.Success(utils.SuccessContent, task)
	}
	c.String(http.StatusOK, result)
}

// 保存任务
func Store(c *gin.Context) {
	var form TaskForm
	if err := c.ShouldBind(&form); err != nil {
		json := utils.JsonResponse{}
		result := json.CommonFailure("表单验证失败, 请检测输入")
		c.String(http.StatusOK, result)
		return
	}

	json := utils.JsonResponse{}
	taskModel := models.Task{}
	var id = form.Id
	nameExists, err := taskModel.NameExist(form.Name, form.Id)
	if err != nil {
		result := json.CommonFailure(utils.FailureContent, err)
		c.String(http.StatusOK, result)
		return
	}
	if nameExists {
		result := json.CommonFailure("任务名称已存在")
		c.String(http.StatusOK, result)
		return
	}

	if form.Protocol == models.TaskRPC && form.HostId == "" {
		result := json.CommonFailure("请选择主机名")
		c.String(http.StatusOK, result)
		return
	}

	taskModel.Name = form.Name
	taskModel.Protocol = form.Protocol
	taskModel.Command = strings.TrimSpace(form.Command)
	taskModel.Timeout = form.Timeout
	taskModel.Tag = form.Tag
	taskModel.Remark = form.Remark
	taskModel.Multi = form.Multi
	taskModel.RetryTimes = form.RetryTimes
	taskModel.RetryInterval = form.RetryInterval
	if taskModel.Multi != 1 {
		taskModel.Multi = 0
	}
	taskModel.NotifyStatus = form.NotifyStatus - 1
	taskModel.NotifyType = form.NotifyType - 1
	taskModel.NotifyReceiverId = form.NotifyReceiverId
	taskModel.NotifyKeyword = form.NotifyKeyword
	taskModel.Spec = form.Spec
	taskModel.Level = form.Level
	taskModel.DependencyStatus = form.DependencyStatus
	taskModel.DependencyTaskId = strings.TrimSpace(form.DependencyTaskId)
	if taskModel.NotifyStatus > 0 && taskModel.NotifyType != 3 && taskModel.NotifyReceiverId == "" {
		result := json.CommonFailure("至少选择一个通知接收者")
		c.String(http.StatusOK, result)
		return
	}
	taskModel.HttpMethod = form.HttpMethod
	if taskModel.Protocol == models.TaskHTTP {
		command := strings.ToLower(taskModel.Command)
		if !strings.HasPrefix(command, "http://") && !strings.HasPrefix(command, "https://") {
			result := json.CommonFailure("请输入正确的URL地址")
			c.String(http.StatusOK, result)
			return
		}
		if taskModel.Timeout > 300 {
			result := json.CommonFailure("HTTP任务超时时间不能超过300秒")
			c.String(http.StatusOK, result)
			return
		}
	}

	if taskModel.RetryTimes > 10 || taskModel.RetryTimes < 0 {
		result := json.CommonFailure("任务重试次数取值0-10")
		c.String(http.StatusOK, result)
		return
	}

	if taskModel.RetryInterval > 3600 || taskModel.RetryInterval < 0 {
		result := json.CommonFailure("任务重试间隔时间取值0-3600")
		c.String(http.StatusOK, result)
		return
	}

	if taskModel.DependencyStatus != models.TaskDependencyStatusStrong &&
		taskModel.DependencyStatus != models.TaskDependencyStatusWeak {
		result := json.CommonFailure("请选择依赖关系")
		c.String(http.StatusOK, result)
		return
	}

	if taskModel.Level == models.TaskLevelParent {
		err = goutil.PanicToError(func() {
			cron.Parse(form.Spec)
		})
		if err != nil {
			result := json.CommonFailure("crontab表达式解析失败", err)
			c.String(http.StatusOK, result)
			return
		}
	} else {
		taskModel.DependencyTaskId = ""
		taskModel.Spec = ""
	}

	if id > 0 && taskModel.DependencyTaskId != "" {
		dependencyTaskIds := strings.Split(taskModel.DependencyTaskId, ",")
		if utils.InStringSlice(dependencyTaskIds, strconv.Itoa(id)) {
			result := json.CommonFailure("不允许设置当前任务为子任务")
			c.String(http.StatusOK, result)
			return
		}
	}

	if id == 0 {
		taskModel.Status = models.Running
		id, err = taskModel.Create()
	} else {
		_, err = taskModel.UpdateBean(id)
	}

	if err != nil {
		result := json.CommonFailure("保存失败", err)
		c.String(http.StatusOK, result)
		return
	}

	taskHostModel := new(models.TaskHost)
	if form.Protocol == models.TaskRPC {
		hostIdStrList := strings.Split(form.HostId, ",")
		hostIds := make([]int, len(hostIdStrList))
		for i, hostIdStr := range hostIdStrList {
			hostIds[i], _ = strconv.Atoi(hostIdStr)
		}
		taskHostModel.Add(id, hostIds)
	} else {
		taskHostModel.Remove(id)
	}

	status, _ := taskModel.GetStatus(id)
	if status == models.Enabled && taskModel.Level == models.TaskLevelParent {
		addTaskToTimer(id)
	}

	result := json.Success("保存成功", nil)
	c.String(http.StatusOK, result)
}

// 删除任务
func Remove(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	json := utils.JsonResponse{}
	taskModel := new(models.Task)
	_, err := taskModel.Delete(id)
	var result string
	if err != nil {
		result = json.CommonFailure(utils.FailureContent, err)
	} else {
		taskHostModel := new(models.TaskHost)
		taskHostModel.Remove(id)
		service.ServiceTask.Remove(id)
		result = json.Success(utils.SuccessContent, nil)
	}
	c.String(http.StatusOK, result)
}

// 激活任务
func Enable(c *gin.Context) {
	changeStatus(c, models.Enabled)
}

// 暂停任务
func Disable(c *gin.Context) {
	changeStatus(c, models.Disabled)
}

// 手动运行任务
func Run(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	json := utils.JsonResponse{}
	taskModel := new(models.Task)
	task, err := taskModel.Detail(id)
	var result string
	if err != nil || task.Id <= 0 {
		result = json.CommonFailure("获取任务详情失败", err)
	} else {
		task.Spec = "手动运行"
		service.ServiceTask.Run(task)
		result = json.Success("任务已开始运行, 请到任务日志中查看结果", nil)
	}
	c.String(http.StatusOK, result)
}

// 改变任务状态
func changeStatus(c *gin.Context, status models.Status) {
	id, _ := strconv.Atoi(c.Param("id"))
	json := utils.JsonResponse{}
	taskModel := new(models.Task)
	_, err := taskModel.Update(id, models.CommonMap{
		"status": status,
	})
	var result string
	if err != nil {
		result = json.CommonFailure(utils.FailureContent, err)
	} else {
		if status == models.Enabled {
			addTaskToTimer(id)
		} else {
			service.ServiceTask.Remove(id)
		}
		result = json.Success(utils.SuccessContent, nil)
	}
	c.String(http.StatusOK, result)
}

// 添加任务到定时器
func addTaskToTimer(id int) {
	taskModel := new(models.Task)
	task, err := taskModel.Detail(id)
	if err != nil {
		logger.Error(err)
		return
	}

	service.ServiceTask.RemoveAndAdd(task)
}

// 解析查询参数
func parseQueryParams(c *gin.Context) models.CommonMap {
	var params models.CommonMap = models.CommonMap{}
	id, _ := strconv.Atoi(c.Query("id"))
	hostId, _ := strconv.Atoi(c.Query("host_id"))
	protocol, _ := strconv.Atoi(c.Query("protocol"))
	status, _ := strconv.Atoi(c.Query("status"))
	params["Id"] = id
	params["HostId"] = hostId
	params["Name"] = strings.TrimSpace(c.Query("name"))
	params["Protocol"] = protocol
	params["Tag"] = strings.TrimSpace(c.Query("tag"))
	if status >= 0 {
		status -= 1
	}
	params["Status"] = status
	base.ParsePageAndPageSize(c, params)

	return params
}
