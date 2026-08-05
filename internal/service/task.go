package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gocronx-team/cron"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/app"
	"github.com/gocronx-team/gocron/internal/modules/crypto"
	"github.com/gocronx-team/gocron/internal/modules/diagnosis"
	"github.com/gocronx-team/gocron/internal/modules/httpclient"
	"github.com/gocronx-team/gocron/internal/modules/llm"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/notify"
	rpcClient "github.com/gocronx-team/gocron/internal/modules/rpc/client"
	pb "github.com/gocronx-team/gocron/internal/modules/rpc/proto"
	"github.com/gocronx-team/gocron/internal/modules/utils"
)

var (
	ServiceTask Task
)

var (
	httpGetFunc              = httpclient.Get
	httpPostParamsFunc       = httpclient.PostParams
	httpPostJsonFunc         = httpclient.PostJson
	httpGetWithHeadersFunc   = httpclient.GetWithHeaders
	httpPostJsonWithHdrsFunc = httpclient.PostJsonWithHeaders
	httpPostParamsWithHdrs   = httpclient.PostParamsWithHeaders
	notifyPushFunc           = notify.Push
	sleepFunc                = time.Sleep

	// 定时任务调度管理器
	serviceCron *cron.Cron

	// 同一任务是否有实例处于运行中
	runInstance Instance

	// 调度器运行状态
	schedulerMu      sync.Mutex
	schedulerRunning bool

	// 任务计数-正在运行的任务
	taskCount TaskCount

	// 并发队列, 限制同时运行的任务数量
	concurrencyQueue ConcurrencyQueue
)

// 并发队列
type ConcurrencyQueue struct {
	queue chan struct{}
}

func (cq *ConcurrencyQueue) Add() {
	cq.queue <- struct{}{}
}

func (cq *ConcurrencyQueue) Done() {
	<-cq.queue
}

// 任务计数
type TaskCount struct {
	wg   sync.WaitGroup
	exit chan struct{}
}

func (tc *TaskCount) Add() {
	tc.wg.Add(1)
}

func (tc *TaskCount) Done() {
	tc.wg.Done()
}

func (tc *TaskCount) Exit() {
	tc.wg.Done()
	<-tc.exit
}

func (tc *TaskCount) Wait() {
	tc.Add()
	tc.wg.Wait()
	close(tc.exit)
}

// 任务ID作为Key
type Instance struct {
	m sync.Map
}

// 是否有任务处于运行中
func (i *Instance) has(key int) bool {
	_, ok := i.m.Load(key)

	return ok
}

func (i *Instance) add(key int) {
	i.m.Store(key, struct{}{})
}

func (i *Instance) done(key int) {
	i.m.Delete(key)
}

// tryAdd 原子地尝试添加任务实例
// 返回 true 表示成功添加（任务未在运行），false 表示任务已在运行
func (i *Instance) tryAdd(key int) bool {
	_, loaded := i.m.LoadOrStore(key, struct{}{})
	return !loaded
}

type Task struct{}

type TaskResult struct {
	Result     string
	Err        error
	RetryTimes int8
}

// Initialize 初始化调度器基础设施（不加载任务）
// 任务加载由 StartScheduler 完成，配合 leader election 使用
func (task Task) Initialize() {
	concurrencyQueue = ConcurrencyQueue{queue: make(chan struct{}, app.Setting.ConcurrencyQueue)}
	taskCount = TaskCount{sync.WaitGroup{}, make(chan struct{})}
	go taskCount.Wait()
	logger.Info("Scheduler infrastructure initialized")
}

// StartScheduler 启动调度器并加载所有任务（当选 leader 时调用）
func (task Task) StartScheduler() {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()

	if schedulerRunning {
		return
	}

	serviceCron = cron.New()
	serviceCron.Start()

	logger.Info("Starting to load scheduled tasks (this node is leader)")
	taskModel := new(models.Task)
	taskNum := 0
	page := 1
	pageSize := 1000
	maxPage := 1000
	for page < maxPage {
		taskList, err := taskModel.ActiveList(page, pageSize)
		if err != nil {
			logger.Fatalf("Scheduled task initialization#Failed to get task list: %s", err)
		}
		if len(taskList) == 0 {
			break
		}
		for _, item := range taskList {
			logger.Infof("Adding task to scheduler#ID-%d#Name-%s#Protocol-%d#Host count-%d", item.Id, item.Name, item.Protocol, len(item.Hosts))
			task.Add(item)
			taskNum++
		}
		page++
	}
	logger.Infof("Scheduled task initialization completed, %d tasks added to scheduler", taskNum)

	task.initLogCleanupTask()
	schedulerRunning = true
}

// StopScheduler 停止调度器（失去 leader 时调用）
func (task Task) StopScheduler() {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()

	if !schedulerRunning {
		return
	}

	logger.Info("Stopping scheduler (this node lost leadership)")
	serviceCron.Stop()
	serviceCron = nil
	schedulerRunning = false
}

// IsSchedulerRunning 返回调度器是否正在运行
func (task Task) IsSchedulerRunning() bool {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()
	return schedulerRunning
}

