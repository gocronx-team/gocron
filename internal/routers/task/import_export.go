package task

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-yaml"
	"github.com/gocronx-team/cron"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers/base"
)

const yamlExportVersion = 1

// yamlTask 是任务的可移植表示:主机以 name 引用(便于跨实例迁移),
// 不含执行历史、脚本版本与依赖关系(dependency_task_id 是本实例的 id,跨实例无意义)。
type yamlTask struct {
	Name               string   `yaml:"name"`
	Level              int8     `yaml:"level"`
	Spec               string   `yaml:"spec,omitempty"`
	Protocol           int8     `yaml:"protocol"`
	Command            string   `yaml:"command"`
	HttpMethod         int8     `yaml:"http_method,omitempty"`
	HttpBody           string   `yaml:"http_body,omitempty"`
	HttpHeaders        string   `yaml:"http_headers,omitempty"`
	SuccessPattern     string   `yaml:"success_pattern,omitempty"`
	Timeout            int      `yaml:"timeout,omitempty"`
	Multi              int8     `yaml:"multi,omitempty"`
	RetryTimes         int8     `yaml:"retry_times,omitempty"`
	RetryInterval      int16    `yaml:"retry_interval,omitempty"`
	Tag                string   `yaml:"tag,omitempty"`
	Remark             string   `yaml:"remark,omitempty"`
	NotifyStatus       int8     `yaml:"notify_status,omitempty"`
	NotifyType         int8     `yaml:"notify_type,omitempty"`
	NotifyKeyword      string   `yaml:"notify_keyword,omitempty"`
	NotifyKeywordRegex int8     `yaml:"notify_keyword_regex,omitempty"`
	LogRetentionDays   int      `yaml:"log_retention_days,omitempty"`
	Hosts              []string `yaml:"hosts,omitempty"` // RPC 任务的执行节点(host name)
}

type yamlDoc struct {
	Version int        `yaml:"version"`
	Tasks   []yamlTask `yaml:"tasks"`
}

// ImportResult 导入结果汇总。
type ImportResult struct {
	Created  int      `json:"created"`
	Skipped  int      `json:"skipped"`
	Messages []string `json:"messages"`
}

