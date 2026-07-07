package service

import (
	"errors"
	"testing"

	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/notify"
)

func TestMatchNotifyKeyword(t *testing.T) {
	tests := []struct {
		name    string
		keyword string
		regex   int8
		output  string
		want    bool
	}{
		{"substring hit", "ERROR", 0, "some ERROR here", true},
		{"substring miss", "ERROR", 0, "all good", false},
		{"empty keyword", "", 0, "anything", false},
		{"regex hit", "ERR[0-9]+", 1, "ERR503 occurred", true},
		{"regex miss", "ERR[0-9]+", 1, "ERR occurred", false},
		{"invalid regex is safe", "ERR[0-9", 1, "ERR503", false},
		// 大小写:正则默认区分大小写
		{"regex case-sensitive miss", "error", 1, "ERROR happened", false},
		{"regex case-insensitive flag", "(?i)error", 1, "ERROR happened", true},
		// 或、锚点
		{"regex alternation", "timeout|refused", 1, "connection refused", true},
		{"regex anchor hit", "^FAIL", 1, "FAIL: boom", true},
		{"regex anchor miss single-line", "^FAIL", 1, "job FAIL", false},
		// 多行输出:MatchString 在整段任意位置匹配
		{"regex matches within multiline", "panic", 1, "line1\nfatal panic\nline3", true},
		// 默认 ^ 只锚定整串开头(非每行);多行需显式 (?m)
		{"regex anchor without multiline flag", "^panic", 1, "line1\npanic here", false},
		{"regex multiline flag anchors each line", "(?m)^panic", 1, "line1\npanic here", true},
		// 默认 . 不跨行;需 (?s) 才跨行
		{"regex dot no newline by default", "start.end", 1, "start\nend", false},
		{"regex dotall flag crosses newline", "(?s)start.end", 1, "start\nend", true},
		// 中文关键字
		{"regex chinese keyword", "错误|失败", 1, "任务执行失败", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := models.Task{NotifyKeyword: tt.keyword, NotifyKeywordRegex: tt.regex}
			if got := matchNotifyKeyword(task, tt.output); got != tt.want {
				t.Errorf("matchNotifyKeyword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendNotificationBitmask(t *testing.T) {
	orig := notifyPushFunc
	defer func() { notifyPushFunc = orig }()
	var sent bool
	notifyPushFunc = func(_ notify.Message) { sent = true }

	// NotifyType=2 (WebHook) 不需要 receiver,聚焦触发条件
	newTask := func(ns int8) models.Task {
		return models.Task{NotifyStatus: ns, NotifyType: 2, NotifyKeyword: "BOOM"}
	}
	failRes := TaskResult{Err: errors.New("fail"), Result: "output BOOM here"}
	okKw := TaskResult{Err: nil, Result: "output BOOM here"}
	okNoKw := TaskResult{Err: nil, Result: "clean output"}
	failNoKw := TaskResult{Err: errors.New("fail"), Result: "clean output"}

	cases := []struct {
		name string
		ns   int8
		res  TaskResult
		want bool
	}{
		{"disabled", 0, failRes, false},
		{"failure-bit + failed", 1, failNoKw, true},
		{"failure-bit + success", 1, okNoKw, false},
		{"success-bit + success", 2, okNoKw, true},
		{"success-bit + failed", 2, failNoKw, false},
		{"fail|success + failed", 3, failNoKw, true},
		{"fail|success + success", 3, okNoKw, true},
		{"keyword-bit + match", 4, okKw, true},
		{"keyword-bit + no match", 4, okNoKw, false},
		{"fail|keyword + success but keyword matches", 5, okKw, true},
		{"fail|keyword + success no keyword", 5, okNoKw, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			sent = false
			SendNotification(newTask(tt.ns), tt.res)
			if sent != tt.want {
				t.Errorf("ns=%d: sent=%v, want=%v", tt.ns, sent, tt.want)
			}
		})
	}
}

func TestSendNotificationRequiresReceiver(t *testing.T) {
	orig := notifyPushFunc
	defer func() { notifyPushFunc = orig }()
	sent := false
	notifyPushFunc = func(_ notify.Message) { sent = true }

	// NotifyType=0(邮件)且无 receiver → 即使触发条件满足也不发
	task := models.Task{NotifyStatus: 1, NotifyType: 0, NotifyReceiverId: ""}
	SendNotification(task, TaskResult{Err: errors.New("x")})
	if sent {
		t.Error("should not send when receiver is missing for non-webhook channel")
	}
}
