package service

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gocronx-team/gocron/internal/models"
)

// 模拟RPC处理器，用于测试部分输出功能
type mockRPCHandlerWithPartialOutput struct {
	partialOutput string
	errorType     string // "timeout", "manual_stop", "normal_error", "success"
}

func (m *mockRPCHandlerWithPartialOutput) Run(taskModel models.Task, taskUniqueId int64, _ map[string]string) (string, error) {
	switch m.errorType {
	case "timeout":
		// 模拟超时情况，返回部分输出和超时错误
		return m.partialOutput + "\n\n[执行超时，强制结束]",
			fmt.Errorf("执行超时, 强制结束")
	case "manual_stop":
		// 模拟手动停止情况，返回部分输出和手动停止错误
		return m.partialOutput + "\n\n[手动停止]",
			fmt.Errorf("手动停止")
	case "normal_error":
		// 模拟普通错误，返回完整输出
		return m.partialOutput, fmt.Errorf("command failed")
	case "success":
		// 模拟成功情况
		return m.partialOutput, nil
	default:
		return "", fmt.Errorf("unknown error type")
	}
}

func TestExecJobWithPartialOutput_Timeout(t *testing.T) {
	handler := &mockRPCHandlerWithPartialOutput{
		partialOutput: "Task started\nProcessing data...\nPartial result: 50%",
		errorType:     "timeout",
	}

	task := models.Task{
		Id:         1,
		Name:       "timeout-test",
		RetryTimes: 0, // 不重试，直接测试超时
	}

	result := execJob(handler, task, 1, nil, nil)

	if result.Err == nil {
		t.Fatal("Expected timeout error")
	}
	if result.Err.Error() != "执行超时, 强制结束" {
		t.Fatalf("Expected timeout error, got: %v", result.Err)
	}
	if !strings.Contains(result.Result, "Task started") {
		t.Fatalf("Expected partial output to contain 'Task started', got: %s", result.Result)
	}
	if !strings.Contains(result.Result, "Partial result: 50%") {
		t.Fatalf("Expected partial output to contain 'Partial result: 50%%', got: %s", result.Result)
	}
	if !strings.Contains(result.Result, "[执行超时，强制结束]") {
		t.Fatalf("Expected timeout marker in output, got: %s", result.Result)
	}
}

func TestExecJobWithPartialOutput_ManualStop(t *testing.T) {
	handler := &mockRPCHandlerWithPartialOutput{
		partialOutput: "Task started\nProcessing batch 1\nProcessing batch 2",
		errorType:     "manual_stop",
	}

	task := models.Task{
		Id:         2,
		Name:       "manual-stop-test",
		RetryTimes: 0,
	}

	result := execJob(handler, task, 2, nil, nil)

	if result.Err == nil {
		t.Fatal("Expected manual stop error")
	}
	if result.Err.Error() != "手动停止" {
		t.Fatalf("Expected manual stop error, got: %v", result.Err)
	}
	if !strings.Contains(result.Result, "Processing batch 1") {
		t.Fatalf("Expected partial output to contain 'Processing batch 1', got: %s", result.Result)
	}
	if !strings.Contains(result.Result, "[手动停止]") {
		t.Fatalf("Expected manual stop marker in output, got: %s", result.Result)
	}
}

func TestExecJobWithPartialOutput_NormalError(t *testing.T) {
	handler := &mockRPCHandlerWithPartialOutput{
		partialOutput: "Task started\nError occurred: file not found",
		errorType:     "normal_error",
	}

	task := models.Task{
		Id:         3,
		Name:       "error-test",
		RetryTimes: 0,
	}

	result := execJob(handler, task, 3, nil, nil)

	if result.Err == nil {
		t.Fatal("Expected normal error")
	}
	if result.Err.Error() != "command failed" {
		t.Fatalf("Expected 'command failed' error, got: %v", result.Err)
	}
	// 对于普通错误，应该返回完整输出，不添加特殊标记
	if !strings.Contains(result.Result, "Error occurred: file not found") {
		t.Fatalf("Expected full output for normal error, got: %s", result.Result)
	}
	if strings.Contains(result.Result, "[执行超时，强制结束]") || strings.Contains(result.Result, "[手动停止]") {
		t.Fatalf("Should not contain timeout/stop markers for normal error, got: %s", result.Result)
	}
}

func TestExecJobWithPartialOutput_Success(t *testing.T) {
	handler := &mockRPCHandlerWithPartialOutput{
		partialOutput: "Task started\nProcessing completed\nResult: success",
		errorType:     "success",
	}

	task := models.Task{
		Id:         4,
		Name:       "success-test",
		RetryTimes: 0,
	}

	result := execJob(handler, task, 4, nil, nil)

	if result.Err != nil {
		t.Fatalf("Expected no error for success, got: %v", result.Err)
	}
	if !strings.Contains(result.Result, "Result: success") {
		t.Fatalf("Expected success output, got: %s", result.Result)
	}
	if strings.Contains(result.Result, "[执行超时，强制结束]") || strings.Contains(result.Result, "[手动停止]") {
		t.Fatalf("Should not contain error markers for success, got: %s", result.Result)
	}
}

// 可重写的假处理器，用于测试
type overridableHandler struct {
	runFunc func(models.Task, int64) (string, error)
}

func (h *overridableHandler) Run(taskModel models.Task, taskUniqueId int64, _ map[string]string) (string, error) {
	return h.runFunc(taskModel, taskUniqueId)
}

func TestExecJobWithPartialOutput_RetryWithTimeout(t *testing.T) {
	// 测试重试机制与部分输出的结合
	callCount := 0
	results := []handlerResponse{
		{result: "First attempt\nPartial output\n\n[执行超时，强制结束]", err: fmt.Errorf("执行超时, 强制结束")},
		{result: "Second attempt\nSuccess!", err: nil},
	}

	handler := &overridableHandler{
		runFunc: func(taskModel models.Task, taskUniqueId int64) (string, error) {
			result := results[callCount]
			callCount++
			return result.result, result.err
		},
	}

	task := models.Task{
		Id:            5,
		Name:          "retry-test",
		RetryTimes:    1,
		RetryInterval: 0, // 不等待，加快测试
	}

	// 模拟sleep函数，避免实际等待
	originalSleep := sleepFunc
	sleepFunc = func(d time.Duration) {
		// 不实际睡眠
	}
	defer func() { sleepFunc = originalSleep }()

	result := execJob(handler, task, 5, nil, nil)

	if result.Err != nil {
		t.Fatalf("Expected success after retry, got: %v", result.Err)
	}
	if !strings.Contains(result.Result, "Second attempt") {
		t.Fatalf("Expected second attempt output, got: %s", result.Result)
	}
	if result.RetryTimes != 1 {
		t.Fatalf("Expected 1 retry, got: %d", result.RetryTimes)
	}
	if callCount != 2 {
		t.Fatalf("Expected 2 handler calls, got: %d", callCount)
	}
}