// 初始化日志清理任务
func (task Task) initLogCleanupTask() {
	if serviceCron == nil {
		return
	}
	settingModel := new(models.Setting)
	cleanupTime := settingModel.GetLogCleanupTime()
	// 解析时间 HH:MM
	var hour, minute int
	if n, err := fmt.Sscanf(cleanupTime, "%d:%d", &hour, &minute); err != nil || n != 2 ||
		hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		logger.Warnf("日志清理时间解析失败，使用默认值 00:00 (cleanupTime=%q)", cleanupTime)
		hour, minute = 0, 0
	}
	// 生成cron表达式: 秒 分 时 日 月 周
	cronSpec := fmt.Sprintf("0 %d %d * * *", minute, hour)

	serviceCron.AddFunc(cronSpec, func() {
		// 1. Task-level log retention: clean logs for tasks with custom retention days
		taskLogModel := new(models.TaskLog)
		page := 1
		pageSize := 1000
		for {
			var tasks []models.Task
			err := models.Db.Where("log_retention_days > 0").
				Limit(pageSize).Offset((page - 1) * pageSize).
				Find(&tasks).Error
			if err != nil {
				logger.Errorf("Failed to query tasks with custom log retention: %s", err)
				break
			}
			if len(tasks) == 0 {
				break
			}
			for _, t := range tasks {
				count, err := taskLogModel.RemoveByTaskIdAndDays(t.Id, t.LogRetentionDays)
				if err != nil {
					logger.Errorf("Failed to cleanup logs for task %d: %s", t.Id, err)
				} else if count > 0 {
					logger.Infof("Task %d: cleaned up %d logs older than %d days", t.Id, count, t.LogRetentionDays)
				}
			}
			page++
		}

		// 2. Global log retention: clean remaining logs
		settingModel := new(models.Setting)
		days := settingModel.GetLogRetentionDays()
		if days > 0 {
			count, err := taskLogModel.RemoveByDaysExcludingCustomRetention(days)
			if err != nil {
				logger.Errorf("Failed to auto-cleanup database logs: %s", err)
			} else {
				logger.Infof("Auto-cleanup database logs older than %d days, deleted %d records", days, count)
			}
			// 清理日志文件
			cleanupLogFiles()
		}
	}, "log-cleanup")
	logger.Infof("Log auto-cleanup task added, execution time: %s", cleanupTime)
}

// 重新加载日志清理任务
func (task Task) ReloadLogCleanupTask() {
	if serviceCron == nil {
		return // follower node or scheduler not started, skip
	}
	// 先移除旧任务
	serviceCron.RemoveJob("log-cleanup")
	// 重新添加任务
	task.initLogCleanupTask()
	logger.Info("Log cleanup task reloaded")
}

// 批量添加任务
func (task Task) BatchAdd(tasks []models.Task) {
	for _, item := range tasks {
		task.RemoveAndAdd(item)
	}
}

// 删除任务后添加
func (task Task) RemoveAndAdd(taskModel models.Task) {
	task.Remove(taskModel.Id)
	task.Add(taskModel)
}

// 添加任务
func (task Task) Add(taskModel models.Task) {
	if taskModel.Level == models.TaskLevelChild {
		logger.Errorf("Failed to add task#Child tasks cannot be added to scheduler#Task ID-%d", taskModel.Id)
		return
	}
	if serviceCron == nil {
		return // follower node, skip
	}
	taskFunc := createJob(taskModel)
	if taskFunc == nil {
		logger.Error("Failed to create task job#Unsupported task protocol#", taskModel.Protocol)
		return
	}

	cronName := strconv.Itoa(taskModel.Id)
	err := utils.PanicToError(func() {
		serviceCron.AddFunc(taskModel.Spec, taskFunc, cronName)
	})
	if err != nil {
		logger.Error("Failed to add task to scheduler#", err)
	}
}

func (task Task) NextRunTime(taskModel models.Task) time.Time {
	if serviceCron == nil {
		return time.Time{}
	}
	if taskModel.Level != models.TaskLevelParent ||
		taskModel.Status != models.Enabled {
		return time.Time{}
	}
	entries := serviceCron.Entries()
	taskName := strconv.Itoa(taskModel.Id)
	for _, item := range entries {
		if item.Name == taskName {
			return item.Next
		}
	}

	return time.Time{}
}

// 停止运行中的任务
func (task Task) Stop(ip string, port int, id int64) error {
	return rpcClient.Stop(ip, port, id)
}

// ReconcileStoppedLog finalizes a stale Running row after every target node
// has authoritatively reported that the execution ID is no longer present.
func (task Task) ReconcileStoppedLog(id int64) error {
	var output strings.Builder
	var seq int64
	for {
		chunks, err := new(models.TaskLogChunk).ListAfter(id, seq, 100)
		if err != nil {
			return err
		}
		for _, chunk := range chunks {
			output.WriteString(chunk.Content)
			seq = chunk.Seq
		}
		if len(chunks) < 100 {
			break
		}
	}
	result := limitTaskLog(output.String(), false)
	rows, err := new(models.TaskLog).FinalizeIfRunning(id, models.CommonMap{
		"status": models.Cancel, "result": result, "end_time": time.Now(),
	})
	if err != nil || rows == 0 {
		return err
	}
	return new(models.TaskLogChunk).DeleteByTaskLogId(id)
}

func (task Task) Remove(id int) {
	if serviceCron == nil {
		return
	}
	serviceCron.RemoveJob(strconv.Itoa(id))
}

// 等待所有任务结束后退出
func (task Task) WaitAndExit() {
	schedulerMu.Lock()
	if schedulerRunning && serviceCron != nil {
		serviceCron.Stop()
		schedulerRunning = false
	}
	schedulerMu.Unlock()
	taskCount.Exit()
}

