package service

import (
	"strings"
	"testing"
	"time"

	"github.com/gocronx-team/gocron/internal/modules/diagnosis"
)

func TestFormatDiagnosisBlock(t *testing.T) {
	tests := []struct {
		name      string
		in        diagnosis.Result
		english   bool
		wantSubs  []string // 期望包含的子串
		wantEmpty bool
	}{
		{
			name:     "zh with cause and suggestions",
			in:       diagnosis.Result{RootCause: "连接超时", Suggestions: []string{"检查网络", "增大超时"}},
			english:  false,
			wantSubs: []string{"【AI 诊断】", "根因: 连接超时", "建议:", "- 检查网络", "- 增大超时"},
		},
		{
			name:     "english",
			in:       diagnosis.Result{RootCause: "timeout", Suggestions: []string{"check network"}},
			english:  true,
			wantSubs: []string{"[AI Diagnosis]", "Root cause: timeout", "Suggestions:", "- check network"},
		},
		{
			name:     "cause only, no suggestions",
			in:       diagnosis.Result{RootCause: "boom"},
			english:  false,
			wantSubs: []string{"根因: boom"},
		},
		{
			name:      "empty result yields empty block",
			in:        diagnosis.Result{},
			english:   false,
			wantEmpty: true,
		},
		{
			name:     "blank suggestions filtered out",
			in:       diagnosis.Result{RootCause: "x", Suggestions: []string{"  ", ""}},
			english:  false,
			wantSubs: []string{"根因: x"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDiagnosisBlock(tt.in, tt.english)
			if tt.wantEmpty {
				if got != "" {
					t.Errorf("expected empty, got %q", got)
				}
				return
			}
			for _, sub := range tt.wantSubs {
				if !strings.Contains(got, sub) {
					t.Errorf("block missing %q; got:\n%s", sub, got)
				}
			}
			// 空建议不应产生游离的 "- " 行
			if strings.Contains(got, "-  \n") || strings.HasSuffix(got, "- ") {
				t.Errorf("blank suggestion leaked into block:\n%s", got)
			}
		})
	}
}

func TestCooldownTrackerAllow(t *testing.T) {
	c := &cooldownTracker{last: make(map[int]time.Time)}
	base := time.Unix(1_700_000_000, 0)
	window := 10 * time.Minute

	if !c.allow(1, base, window) {
		t.Fatal("first call should be allowed")
	}
	// 窗口内:拒绝
	if c.allow(1, base.Add(5*time.Minute), window) {
		t.Error("within cooldown window should be denied")
	}
	// 超过窗口:放行
	if !c.allow(1, base.Add(11*time.Minute), window) {
		t.Error("after cooldown window should be allowed")
	}
	// 不同任务互不影响
	if !c.allow(2, base.Add(5*time.Minute), window) {
		t.Error("different task id should have independent cooldown")
	}
}
