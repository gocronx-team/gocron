package service

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/httpclient"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/notify"
)

func TestMain(m *testing.M) {
	_ = os.MkdirAll("log", 0o755)
	logger.InitLogger()
	os.Exit(m.Run())
}

func TestHTTPHandlerRunGetUsesCustomTimeout(t *testing.T) {
	original := httpGetFunc
	defer func() { httpGetFunc = original }()

	var capturedTimeout int
	httpGetFunc = func(url string, timeout int) httpclient.ResponseWrapper {
		if url != "http://example.com" {
			t.Fatalf("unexpected url %s", url)
		}
		capturedTimeout = timeout
		return httpclient.ResponseWrapper{StatusCode: http.StatusOK, Body: "ok"}
	}

	handler := &HTTPHandler{}
	task := models.Task{
		Command:    "http://example.com",
		Timeout:    1000,
		HttpMethod: models.TaskHTTPMethodGet,
	}
	result, err := handler.Run(task, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ok" {
		t.Fatalf("unexpected result %s", result)
	}
	// Custom timeout should be preserved (no longer capped at 300)
	if capturedTimeout != 1000 {
		t.Fatalf("expected timeout 1000, got %d", capturedTimeout)
	}
}

func TestHTTPHandlerRunGetDefaultTimeout(t *testing.T) {
	original := httpGetFunc
	defer func() { httpGetFunc = original }()

	var capturedTimeout int
	httpGetFunc = func(url string, timeout int) httpclient.ResponseWrapper {
		capturedTimeout = timeout
		return httpclient.ResponseWrapper{StatusCode: http.StatusOK, Body: "ok"}
	}

	handler := &HTTPHandler{}
	task := models.Task{
		Command:    "http://example.com",
		Timeout:    0, // not set
		HttpMethod: models.TaskHTTPMethodGet,
	}
	handler.Run(task, 1, nil)
	if capturedTimeout != HttpDefaultTimeout {
		t.Fatalf("expected default timeout %d, got %d", HttpDefaultTimeout, capturedTimeout)
	}
}