// 直接运行任务
func (task Task) Run(taskModel models.Task) {
	go createJob(taskModel)()
}

type Handler interface {
	// secretEnv 为本次执行需注入的机密环境变量(name->明文);由调用方一次性加载并复用,
	// 保证注入与后续脱敏用的是同一份快照。HTTP 任务忽略该参数。
	Run(taskModel models.Task, taskUniqueId int64, secretEnv map[string]string) (string, error)
}

type progressHandler interface {
	RunWithProgress(taskModel models.Task, taskUniqueId int64, secretEnv map[string]string, onChunk func(string)) (string, error)
}

const (
	taskLogFlushInterval = time.Second
	maxTaskLogBytes      = 1 << 20
	taskLogTruncatedMark = "\n...[task output truncated at 1 MiB]"
)

type liveLogAccumulator struct {
	mu        sync.Mutex
	taskLogID int64
	masker    *streamingSecretMasker
	batch     strings.Builder
	seq       int64
	total     int
	truncated bool
	dirty     bool
	stop      chan struct{}
	done      chan struct{}
}

func newLiveLogAccumulator(taskLogID int64, secretEnv map[string]string) *liveLogAccumulator {
	a := &liveLogAccumulator{
		taskLogID: taskLogID, masker: newStreamingSecretMasker(secretValues(secretEnv)),
		stop: make(chan struct{}), done: make(chan struct{}),
	}
	go a.run()
	return a
}

func (a *liveLogAccumulator) run() {
	ticker := time.NewTicker(taskLogFlushInterval)
	defer ticker.Stop()
	defer close(a.done)
	for {
		select {
		case <-ticker.C:
			a.flush()
		case <-a.stop:
			a.flush()
			return
		}
	}
}

func (a *liveLogAccumulator) close() {
	a.mu.Lock()
	a.appendSanitizedLocked(a.masker.flush())
	a.mu.Unlock()
	close(a.stop)
	<-a.done
}

func (a *liveLogAccumulator) append(chunk string) {
	if chunk == "" {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.appendSanitizedLocked(a.masker.write(chunk))
	if a.batch.Len() >= 32*1024 {
		a.flushLocked()
	}
}

func (a *liveLogAccumulator) appendSanitizedLocked(content string) {
	if content == "" || a.truncated {
		return
	}
	payloadLimit := maxTaskLogBytes - len(taskLogTruncatedMark)
	remaining := payloadLimit - a.total
	if len(content) > remaining {
		if remaining < 0 {
			remaining = 0
		}
		if len(content) > remaining {
			content = content[:remaining]
		}
		for len(content) > 0 && !utf8.ValidString(content) {
			content = content[:len(content)-1]
		}
		content += taskLogTruncatedMark
		a.truncated = true
	}
	if content != "" {
		a.batch.WriteString(content)
		a.total += len(content)
		a.dirty = true
	}
}

func (a *liveLogAccumulator) flush() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.flushLocked()
}

func (a *liveLogAccumulator) flushLocked() {
	if !a.dirty {
		return
	}
	content := a.batch.String()
	item := models.TaskLogChunk{TaskLogId: a.taskLogID, Seq: a.seq + 1, Content: content}
	if err := new(models.TaskLogChunk).Append([]models.TaskLogChunk{item}); err != nil {
		logger.Errorf("实时任务日志刷新失败#log-%d#%v", a.taskLogID, err)
		return
	}
	a.seq++
	a.batch.Reset()
	a.dirty = false
}

type streamingSecretMasker struct {
	secrets []string
	pending string
}

func newStreamingSecretMasker(values []string) *streamingSecretMasker {
	m := &streamingSecretMasker{}
	for _, value := range values {
		if value == "" {
			continue
		}
		m.secrets = append(m.secrets, value)
	}
	return m
}

func (m *streamingSecretMasker) write(chunk string) string {
	m.pending += chunk
	return m.drain(false)
}

func (m *streamingSecretMasker) flush() string {
	return m.drain(true)
}

func (m *streamingSecretMasker) drain(final bool) string {
	if m.pending == "" {
		return ""
	}
	var out strings.Builder
	pos := 0
	for pos < len(m.pending) {
		matchAt, match := -1, ""
		for _, secret := range m.secrets {
			idx := strings.Index(m.pending[pos:], secret)
			if idx < 0 {
				continue
			}
			idx += pos
			if matchAt < 0 || idx < matchAt || (idx == matchAt && len(secret) > len(match)) {
				matchAt, match = idx, secret
			}
		}
		if matchAt < 0 {
			break
		}
		out.WriteString(m.pending[pos:matchAt])
		out.WriteString("***")
		pos = matchAt + len(match)
	}
	remaining := m.pending[pos:]
	if final {
		out.WriteString(remaining)
		m.pending = ""
		return out.String()
	}

	// Only retain a suffix that could become a secret when the next fragment
	// arrives. Ordinary short output is released immediately instead of being
	// delayed by the length of the longest configured secret.
	keep := 0
	for _, secret := range m.secrets {
		maxPrefix := len(secret) - 1
		if maxPrefix > len(remaining) {
			maxPrefix = len(remaining)
		}
		for n := maxPrefix; n > keep; n-- {
			if strings.HasSuffix(remaining, secret[:n]) {
				keep = n
				break
			}
		}
	}
	out.WriteString(remaining[:len(remaining)-keep])
	m.pending = remaining[len(remaining)-keep:]
	return out.String()
}

