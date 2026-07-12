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

func TestMatchNotifyKeywordExclude(t *testing.T) {
	tests := []struct {
		name    string
		keyword string
		exclude string
		regex   int8
		output  string
		want    bool
	}{
		// 子串模式
		{"substring: exclude empty keeps match", "ERROR", "", 0, "an ERROR here", true},
		{"substring: exclude hit suppresses", "ERROR", "ERROR: ignored", 0, "ERROR: ignored by config", false},
		{"substring: exclude miss keeps match", "ERROR", "ERROR: ignored", 0, "fatal ERROR occurred", true},
		{"substring: exclude without keyword hit", "ERROR", "ERROR: ignored", 0, "all good", false},
		// 正则模式(排除与关键字共用正则开关)
		{"regex: exclude hit suppresses", "ERROR", "ERROR: (ignored|known)", 1, "ERROR: known issue", false},
		{"regex: exclude miss keeps match", "ERROR", "ERROR: (ignored|known)", 1, "ERROR: disk full", true},
		{"regex: lookaround replacement case", "ERROR", "ERROR: ignored", 1, "step1 ok\nERROR: ignored\nstep2 ok", false},
		{"regex: exclude anchors", "FAIL", "(?m)^FAIL: expected$", 1, "FAIL: expected", false},
		{"regex: exclude case-insensitive", "ERROR", "(?i)error: ignored", 1, "ERROR: IGNORED", false},
		// 非法排除正则:忽略排除、照常通知(fail-open,不能静默漏发告警)
		{"regex: invalid exclude fails open", "ERROR", "ERROR: [unclosed", 1, "an ERROR here", true},
		// 排除仅在关键字命中后才有意义
		{"regex: keyword miss ignores exclude", "ERROR", "whatever", 1, "all good", false},
		{"empty keyword never matches", "", "ERROR", 0, "ERROR here", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := models.Task{
				NotifyKeyword:        tt.keyword,
				NotifyKeywordExclude: tt.exclude,
				NotifyKeywordRegex:   tt.regex,
			}
			if got := matchNotifyKeyword(task, tt.output); got != tt.want {
				t.Errorf("matchNotifyKeyword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchOutputPattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		useRegex bool
		output   string
		want     bool
		wantErr  bool
	}{
		{"substring hit", "boom", false, "kaboom", true, false},
		{"substring miss", "boom", false, "ok", false, false},
		{"regex hit", "b.om", true, "kaboom", true, false},
		{"regex invalid returns error", "b[om", true, "kaboom", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := matchOutputPattern(tt.pattern, tt.useRegex, tt.output)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("matchOutputPattern() = %v, want %v", got, tt.want)
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

func TestSendNotificationKeywordExclude(t *testing.T) {
	orig := notifyPushFunc
	defer func() { notifyPushFunc = orig }()
	var sent bool
	notifyPushFunc = func(_ notify.Message) { sent = true }

	newTask := func(ns int8) models.Task {
		return models.Task{
			NotifyStatus:         ns,
			NotifyType:           2,
			NotifyKeyword:        "BOOM",
			NotifyKeywordExclude: "BOOM (ignored)",
		}
	}
	okKwExcluded := TaskResult{Err: nil, Result: "output BOOM (ignored) here"}
	okKwOnly := TaskResult{Err: nil, Result: "output BOOM here"}
	failKwExcluded := TaskResult{Err: errors.New("fail"), Result: "output BOOM (ignored) here"}

	cases := []struct {
		name string
		ns   int8
		res  TaskResult
		want bool
	}{
		// 排除命中时抑制关键字条件
		{"keyword-bit + excluded", 4, okKwExcluded, false},
		{"keyword-bit + not excluded", 4, okKwOnly, true},
		// 排除只作用于关键字条件,不影响失败/成功条件
		{"fail|keyword + failed + excluded still notifies via failure bit", 5, failKwExcluded, true},
		{"success|keyword + success + excluded still notifies via success bit", 6, okKwExcluded, true},
		{"keyword-only + failed + excluded stays silent", 4, failKwExcluded, false},
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
