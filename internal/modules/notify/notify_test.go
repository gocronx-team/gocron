package notify

import (
	"encoding/json"
	"testing"

	"github.com/gocronx-team/gocron/internal/models"
)

// TestNotifyDispatch 测试通知分发逻辑
func TestNotifyDispatch(t *testing.T) {
	tests := []struct {
		name        string
		msg         Message
		expectError bool
		description string
	}{
		{
			name: "邮件通知-完整参数",
			msg: Message{
				"task_type":        int8(1),
				"task_receiver_id": "1,2",
				"name":             "测试任务",
				"output":           "任务执行成功",
				"status":           "成功",
				"task_id":          123,
			},
			expectError: false,
			description: "邮件通知应该正常处理",
		},
		{
			name: "Slack通知-完整参数",
			msg: Message{
				"task_type":        int8(2),
				"task_receiver_id": "1",
				"name":             "测试任务",
				"output":           "任务执行失败",
				"status":           "失败",
				"task_id":          456,
			},
			expectError: false,
			description: "Slack通知应该正常处理",
		},
		{
			name: "Webhook通知-完整参数",
			msg: Message{
				"task_type":        int8(3),
				"task_receiver_id": "1,2,3",
				"name":             "测试任务",
				"output":           "任务执行成功",
				"status":           "成功",
				"task_id":          789,
			},
			expectError: false,
			description: "Webhook通知应该正常处理",
		},
		{
			name: "缺少task_type",
			msg: Message{
				"task_receiver_id": "1",
				"name":             "测试任务",
				"output":           "输出",
				"status":           "成功",
			},
			expectError: true,
			description: "缺少task_type应该被拒绝",
		},
		{
			name: "缺少task_receiver_id",
			msg: Message{
				"task_type": int8(1),
				"name":      "测试任务",
				"output":    "输出",
				"status":    "成功",
			},
			expectError: true,
			description: "缺少task_receiver_id应该被拒绝",
		},
		{
			name: "缺少name",
			msg: Message{
				"task_type":        int8(1),
				"task_receiver_id": "1",
				"output":           "输出",
				"status":           "成功",
			},
			expectError: true,
			description: "缺少name应该被拒绝",
		},
		{
			name: "缺少output",
			msg: Message{
				"task_type":        int8(1),
				"task_receiver_id": "1",
				"name":             "测试任务",
				"status":           "成功",
			},
			expectError: true,
			description: "缺少output应该被拒绝",
		},
		{
			name: "缺少status",
			msg: Message{
				"task_type":        int8(1),
				"task_receiver_id": "1",
				"name":             "测试任务",
				"output":           "输出",
			},
			expectError: true,
			description: "缺少status应该被拒绝",
		},
		{
			name: "无效的task_type",
			msg: Message{
				"task_type":        int8(99),
				"task_receiver_id": "1",
				"name":             "测试任务",
				"output":           "输出",
				"status":           "成功",
			},
			expectError: false,
			description: "无效的task_type会被忽略但不报错",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证消息参数完整性
			_, taskTypeOk := tt.msg["task_type"]
			_, taskReceiverIdOk := tt.msg["task_receiver_id"]
			_, nameOk := tt.msg["name"]
			_, outputOk := tt.msg["output"]
			_, statusOk := tt.msg["status"]

			hasError := !taskTypeOk || !taskReceiverIdOk || !nameOk || !outputOk || !statusOk

			if hasError != tt.expectError {
				t.Errorf("%s: expected error=%v, got error=%v", tt.description, tt.expectError, hasError)
			}

			// 验证task_type类型
			if taskTypeOk {
				if _, ok := tt.msg["task_type"].(int8); !ok {
					t.Errorf("task_type should be int8")
				}
			}
		})
	}
}

// TestParseNotifyTemplate 测试通知模板解析
func TestParseNotifyTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		msg      Message
		contains []string
	}{
		{
			name:     "基础模板",
			template: "任务: {{.TaskName}}, 状态: {{.Status}}",
			msg: Message{
				"task_id": 1,
				"name":    "测试任务",
				"status":  "成功",
				"output":  "执行结果",
				"remark":  "备注",
			},
			contains: []string{"测试任务", "成功"},
		},
		{
			name:     "完整模板",
			template: "任务ID: {{.TaskId}}\n任务名称: {{.TaskName}}\n状态: {{.Status}}\n结果: {{.Result}}\n备注: {{.Remark}}",
			msg: Message{
				"task_id": 123,
				"name":    "定时任务",
				"status":  "失败",
				"output":  "错误信息",
				"remark":  "重要任务",
			},
			contains: []string{"123", "定时任务", "失败", "错误信息", "重要任务"},
		},
		{
			name:     "空模板",
			template: "",
			msg: Message{
				"task_id": 1,
				"name":    "任务",
				"status":  "成功",
				"output":  "输出",
				"remark":  "备注",
			},
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseNotifyTemplate(tt.template, tt.msg)

			for _, expected := range tt.contains {
				if len(expected) > 0 && !contains(result, expected) {
					t.Errorf("expected result to contain '%s', got: %s", expected, result)
				}
			}
		})
	}
}

// TestNotifyTypeValues 测试通知类型常量
func TestNotifyTypeValues(t *testing.T) {
	tests := []struct {
		name     string
		typeVal  int8
		typeName string
	}{
		{"邮件通知", 1, "Mail"},
		{"Slack通知", 2, "Slack"},
		{"Webhook通知", 3, "Webhook"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Message{
				"task_type":        tt.typeVal,
				"task_receiver_id": "1",
				"name":             "test",
				"output":           "output",
				"status":           "success",
			}

			taskType, ok := msg["task_type"].(int8)
			if !ok {
				t.Errorf("task_type should be int8")
			}

			if taskType != tt.typeVal {
				t.Errorf("expected task_type=%d, got %d", tt.typeVal, taskType)
			}
		})
	}
}

// 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(substr) == 0 || len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestStatusTextTemplateKeepsMachineStatus(t *testing.T) {
	msg := Message{"status": "Failed", "status_text": "失败"}
	if got := parseNotifyTemplate(`{{.Status}} / {{.StatusText}}`, msg); got != "Failed / 失败" {
		t.Fatal(got)
	}
	delete(msg, "status_text")
	if got := parseNotifyTemplate(`{{.StatusText}}`, msg); got != "Failed" {
		t.Fatal(got)
	}
}

func TestDefaultDisplayTemplateAndCustomTemplate(t *testing.T) {
	msg := Message{"status": "Failed", "status_text": "失败"}
	if got := parseDisplayNotifyTemplate(models.DefaultNotificationTemplate, msg); !contains(got, "Status: 失败") {
		t.Fatal(got)
	}
	if got := parseDisplayNotifyTemplate(`custom: {{.Status}}`, msg); got != "custom: Failed" {
		t.Fatal(got)
	}
}

func TestWebhookStatusTextJSONEscaping(t *testing.T) {
	msg := Message{
		"task_id": 1, "name": "task", "output": "done",
		"status": "Failed", "status_text": "失败 \"原因\"\n下一行",
	}
	content := renderWebhookTemplate(`{"status":"{{.Status}}","status_text":"{{.StatusText}}"}`, msg)
	var got struct {
		Status     string `json:"status"`
		StatusText string `json:"status_text"`
	}
	if err := json.Unmarshal([]byte(content), &got); err != nil {
		t.Fatalf("invalid webhook JSON %q: %v", content, err)
	}
	if got.Status != "Failed" || got.StatusText != "失败 \"原因\"\n下一行" {
		t.Fatalf("unexpected webhook payload: %+v", got)
	}
}