func limitTaskLog(result string, alreadyTruncated bool) string {
	limit := maxTaskLogBytes
	truncated := alreadyTruncated || len(result) > limit
	if truncated {
		limit -= len(taskLogTruncatedMark)
		if limit < 0 {
			limit = 0
		}
		if len(result) > limit {
			result = result[:limit]
			for len(result) > 0 && !utf8.ValidString(result) {
				result = result[:len(result)-1]
			}
		}
		result += taskLogTruncatedMark
	}
	return result
}

// HTTP任务
type HTTPHandler struct{}

// HttpDefaultTimeout HTTP 任务默认超时（秒），用户未设置时使用
const HttpDefaultTimeout = 300

func (h *HTTPHandler) Run(taskModel models.Task, taskUniqueId int64, _ map[string]string) (result string, err error) {
	if taskModel.Timeout <= 0 {
		taskModel.Timeout = HttpDefaultTimeout
	}

	headers := strings.TrimSpace(taskModel.HttpHeaders)
	var resp httpclient.ResponseWrapper
	if taskModel.HttpMethod == models.TaskHTTPMethodGet {
		if headers != "" {
			resp = httpGetWithHeadersFunc(taskModel.Command, headers, taskModel.Timeout)
		} else {
			resp = httpGetFunc(taskModel.Command, taskModel.Timeout)
		}
	} else {
		// POST: 优先使用 HttpBody (JSON)，否则回退到 URL query 参数
		if strings.TrimSpace(taskModel.HttpBody) != "" {
			if headers != "" {
				resp = httpPostJsonWithHdrsFunc(taskModel.Command, taskModel.HttpBody, headers, taskModel.Timeout)
			} else {
				resp = httpPostJsonFunc(taskModel.Command, taskModel.HttpBody, taskModel.Timeout)
			}
		} else {
			urlFields := strings.Split(taskModel.Command, "?")
			url := urlFields[0]
			var params string
			if len(urlFields) >= 2 {
				params = urlFields[1]
			}
			if headers != "" {
				resp = httpPostParamsWithHdrs(url, params, headers, taskModel.Timeout)
			} else {
				resp = httpPostParamsFunc(url, params, taskModel.Timeout)
			}
		}
	}

	// 返回状态码非200，均为失败
	if resp.StatusCode != http.StatusOK {
		return resp.Body, fmt.Errorf("HTTP status code is not 200-->%d", resp.StatusCode)
	}

	// 响应内容断言
	if taskModel.SuccessPattern != "" {
		re, regexErr := regexp.Compile(taskModel.SuccessPattern)
		if regexErr != nil {
			return resp.Body, fmt.Errorf("invalid success_pattern regex: %v", regexErr)
		}
		// 先匹配原始响应体，不匹配再尝试压缩 JSON 后匹配（兼容格式化空白差异）
		if !re.MatchString(resp.Body) {
			compacted := compactJSON(resp.Body)
			if compacted == resp.Body || !re.MatchString(compacted) {
				return resp.Body, fmt.Errorf("response body does not match success_pattern: %s", taskModel.SuccessPattern)
			}
		}
	}

	return resp.Body, err
}

// compactJSON 压缩 JSON 字符串，去掉格式化空白。非 JSON 则原样返回。
func compactJSON(s string) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(s)); err != nil {
		return s
	}
	return buf.String()
}

// RPC调用执行任务
type RPCHandler struct{}

func (h *RPCHandler) Run(taskModel models.Task, taskUniqueId int64, secretEnv map[string]string) (result string, err error) {
	return h.RunWithProgress(taskModel, taskUniqueId, secretEnv, nil)
}

func (h *RPCHandler) RunWithProgress(taskModel models.Task, taskUniqueId int64, secretEnv map[string]string, onChunk func(string)) (result string, err error) {
	logger.Infof("RPC task execution started#Task ID-%d#Host count-%d", taskModel.Id, len(taskModel.Hosts))
	if len(taskModel.Hosts) == 0 {
		return "", fmt.Errorf("task is not associated with any host")
	}
	resultChan := make(chan TaskResult, len(taskModel.Hosts))
	for _, taskHost := range taskModel.Hosts {
		logger.Infof("Preparing RPC call#Host-%s:%d#Command-%s", taskHost.Name, taskHost.Port, taskModel.Command)
		go func(th models.TaskHostDetail) {
			// 每个 host 使用独立的 TaskRequest,避免多 goroutine 并发写同一请求指针
			// (rpcClient.Exec 内部会改写 Timeout)产生数据竞态。Env 为只读共享,安全。
			req := &pb.TaskRequest{
				Command: taskModel.Command,
				Timeout: int32(taskModel.Timeout),
				Id:      taskUniqueId,
				Env:     secretEnv,
			}
			header := fmt.Sprintf("Host: [%s-%s:%d]\n", th.Alias, th.Name, th.Port)
			if onChunk != nil {
				onChunk(header)
			}
			output, err := rpcClient.ExecStream(th.Name, th.Port, req, onChunk)
			errorMessage := ""
			if err != nil {
				// 如果是手动停止错误，保留原始错误以便后续判断，但显示翻译后的文本
				if errors.Is(err, rpcClient.ErrManualStop) {
					errorMessage = "Manually stopped"
				} else {
					errorMessage = err.Error()
				}
			}
			output = strings.TrimSpace(output)
			if errorMessage != "" {
				errorMessage = strings.TrimSpace(errorMessage) + "\n"
			}
			if errorMessage != "" && onChunk != nil {
				onChunk(errorMessage)
			}
			outputMessage := header + errorMessage + output
			logger.Infof("RPC call completed#Host-%s:%d#Output length-%d#Error-%v", th.Name, th.Port, len(output), err)
			resultChan <- TaskResult{Err: err, Result: outputMessage}
		}(taskHost)
	}

	var aggregationErr error
	var resultBuilder strings.Builder
	for i := 0; i < len(taskModel.Hosts); i++ {
		taskResult := <-resultChan
		resultBuilder.WriteString(taskResult.Result)
		if taskResult.Err != nil {
			aggregationErr = taskResult.Err
		}
	}

	return resultBuilder.String(), aggregationErr
}