// Export 导出全部任务为 YAML(附件下载)。
func Export(c *gin.Context) {
	taskModel := new(models.Task)
	list, err := taskModel.List(models.CommonMap{"Page": 1, "PageSize": 100000})
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}

	doc := yamlDoc{Version: yamlExportVersion, Tasks: make([]yamlTask, 0, len(list))}
	for _, t := range list {
		yt := yamlTask{
			Name:               t.Name,
			Level:              int8(t.Level),
			Spec:               t.Spec,
			Protocol:           int8(t.Protocol),
			Command:            t.Command,
			HttpMethod:         int8(t.HttpMethod),
			HttpBody:           t.HttpBody,
			HttpHeaders:        t.HttpHeaders,
			SuccessPattern:     t.SuccessPattern,
			Timeout:            t.Timeout,
			Multi:              t.Multi,
			RetryTimes:         t.RetryTimes,
			RetryInterval:      t.RetryInterval,
			Tag:                t.Tag,
			Remark:             t.Remark,
			NotifyStatus:       t.NotifyStatus,
			NotifyType:         t.NotifyType,
			NotifyKeyword:      t.NotifyKeyword,
			NotifyKeywordRegex: t.NotifyKeywordRegex,
			LogRetentionDays:   t.LogRetentionDays,
		}
		for _, h := range t.Hosts {
			yt.Hosts = append(yt.Hosts, h.Name)
		}
		doc.Tasks = append(doc.Tasks, yt)
	}

	data, err := yaml.Marshal(doc)
	if err != nil {
		base.RespondErrorWithDefaultMsg(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=gocron-tasks.yaml")
	c.Data(http.StatusOK, "application/x-yaml; charset=utf-8", data)
}

// Import 从 YAML 导入任务。同名任务跳过;引用的主机以 name 映射到本实例,
// 不存在的主机关联会被跳过并在结果中提示。不导入依赖关系。
func Import(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}
	var doc yamlDoc
	if err := yaml.Unmarshal(body, &doc); err != nil {
		base.RespondError(c, i18n.T(c, "task_import_parse_failed"))
		return
	}

	// host name -> id(本实例)。Host.AllList 出于前端需要只返回 name/port(不含 id),
	// 这里直接查询 id+name 建映射。
	var hosts []models.Host
	_ = models.Db.Model(&models.Host{}).Select("id", "name").Find(&hosts).Error
	hostNameToID := make(map[string]int, len(hosts))
	for _, h := range hosts {
		hostNameToID[h.Name] = h.Id
	}

	taskModel := new(models.Task)
	res := ImportResult{Messages: make([]string, 0)}

	for _, yt := range doc.Tasks {
		if yt.Name == "" || yt.Command == "" {
			res.Skipped++
			res.Messages = append(res.Messages, "skipped: empty name or command")
			continue
		}
		if yt.Protocol != int8(models.TaskHTTP) && yt.Protocol != int8(models.TaskRPC) {
			res.Skipped++
			res.Messages = append(res.Messages, fmt.Sprintf("%s: invalid protocol", yt.Name))
			continue
		}
		if models.TaskLevel(yt.Level) != models.TaskLevelParent &&
			models.TaskLevel(yt.Level) != models.TaskLevelChild {
			yt.Level = int8(models.TaskLevelParent)
		}
		if models.TaskHTTPMethod(yt.HttpMethod) != models.TaskHTTPMethodGet &&
			models.TaskHTTPMethod(yt.HttpMethod) != models.TaskHttpMethodPost {
			yt.HttpMethod = int8(models.TaskHTTPMethodGet)
		}

		exists, _ := taskModel.NameExist(yt.Name, 0)
		if exists {
			res.Skipped++
			res.Messages = append(res.Messages, fmt.Sprintf("%s: already exists, skipped", yt.Name))
			continue
		}

		// 父任务校验 cron;子任务不独立调度,清空 spec
		spec := yt.Spec
		if models.TaskLevel(yt.Level) == models.TaskLevelParent {
			if perr := utils.PanicToError(func() { cron.Parse(spec) }); perr != nil {
				res.Skipped++
				res.Messages = append(res.Messages, fmt.Sprintf("%s: invalid cron spec", yt.Name))
				continue
			}
		} else {
			spec = ""
		}

		t := models.Task{
			Name:               yt.Name,
			Level:              models.TaskLevel(yt.Level),
			Spec:               spec,
			Protocol:           models.TaskProtocol(yt.Protocol),
			Command:            yt.Command,
			HttpMethod:         models.TaskHTTPMethod(yt.HttpMethod),
			HttpBody:           yt.HttpBody,
			HttpHeaders:        yt.HttpHeaders,
			SuccessPattern:     yt.SuccessPattern,
			Timeout:            yt.Timeout,
			Multi:              yt.Multi,
			RetryTimes:         yt.RetryTimes,
			RetryInterval:      yt.RetryInterval,
			Tag:                yt.Tag,
			Remark:             yt.Remark,
			NotifyStatus:       yt.NotifyStatus,
			NotifyType:         yt.NotifyType,
			NotifyKeyword:      yt.NotifyKeyword,
			NotifyKeywordRegex: yt.NotifyKeywordRegex,
			LogRetentionDays:   yt.LogRetentionDays,
			DependencyStatus:   models.TaskDependencyStatusWeak, // 不导入依赖关系
			Status:             models.Enabled,
		}

		id, cerr := t.Create()
		if cerr != nil {
			res.Skipped++
			res.Messages = append(res.Messages, fmt.Sprintf("%s: create failed: %v", yt.Name, cerr))
			continue
		}
		res.Created++

		// RPC 任务:按 name 关联本实例主机
		if t.Protocol == models.TaskRPC {
			hostIds := make([]int, 0, len(yt.Hosts))
			for _, hn := range yt.Hosts {
				if hid, ok := hostNameToID[hn]; ok {
					hostIds = append(hostIds, hid)
				} else {
					res.Messages = append(res.Messages, fmt.Sprintf("%s: host %q not found, association skipped", yt.Name, hn))
				}
			}
			if len(hostIds) > 0 {
				if aerr := new(models.TaskHost).Add(id, hostIds); aerr != nil {
					logger.Errorf("导入任务主机关联失败#任务ID-%d#%s", id, aerr)
				}
			}
		}

		// 启用的父任务加入调度
		if t.Level == models.TaskLevelParent {
			addTaskToTimer(id)
		}
	}

	c.Set("audit_target_name", fmt.Sprintf("imported %d tasks", res.Created))
	base.RespondSuccess(c, i18n.T(c, "task_import_success"), res)
}