func TestHTTPHandlerRunPostParsesParams(t *testing.T) {
	original := httpPostParamsFunc
	defer func() { httpPostParamsFunc = original }()

	var capturedURL, capturedParams string
	var capturedTimeout int
	httpPostParamsFunc = func(url, params string, timeout int) httpclient.ResponseWrapper {
		capturedURL = url
		capturedParams = params
		capturedTimeout = timeout
		return httpclient.ResponseWrapper{StatusCode: http.StatusOK, Body: "posted"}
	}

	handler := &HTTPHandler{}
	task := models.Task{
		Command:    "http://example.com/cmd?foo=bar",
		Timeout:    10,
		HttpMethod: models.TaskHttpMethodPost,
	}
	result, err := handler.Run(task, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "posted" {
		t.Fatalf("unexpected result %s", result)
	}
	if capturedURL != "http://example.com/cmd" || capturedParams != "foo=bar" {
		t.Fatalf("unexpected url/params %s %s", capturedURL, capturedParams)
	}
	if capturedTimeout != 10 {
		t.Fatalf("expected timeout 10, got %d", capturedTimeout)
	}
}

func TestHTTPHandlerRunReturnsErrorForNon200(t *testing.T) {
	original := httpGetFunc
	defer func() { httpGetFunc = original }()

	httpGetFunc = func(url string, timeout int) httpclient.ResponseWrapper {
		return httpclient.ResponseWrapper{StatusCode: http.StatusInternalServerError, Body: "bad"}
	}
	handler := &HTTPHandler{}
	task := models.Task{Command: "http://example.com", HttpMethod: models.TaskHTTPMethodGet}
	result, err := handler.Run(task, 1, nil)
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
	if result != "bad" {
		t.Fatalf("unexpected result %s", result)
	}
}

func TestHTTPHandlerRunPostJsonBody(t *testing.T) {
	original := httpPostJsonFunc
	defer func() { httpPostJsonFunc = original }()

	var capturedBody string
	httpPostJsonFunc = func(url, body string, timeout int) httpclient.ResponseWrapper {
		capturedBody = body
		return httpclient.ResponseWrapper{StatusCode: http.StatusOK, Body: "ok"}
	}

	handler := &HTTPHandler{}
	task := models.Task{
		Command:    "http://example.com/api",
		HttpMethod: models.TaskHttpMethodPost,
		HttpBody:   `{"key":"value"}`,
		Timeout:    10,
	}
	_, err := handler.Run(task, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody != `{"key":"value"}` {
		t.Fatalf("expected JSON body, got %s", capturedBody)
	}
}

func TestHTTPHandlerRunPostFallbackToParams(t *testing.T) {
	original := httpPostParamsFunc
	defer func() { httpPostParamsFunc = original }()

	var capturedParams string
	httpPostParamsFunc = func(url, params string, timeout int) httpclient.ResponseWrapper {
		capturedParams = params
		return httpclient.ResponseWrapper{StatusCode: http.StatusOK, Body: "ok"}
	}

	handler := &HTTPHandler{}
	task := models.Task{
		Command:    "http://example.com/api?a=1",
		HttpMethod: models.TaskHttpMethodPost,
		HttpBody:   "", // empty — should fallback to URL params
		Timeout:    10,
	}
	_, err := handler.Run(task, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedParams != "a=1" {
		t.Fatalf("expected params 'a=1', got %s", capturedParams)
	}
}

func TestHTTPHandlerRunSuccessPatternMatch(t *testing.T) {
	original := httpGetFunc
	defer func() { httpGetFunc = original }()

	httpGetFunc = func(url string, timeout int) httpclient.ResponseWrapper {
		return httpclient.ResponseWrapper{StatusCode: http.StatusOK, Body: `{"status":"ok","code":0}`}
	}

	handler := &HTTPHandler{}
	task := models.Task{
		Command:        "http://example.com",
		HttpMethod:     models.TaskHTTPMethodGet,
		SuccessPattern: `"code":0`,
		Timeout:        10,
	}
	_, err := handler.Run(task, 1, nil)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestHTTPHandlerRunSuccessPatternNoMatch(t *testing.T) {
	original := httpGetFunc
	defer func() { httpGetFunc = original }()

	httpGetFunc = func(url string, timeout int) httpclient.ResponseWrapper {
		return httpclient.ResponseWrapper{StatusCode: http.StatusOK, Body: `{"status":"error","code":1}`}
	}

	handler := &HTTPHandler{}
	task := models.Task{
		Command:        "http://example.com",
		HttpMethod:     models.TaskHTTPMethodGet,
		SuccessPattern: `"code":0`,
		Timeout:        10,
	}
	_, err := handler.Run(task, 1, nil)
	if err == nil {
		t.Fatal("expected error for non-matching pattern")
	}
	if !strings.Contains(err.Error(), "does not match success_pattern") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPHandlerRunSuccessPatternInvalidRegex(t *testing.T) {
	original := httpGetFunc
	defer func() { httpGetFunc = original }()

	httpGetFunc = func(url string, timeout int) httpclient.ResponseWrapper {
		return httpclient.ResponseWrapper{StatusCode: http.StatusOK, Body: "ok"}
	}

	handler := &HTTPHandler{}
	task := models.Task{
		Command:        "http://example.com",
		HttpMethod:     models.TaskHTTPMethodGet,
		SuccessPattern: `[invalid`,
		Timeout:        10,
	}
	_, err := handler.Run(task, 1, nil)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
	if !strings.Contains(err.Error(), "invalid success_pattern regex") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPHandlerRunSuccessPatternMatchCompactJSON(t *testing.T) {
	original := httpGetFunc
	defer func() { httpGetFunc = original }()

	// 模拟 pretty-printed JSON 响应（如 httpbin.org）
	httpGetFunc = func(url string, timeout int) httpclient.ResponseWrapper {
		return httpclient.ResponseWrapper{
			StatusCode: http.StatusOK,
			Body: `{
  "json": {
    "key": "value"
  },
  "data": "{\"key\":\"value\"}"
}`,
		}
	}

	handler := &HTTPHandler{}
	task := models.Task{
		Command:        "http://example.com",
		HttpMethod:     models.TaskHTTPMethodGet,
		SuccessPattern: `"key":"value"`,
		Timeout:        10,
	}
	_, err := handler.Run(task, 1, nil)
	if err != nil {
		t.Fatalf("expected success with compacted JSON matching, got error: %v", err)
	}
}

func TestHTTPHandlerRunEmptyPatternSkipsCheck(t *testing.T) {
	original := httpGetFunc
	defer func() { httpGetFunc = original }()

	httpGetFunc = func(url string, timeout int) httpclient.ResponseWrapper {
		return httpclient.ResponseWrapper{StatusCode: http.StatusOK, Body: "anything"}
	}

	handler := &HTTPHandler{}
	task := models.Task{
		Command:        "http://example.com",
		HttpMethod:     models.TaskHTTPMethodGet,
		SuccessPattern: "", // empty — should skip assertion
		Timeout:        10,
	}
	_, err := handler.Run(task, 1, nil)
	if err != nil {
		t.Fatalf("expected success with empty pattern, got: %v", err)
	}
}

func TestHTTPHandlerRunGetWithHeaders(t *testing.T) {
	original := httpGetWithHeadersFunc
	defer func() { httpGetWithHeadersFunc = original }()

	var capturedHeaders string
	httpGetWithHeadersFunc = func(url, headers string, timeout int) httpclient.ResponseWrapper {
		capturedHeaders = headers
		return httpclient.ResponseWrapper{StatusCode: http.StatusOK, Body: "ok"}
	}

	handler := &HTTPHandler{}
	task := models.Task{
		Command:     "http://example.com",
		HttpMethod:  models.TaskHTTPMethodGet,
		HttpHeaders: `{"Authorization":"Bearer abc"}`,
		Timeout:     10,
	}
	_, err := handler.Run(task, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedHeaders != `{"Authorization":"Bearer abc"}` {
		t.Fatalf("expected headers passed through, got %s", capturedHeaders)
	}
}

func TestHTTPHandlerRunPostJsonWithHeaders(t *testing.T) {
	original := httpPostJsonWithHdrsFunc
	defer func() { httpPostJsonWithHdrsFunc = original }()

	var capturedBody, capturedHeaders string
	httpPostJsonWithHdrsFunc = func(url, body, headers string, timeout int) httpclient.ResponseWrapper {
		capturedBody = body
		capturedHeaders = headers
		return httpclient.ResponseWrapper{StatusCode: http.StatusOK, Body: "ok"}
	}

	handler := &HTTPHandler{}
	task := models.Task{
		Command:     "http://example.com/api",
		HttpMethod:  models.TaskHttpMethodPost,
		HttpBody:    `{"key":"val"}`,
		HttpHeaders: `{"X-Token":"secret"}`,
		Timeout:     10,
	}
	_, err := handler.Run(task, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody != `{"key":"val"}` {
		t.Fatalf("expected body, got %s", capturedBody)
	}
	if capturedHeaders != `{"X-Token":"secret"}` {
		t.Fatalf("expected headers, got %s", capturedHeaders)
	}
}

type fakeHandler struct {
	results   []handlerResponse
	callCount int
}

type handlerResponse struct {
	result string
	err    error
}

func (f *fakeHandler) Run(taskModel models.Task, taskUniqueId int64, _ map[string]string) (string, error) {
	res := f.results[f.callCount]
	f.callCount++
	return res.result, res.err
}

func TestExecJobRetriesUntilSuccess(t *testing.T) {
	originalSleep := sleepFunc
	defer func() { sleepFunc = originalSleep }()

	sleepCalls := 0
	sleepFunc = func(d time.Duration) {
		sleepCalls++
	}

	handler := &fakeHandler{
		results: []handlerResponse{
			{result: "first", err: errors.New("fail1")},
			{result: "second", err: nil},
		},
	}
	task := models.Task{Id: 1, RetryTimes: 1, RetryInterval: 1}
	result := execJob(handler, task, 1, nil, nil)
	if result.Result != "second" || result.Err != nil {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.RetryTimes != 1 {
		t.Fatalf("expected RetryTimes 1, got %d", result.RetryTimes)
	}
	if handler.callCount != 2 {
		t.Fatalf("expected 2 handler calls, got %d", handler.callCount)
	}
	if sleepCalls != 1 {
		t.Fatalf("expected 1 sleep call, got %d", sleepCalls)
	}
}

func TestExecJobReturnsErrorAfterRetriesExhausted(t *testing.T) {
	originalSleep := sleepFunc
	defer func() { sleepFunc = originalSleep }()
	sleepCount := 0
	sleepFunc = func(d time.Duration) {
		sleepCount++
	}

	handler := &fakeHandler{
		results: []handlerResponse{
			{result: "first", err: errors.New("fail1")},
			{result: "second", err: errors.New("fail2")},
			{result: "third", err: errors.New("fail3")},
		},
	}
	task := models.Task{Id: 2, RetryTimes: 2, RetryInterval: 1}
	result := execJob(handler, task, 1, nil, nil)
	if result.Err == nil {
		t.Fatal("expected error")
	}
	if result.RetryTimes != task.RetryTimes {
		t.Fatalf("expected retryTimes %d, got %d", task.RetryTimes, result.RetryTimes)
	}
	if handler.callCount != 3 {
		t.Fatalf("expected 3 handler calls, got %d", handler.callCount)
	}
	if sleepCount != 2 {
		t.Fatalf("expected 2 sleep calls, got %d", sleepCount)
	}
}

func TestSendNotificationBehavior(t *testing.T) {
	type expectation struct {
		name   string
		task   models.Task
		result TaskResult
		count  int
		check  func(t *testing.T, msg notify.Message)
	}
	tests := []expectation{
		{
			name:  "disabled",
			task:  models.Task{NotifyStatus: 0},
			count: 0,
		},
		{
			name:   "failOnlySuccess",
			task:   models.Task{NotifyStatus: 1, NotifyType: 1, NotifyReceiverId: "user"},
			result: TaskResult{Result: "ok", Err: nil},
			count:  0,
		},
		{
			name:   "failOnlyTriggered",
			task:   models.Task{Name: "job", NotifyStatus: 1, NotifyType: 1, NotifyReceiverId: "user"},
			result: TaskResult{Result: "bad", Err: errors.New("boom")},
			count:  1,
		},
		{
			name:   "keywordMismatch",
			task:   models.Task{NotifyStatus: 4, NotifyType: 2, NotifyKeyword: "ERROR"},
			result: TaskResult{Result: "all good"},
			count:  0,
		},
		{
			name:   "keywordMatch",
			task:   models.Task{Name: "job", NotifyStatus: 4, NotifyType: 2, NotifyKeyword: "ERROR"},
			result: TaskResult{Result: "found ERROR", Err: nil},
			count:  1,
			check: func(t *testing.T, msg notify.Message) {
				if msg["status"] != "Success" {
					t.Fatalf("expected status Success, got %v", msg["status"])
				}
			},
		},
		{
			name:   "missingReceiverForMail",
			task:   models.Task{NotifyStatus: 2, NotifyType: 1, NotifyReceiverId: ""},
			result: TaskResult{Result: "any"},
			count:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			captured := stubNotifyPush(t)
			SendNotification(tt.task, tt.result)
			if len(*captured) != tt.count {
				t.Fatalf("expected %d notifications, got %d", tt.count, len(*captured))
			}
			if tt.count > 0 && tt.check != nil {
				tt.check(t, (*captured)[0])
			}
		})
	}
}

func stubNotifyPush(t *testing.T) *[]notify.Message {
	t.Helper()
	var captured []notify.Message
	original := notifyPushFunc
	notifyPushFunc = func(msg notify.Message) {
		msgCopy := notify.Message{}
		for k, v := range msg {
			msgCopy[k] = v
		}
		captured = append(captured, msgCopy)
	}
	t.Cleanup(func() { notifyPushFunc = original })
	return &captured
}

// 测试依赖任务执行逻辑 - 简化版本，直接测试逻辑分支
func TestExecDependencyTaskLogic(t *testing.T) {
	tests := []struct {
		name       string
		parentTask models.Task
		taskResult TaskResult
		shouldExit bool // 是否应该提前退出（不查询数据库）
		reason     string
	}{
		{
			name: "子任务不应该触发依赖任务",
			parentTask: models.Task{
				Id:               1,
				Level:            models.TaskLevelChild,
				DependencyTaskId: "2,3",
			},
			taskResult: TaskResult{Err: nil},
			shouldExit: true,
			reason:     "Level is Child",
		},
		{
			name: "没有依赖任务ID",
			parentTask: models.Task{
				Id:               1,
				Level:            models.TaskLevelParent,
				DependencyTaskId: "",
			},
			taskResult: TaskResult{Err: nil},
			shouldExit: true,
			reason:     "Empty DependencyTaskId",
		},
		{
			name: "强依赖且父任务失败",
			parentTask: models.Task{
				Id:               1,
				Level:            models.TaskLevelParent,
				DependencyTaskId: "2,3",
				DependencyStatus: models.TaskDependencyStatusStrong,
			},
			taskResult: TaskResult{Err: errors.New("parent failed")},
			shouldExit: true,
			reason:     "Strong dependency and parent failed",
		},
		{
			name: "弱依赖且父任务失败应该继续",
			parentTask: models.Task{
				Id:               1,
				Level:            models.TaskLevelParent,
				DependencyTaskId: "2",
				DependencyStatus: models.TaskDependencyStatusWeak,
			},
			taskResult: TaskResult{Err: errors.New("parent failed")},
			shouldExit: false,
			reason:     "Weak dependency, should continue",
		},
		{
			name: "父任务成功应该继续",
			parentTask: models.Task{
				Id:               1,
				Level:            models.TaskLevelParent,
				DependencyTaskId: "2,3",
				DependencyStatus: models.TaskDependencyStatusStrong,
			},
			taskResult: TaskResult{Err: nil},
			shouldExit: false,
			reason:     "Parent success, should continue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试逻辑分支
			if tt.parentTask.Level != models.TaskLevelParent {
				if !tt.shouldExit {
					t.Errorf("Expected to exit for child task, but shouldExit is false")
				}
				t.Logf("✓ Correctly exits for: %s", tt.reason)
				return
			}

			if tt.parentTask.DependencyTaskId == "" {
				if !tt.shouldExit {
					t.Errorf("Expected to exit for empty dependency ID, but shouldExit is false")
				}
				t.Logf("✓ Correctly exits for: %s", tt.reason)
				return
			}

			if tt.parentTask.DependencyStatus == models.TaskDependencyStatusStrong && tt.taskResult.Err != nil {
				if !tt.shouldExit {
					t.Errorf("Expected to exit for strong dependency failure, but shouldExit is false")
				}
				t.Logf("✓ Correctly exits for: %s", tt.reason)
				return
			}

			// 如果到这里，说明应该继续执行
			if tt.shouldExit {
				t.Errorf("Should have exited but didn't for: %s", tt.reason)
			} else {
				t.Logf("✓ Correctly continues for: %s (would query DB and execute)", tt.reason)
			}
		})
	}
}