// loadSecretEnv 读取机密并解密为 name->明文 映射,用于注入任务执行环境。
// 任务配置了机密白名单(SecretNames)时,仅解密注入白名单内的机密,实现任务级作用域隔离;
// 解密后的全量机密进程内缓存(name -> 明文),按 CacheSignature 失效。
// 高频任务(如每秒级)由此免去每次执行的全表查询 + 逐条 AES 解密。
var (
	secretCacheMu    sync.Mutex
	secretCacheMap   map[string]string
	secretCacheSig   string
	secretCacheValid bool
)

// resetSecretCache 使缓存失效,仅供测试在切换 models.Db 后隔离状态使用。
func resetSecretCache() {
	secretCacheMu.Lock()
	secretCacheValid = false
	secretCacheMap = nil
	secretCacheSig = ""
	secretCacheMu.Unlock()
}

// decryptAllSecrets 全表拉取并逐条解密,构建 name -> 明文 映射(跳过受保护名与解密失败项)。
func decryptAllSecrets() (map[string]string, error) {
	all, err := new(models.Secret).All()
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(all))
	failed := 0
	for _, s := range all {
		// 纵深防御:跳过受保护的系统环境变量名,防止绕过 API 写入的坏数据覆盖 PATH/LD_PRELOAD 等
		if models.IsReservedEnvName(s.Name) {
			logger.Warnf("跳过受保护的机密名#%s(不允许覆盖系统环境变量)", s.Name)
			continue
		}
		plain, err := crypto.Decrypt(s.Value)
		if err != nil {
			failed++
			continue
		}
		m[s.Name] = plain
	}
	// 汇总告警:解密失败通常意味着 GOCRON_SECRET_KEY 已变更,否则任务会静默拿不到机密
	if failed > 0 {
		logger.Errorf("有 %d 条机密解密失败,可能 GOCRON_SECRET_KEY 已变更;相关任务将无法获取这些机密,请检查主密钥或重建机密", failed)
	}
	return m, nil
}

// cachedSecrets 返回全量解密机密映射,命中缓存则免去重复解密;签名变化时重建。
// 返回的 map 为共享只读副本,调用方不得修改。
func cachedSecrets() (map[string]string, error) {
	sig, err := new(models.Secret).CacheSignature()
	if err != nil {
		// 签名查询失败:降级为直接解密(保持行为正确,只是不走缓存)
		logger.Warnf("获取机密缓存签名失败,降级为直接解密: %v", err)
		return decryptAllSecrets()
	}

	secretCacheMu.Lock()
	defer secretCacheMu.Unlock()
	if secretCacheValid && secretCacheSig == sig {
		return secretCacheMap, nil
	}
	m, err := decryptAllSecrets()
	if err != nil {
		return nil, err
	}
	secretCacheMap = m
	secretCacheSig = sig
	secretCacheValid = true
	return m, nil
}

// 未配置白名单的任务保持历史行为,注入全部机密。
// 未配置加密或读取/解密失败时安全降级(返回 nil 或跳过失败项)。
func loadSecretEnv(taskModel models.Task) map[string]string {
	if !crypto.Configured() {
		return nil
	}
	all, err := cachedSecrets()
	if err != nil {
		logger.Warnf("加载机密失败: %v", err)
		return nil
	}
	if len(all) == 0 {
		return nil
	}
	// 白名单集合;nil 表示未配置白名单(注入全部)
	var allowed map[string]bool
	if names := taskModel.SecretNameList(); names != nil {
		allowed = make(map[string]bool, len(names))
		for _, name := range names {
			allowed[name] = true
		}
	}
	env := make(map[string]string, len(all))
	for name, plain := range all {
		if allowed != nil {
			if !allowed[name] {
				continue
			}
			delete(allowed, name)
		}
		env[name] = plain
	}
	// 白名单中引用了不存在的机密名:任务将拿不到该变量,给出告警便于排查
	for name := range allowed {
		logger.Warnf("任务#%d(%s)机密白名单引用了不存在的机密#%s", taskModel.Id, taskModel.Name, name)
	}
	return env
}

// secretValues 返回机密明文值列表,用于任务输出脱敏。
func secretValues(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	values := make([]string, 0, len(env))
	for _, v := range env {
		values = append(values, v)
	}
	return values
}

// 创建任务日志
func createTaskLog(taskModel models.Task, status models.Status) (int64, error) {
	taskLogModel := new(models.TaskLog)
	taskLogModel.TaskId = taskModel.Id
	taskLogModel.Name = taskModel.Name
	taskLogModel.Spec = taskModel.Spec
	taskLogModel.Protocol = taskModel.Protocol
	taskLogModel.Command = taskModel.Command
	taskLogModel.Timeout = taskModel.Timeout
	if taskModel.Protocol == models.TaskRPC {
		var hostBuilder strings.Builder
		for _, host := range taskModel.Hosts {
			hostBuilder.WriteString(host.Alias)
			hostBuilder.WriteString(" - ")
			hostBuilder.WriteString(host.Name)
			hostBuilder.WriteString("<br>")
		}
		taskLogModel.Hostname = hostBuilder.String()
	}
	taskLogModel.StartTime = models.LocalTime(time.Now())
	taskLogModel.Status = status
	insertId, err := taskLogModel.Create()

	return insertId, err
}

// 更新任务日志
func updateTaskLog(taskLogId int64, taskResult TaskResult) (int64, error) {
	taskLogModel := new(models.TaskLog)
	var status models.Status
	result := taskResult.Result

	// 根据错误类型设置状态
	if taskResult.Err != nil {
		// 检查是否是手动停止
		if errors.Is(taskResult.Err, rpcClient.ErrManualStop) {
			status = models.Cancel
		} else {
			status = models.Failure
		}
	} else {
		status = models.Finish
	}

	return taskLogModel.Update(taskLogId, models.CommonMap{
		"retry_times": taskResult.RetryTimes,
		"status":      status,
		"result":      result,
		"end_time":    time.Now(),
	})
}

func createJob(taskModel models.Task) cron.FuncJob {
	handler := createHandler(taskModel)
	if handler == nil {
		return nil
	}
	taskFunc := func() {
		taskCount.Add()
		defer taskCount.Done()

		taskLogId := beforeExecJob(taskModel)
		if taskLogId <= 0 {
			return
		}

		// Multi=0 时，确保清理实例标记
		// 注意：beforeExecJob 已经添加了实例标记，这里只需要清理
		if taskModel.Multi == 0 {
			defer runInstance.done(taskModel.Id)
		}

		concurrencyQueue.Add()
		defer concurrencyQueue.Done()

		// 一次性加载本次执行的机密快照(按任务白名单过滤):注入与后续脱敏复用同一份,避免二次查询与不一致窗口
		secretEnv := loadSecretEnv(taskModel)
		liveLog := newLiveLogAccumulator(taskLogId, secretEnv)
		logger.Infof("Starting task execution#%s#Command-%s", taskModel.Name, taskModel.Command)
		taskResult := execJob(handler, taskModel, taskLogId, secretEnv, liveLog.append)
		liveLog.close()
		logger.Infof("Task completed#%s#Command-%s", taskModel.Name, taskModel.Command)
		afterExecJob(taskModel, taskResult, taskLogId, secretEnv)
	}

	return taskFunc
}

func createHandler(taskModel models.Task) Handler {
	var handler Handler = nil
	switch taskModel.Protocol {
	case models.TaskHTTP:
		handler = new(HTTPHandler)
	case models.TaskRPC:
		handler = new(RPCHandler)
	}

	return handler
}

// 任务前置操作
func beforeExecJob(taskModel models.Task) (taskLogId int64) {
	// Multi=0 时，原子地检查并添加实例标记
	if taskModel.Multi == 0 {
		if !runInstance.tryAdd(taskModel.Id) {
			logger.Infof("Task already running, canceling this execution#ID-%d", taskModel.Id)
			taskLogId, _ = createTaskLog(taskModel, models.Cancel)
			return
		}
	}

	taskLogId, err := createTaskLog(taskModel, models.Running)
	if err != nil {
		logger.Error("Task execution started#Failed to write task log-", err)
		// 如果创建日志失败，需要回滚实例标记
		if taskModel.Multi == 0 {
			runInstance.done(taskModel.Id)
		}
		return
	}

	return taskLogId
}

// 任务执行后置操作
func afterExecJob(taskModel models.Task, taskResult TaskResult, taskLogId int64, secretEnv map[string]string) {
	// 落库前对输出脱敏,避免机密明文写入任务日志(以及后续告警通知);
	// 复用注入时的同一份机密快照,保证脱敏与注入一致
	if values := secretValues(secretEnv); len(values) > 0 {
		taskResult.Result = crypto.MaskSecrets(taskResult.Result, values)
	}
	taskResult.Result = limitTaskLog(taskResult.Result, false)
	_, err := updateTaskLog(taskLogId, taskResult)
	if err != nil {
		logger.Error("Task ended#Failed to update task log-", err)
	} else if err := new(models.TaskLogChunk).DeleteByTaskLogId(taskLogId); err != nil {
		logger.Errorf("清理任务日志分片失败#log-%d#%v", taskLogId, err)
	}

	// 发送邮件
	go SendNotification(taskModel, taskResult)
	// 执行依赖任务
	go execDependencyTask(taskModel, taskResult)
}

// 执行依赖任务, 多个任务并发执行
func execDependencyTask(taskModel models.Task, taskResult TaskResult) {
	// 父任务才能执行子任务
	if taskModel.Level != models.TaskLevelParent {
		return
	}

	// 是否存在子任务
	dependencyTaskId := strings.TrimSpace(taskModel.DependencyTaskId)
	if dependencyTaskId == "" {
		return
	}

	// 父子任务关系为强依赖, 父任务执行失败, 不执行依赖任务
	if taskModel.DependencyStatus == models.TaskDependencyStatusStrong && taskResult.Err != nil {
		logger.Infof("Parent-child tasks have strong dependency, parent task failed, dependency tasks will not run#Parent task ID-%d", taskModel.Id)
		return
	}

	// 获取子任务
	model := new(models.Task)
	tasks, err := model.GetDependencyTaskList(dependencyTaskId)
	if err != nil {
		logger.Errorf("Failed to get dependency tasks#Parent task ID-%d#%s", taskModel.Id, err.Error())
		return
	}
	if len(tasks) == 0 {
		logger.Warnf("Dependency task list is empty or tasks are disabled#Parent task ID-%d#Dependency task ID-%s", taskModel.Id, dependencyTaskId)
		return
	}
	logger.Infof("Starting dependency tasks execution#Parent task ID-%d#Dependency task count-%d", taskModel.Id, len(tasks))
	for _, task := range tasks {
		logger.Infof("Executing dependency task#Parent task ID-%d#Dependency task ID-%d#Dependency task name-%s", taskModel.Id, task.Id, task.Name)
		task.Spec = fmt.Sprintf("Dependency task (Parent task ID-%d)", taskModel.Id)
		ServiceTask.Run(task)
	}
}

// 发送任务结果通知
// 通知触发条件位掩码(可组合):失败 / 成功 / 关键字匹配。0 表示不通知。
// 兼容旧值:旧 1(仅失败)= bit0 语义不变;旧 2(总是)、旧 3(关键字)由迁移转换。
const (
	notifyOnFailure = 1 // bit0
	notifyOnSuccess = 2 // bit1
	notifyOnKeyword = 4 // bit2
)

// matchOutputPattern 按模式对输出做匹配:子串包含或正则(标准库 RE2)。
// 正则编译失败时返回 error,失败语义由调用方决定(匹配侧不通知/排除侧不排除)。
func matchOutputPattern(pattern string, useRegex bool, output string) (bool, error) {
	if useRegex {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false, err
		}
		return re.MatchString(output), nil
	}
	return strings.Contains(output, pattern), nil
}

// matchNotifyKeyword 关键字触发条件:输出命中「关键字」且未命中「排除关键字」。
// 两个字段共用 NotifyKeywordRegex 的匹配模式;排除关键字为空时不做排除。
func matchNotifyKeyword(taskModel models.Task, output string) bool {
	kw := taskModel.NotifyKeyword
	if kw == "" {
		return false
	}
	useRegex := taskModel.NotifyKeywordRegex == 1
	matched, err := matchOutputPattern(kw, useRegex, output)
	if err != nil {
		logger.Warnf("通知关键字正则编译失败#task-%d: %v", taskModel.Id, err)
		return false
	}
	if !matched {
		return false
	}
	excl := taskModel.NotifyKeywordExclude
	if excl == "" {
		return true
	}
	excluded, err := matchOutputPattern(excl, useRegex, output)
	if err != nil {
		// 排除模式编译失败:忽略排除、照常通知(宁可多发,不可静默漏发)
		logger.Warnf("通知排除关键字正则编译失败#task-%d: %v", taskModel.Id, err)
		return true
	}
	return !excluded
}

func SendNotification(taskModel models.Task, taskResult TaskResult) {
	ns := taskModel.NotifyStatus
	// 未开启通知
	if ns == 0 {
		return
	}

	failed := taskResult.Err != nil
	// 多条件:任一勾选的条件满足即发送一条通知(OR 短路,关键字匹配仅在需要时求值)
	shouldNotify := (failed && ns&notifyOnFailure != 0) ||
		(!failed && ns&notifyOnSuccess != 0) ||
		(ns&notifyOnKeyword != 0 && matchNotifyKeyword(taskModel, taskResult.Result))
	if !shouldNotify {
		return
	}

	// NotifyType: 0=邮件, 1=Slack, 2=WebHook
	// WebHook(type=2)不需要receiver_id，其他类型需要
	if taskModel.NotifyType != 2 && taskModel.NotifyReceiverId == "" {
		return
	}
	statusName := "Success"
	if failed {
		statusName = "Failed"
	}

	output := taskResult.Result
	// 失败 + 开启诊断时,尽力附带 AI 根因分析(不阻塞:本函数已在 goroutine 中调用)
	if failed && taskModel.NotifyDiagnosis == 1 {
		if block := diagnoseForNotify(taskModel, output); block != "" {
			output = block + "\n\n" + output
		}
	}

	// 发送通知
	msg := notify.Message{
		"task_type":        taskModel.NotifyType,
		"task_receiver_id": taskModel.NotifyReceiverId,
		"name":             taskModel.Name,
		"output":           output,
		"status":           statusName,
		"task_id":          taskModel.Id,
		"remark":           taskModel.Remark,
	}
	notifyPushFunc(msg)
}

// 失败诊断通知路径的参数
const (
	// notifyDiagnosisTimeout 单次诊断在通知路径的超时:失败告警不能被拖太久,
	// 超时则发不带诊断的原通知(比 diagnosis.Timeout 短)。
	notifyDiagnosisTimeout = 30 * time.Second
	// notifyDiagnosisCooldown 同一任务两次诊断的最小间隔,防止 flapping 任务刷爆 LLM。
	notifyDiagnosisCooldown = 10 * time.Minute
)

// diagnoseCooldown 记录每个任务上次诊断时间,做限频。
var diagnoseCooldown = &cooldownTracker{last: make(map[int]time.Time)}

type cooldownTracker struct {
	mu   sync.Mutex
	last map[int]time.Time
}

// allow 在距上次放行超过 window 时返回 true 并记录 now;否则返回 false。
func (c *cooldownTracker) allow(id int, now time.Time, window time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if t, ok := c.last[id]; ok && now.Sub(t) < window {
		return false
	}
	c.last[id] = now
	return true
}

// formatDiagnosisBlock 把诊断结果渲染成附加到通知输出前的文本块。
func formatDiagnosisBlock(d diagnosis.Result, english bool) string {
	if strings.TrimSpace(d.RootCause) == "" && len(d.Suggestions) == 0 {
		return ""
	}
	title, cause, advice := "【AI 诊断】", "根因", "建议"
	if english {
		title, cause, advice = "[AI Diagnosis]", "Root cause", "Suggestions"
	}
	var b strings.Builder
	b.WriteString(title + "\n")
	if strings.TrimSpace(d.RootCause) != "" {
		fmt.Fprintf(&b, "%s: %s\n", cause, strings.TrimSpace(d.RootCause))
	}
	if len(d.Suggestions) > 0 {
		b.WriteString(advice + ":\n")
		for _, s := range d.Suggestions {
			if strings.TrimSpace(s) != "" {
				fmt.Fprintf(&b, "- %s\n", strings.TrimSpace(s))
			}
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// diagnoseForNotify 尽力为失败任务生成诊断文本块;任何前置条件不满足(未配 LLM、
// 限频命中、调用失败/超时)都返回空串,让通知照常不带诊断发出。
func diagnoseForNotify(taskModel models.Task, output string) string {
	client, err := llm.FromSettings()
	if err != nil {
		return "" // 未配置 LLM,静默跳过
	}
	if !diagnoseCooldown.allow(taskModel.Id, time.Now(), notifyDiagnosisCooldown) {
		return ""
	}
	// 复用诊断模块的输入构造(内部会截断过长输出);通知路径无请求上下文,默认中文
	log := &models.TaskLog{
		Name:     taskModel.Name,
		Protocol: taskModel.Protocol,
		Command:  taskModel.Command,
		Result:   output,
	}
	ctx, cancel := context.WithTimeout(context.Background(), notifyDiagnosisTimeout)
	defer cancel()
	d, err := diagnosis.Diagnose(ctx, client, log, false)
	if err != nil {
		logger.Warnf("失败通知诊断#task-%d#%v", taskModel.Id, err)
		return ""
	}
	return formatDiagnosisBlock(d, false)
}

// 执行具体任务
func execJob(handler Handler, taskModel models.Task, taskUniqueId int64, secretEnv map[string]string, onChunk func(string)) (result TaskResult) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("panic#service/task.go:execJob#", err)
			// 确保 panic 不会被误判为成功：返回失败结果
			result = TaskResult{Err: fmt.Errorf("panic: %v", err)}
		}
	}()
	// 默认只运行任务一次
	var execTimes int8 = 1
	if taskModel.RetryTimes > 0 {
		execTimes += taskModel.RetryTimes
	}
	var i int8 = 0
	var output string
	var err error
	for i < execTimes {
		if streaming, ok := handler.(progressHandler); ok {
			output, err = streaming.RunWithProgress(taskModel, taskUniqueId, secretEnv, onChunk)
		} else {
			output, err = handler.Run(taskModel, taskUniqueId, secretEnv)
		}
		if err == nil {
			return TaskResult{Result: output, Err: err, RetryTimes: i}
		}
		i++
		if i < execTimes {
			// 不打印 output:重试阶段的原始输出可能含注入的机密明文,完整输出会在 afterExecJob 脱敏后落库
			logger.Warnf("Task execution failed#Task ID-%d#Retry attempt %d#Error-%s", taskModel.Id, i, err.Error())
			if taskModel.RetryInterval > 0 {
				sleepFunc(time.Duration(taskModel.RetryInterval) * time.Second)
			} else {
				// 默认重试间隔时间，每次递增1分钟
				sleepFunc(time.Duration(i) * time.Minute)
			}
		}
	}

	return TaskResult{Result: output, Err: err, RetryTimes: taskModel.RetryTimes}
}

// 清理日志文件
func cleanupLogFiles() {
	settingModel := new(models.Setting)
	fileSizeLimit := settingModel.GetLogFileSizeLimit()

	// 如果设置为0，不清理日志文件
	if fileSizeLimit <= 0 {
		return
	}

	logDir := "log"
	logFile := "cron.log"

	// 检查日志文件是否存在
	logPath := fmt.Sprintf("%s/%s", logDir, logFile)
	fileInfo, err := os.Stat(logPath)
	if err != nil {
		if !os.IsNotExist(err) {
			logger.Errorf("Failed to check log file: %s", err)
		}
		return
	}

	// 如果文件大小超过限制，则清空
	maxSize := int64(fileSizeLimit) * 1024 * 1024 // 转换为MB
	if fileInfo.Size() > maxSize {
		err := os.Truncate(logPath, 0)
		if err != nil {
			logger.Errorf("Failed to truncate log file: %s", err)
		} else {
			logger.Infof("Log file exceeded %dMB, truncated: %s", fileSizeLimit, logPath)
		}
	}
}
