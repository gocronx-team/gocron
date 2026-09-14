package service

import (
	"strings"
	"testing"
	"time"
)

// 固定 now 让测试完全确定：2026-04-20 周一 12:00 UTC
var testNow = time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)

func TestPreviewCron_ValidStandardExpressions(t *testing.T) {
	tests := []struct {
		name       string
		spec       string
		timezone   string
		wantNextN  int // 预期返回的 next_runs 条数
		firstRunOK func(time.Time) bool
	}{
		{
			name:      "每分钟",
			spec:      "0 * * * * *",
			wantNextN: 10,
			firstRunOK: func(tm time.Time) bool {
				// 下次执行应该在 now 的下一分钟
				return tm.Second() == 0 && tm.After(testNow) && tm.Sub(testNow) <= time.Minute
			},
		},
		{
			name:      "周一至周五 9:30",
			spec:      "0 30 9 * * 1-5",
			wantNextN: 10,
			firstRunOK: func(tm time.Time) bool {
				wd := tm.Weekday()
				return tm.Hour() == 9 && tm.Minute() == 30 && wd >= time.Monday && wd <= time.Friday
			},
		},
		{
			name:      "@daily 快捷",
			spec:      "@daily",
			wantNextN: 10,
			firstRunOK: func(tm time.Time) bool {
				return tm.Hour() == 0 && tm.Minute() == 0 && tm.Second() == 0
			},
		},
		{
			name:      "@every 30s",
			spec:      "@every 30s",
			wantNextN: 10,
			firstRunOK: func(tm time.Time) bool {
				return tm.Sub(testNow) <= 30*time.Second && tm.After(testNow)
			},
		},
		{
			name:      "5 段格式（省略 dow，自动补 *）",
			spec:      "0 30 9 * *",
			wantNextN: 10,
			firstRunOK: func(tm time.Time) bool {
				// second=0 minute=30 hour=9 day=* month=* dow=*（自动补）
				return tm.Hour() == 9 && tm.Minute() == 30 && tm.Second() == 0
			},
		},
		{
			name:      "自定义 count",
			spec:      "0 * * * * *",
			wantNextN: 10, // count=0 走默认 10
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := previewCronAt(tc.spec, tc.timezone, 0, testNow)
			if !got.Valid {
				t.Fatalf("expected valid=true, got error=%q", got.Error)
			}
			if len(got.NextRuns) != tc.wantNextN {
				t.Errorf("next_runs length=%d want=%d", len(got.NextRuns), tc.wantNextN)
			}
			// 单调递增
			for i := 1; i < len(got.NextRuns); i++ {
				if got.NextRuns[i].Unix <= got.NextRuns[i-1].Unix {
					t.Errorf("next_runs not strictly increasing at index %d", i)
				}
			}
			if tc.firstRunOK != nil && len(got.NextRuns) > 0 {
				first := time.Unix(got.NextRuns[0].Unix, 0).UTC()
				if !tc.firstRunOK(first) {
					t.Errorf("first run %v failed validation", first)
				}
			}
		})
	}
}

func TestPreviewCron_InvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		timezone string
		wantErr  string // 子串匹配
	}{
		{"空字符串", "", "", "required"},
		{"只有空格", "   ", "", "required"},
		{"换行注入", "0 * * * *\n*", "", "single-line"},
		{"回车注入", "0 * * * *\r*", "", "single-line"},
		{"超长", strings.Repeat("0 ", 80), "", "too long"},
		{"字段不够", "0 0", "", "expected 5 or 6 fields"},
		{"字段过多", "0 0 0 0 0 0 0 0", "", "expected 5 or 6 fields"},
		{"非法 @ 快捷", "@nonexistent", "", "unrecognized"},
		{"非法时区", "0 * * * * *", "Mars/Phobos", "unknown timezone"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := previewCronAt(tc.spec, tc.timezone, 0, testNow)
			if got.Valid {
				t.Fatalf("expected valid=false for %q", tc.spec)
			}
			if !strings.Contains(strings.ToLower(got.Error), strings.ToLower(tc.wantErr)) {
				t.Errorf("error %q does not contain %q", got.Error, tc.wantErr)
			}
		})
	}
}

func TestPreviewCron_Timezone(t *testing.T) {
	t.Run("显式 timezone 生效", func(t *testing.T) {
		// Asia/Shanghai 09:30 = UTC 01:30；now=UTC 12:00
		// 下次 SH 09:30 应该是次日 UTC 01:30
		got := previewCronAt("0 30 9 * * *", "Asia/Shanghai", 3, testNow)
		if !got.Valid {
			t.Fatalf("expected valid, got %q", got.Error)
		}
		if got.Timezone != "Asia/Shanghai" {
			t.Errorf("timezone=%q want=Asia/Shanghai", got.Timezone)
		}
		if len(got.NextRuns) == 0 {
			t.Fatal("no next runs")
		}
		// 验证首次执行换算到 UTC 是 01:30
		first := time.Unix(got.NextRuns[0].Unix, 0).UTC()
		if first.Hour() != 1 || first.Minute() != 30 {
			t.Errorf("first run in UTC = %v, expected hour=1 min=30", first)
		}
	})

	t.Run("spec 自带 CRON_TZ 前缀", func(t *testing.T) {
		got := previewCronAt("CRON_TZ=America/New_York 0 30 9 * * *", "", 3, testNow)
		if !got.Valid {
			t.Fatalf("expected valid, got %q", got.Error)
		}
		if got.Timezone != "America/New_York" {
			t.Errorf("timezone=%q want=America/New_York", got.Timezone)
		}
	})

	t.Run("显式 timezone 覆盖 spec 自带前缀", func(t *testing.T) {
		// spec 里是 NY，参数传 SH，应该以参数为准
		got := previewCronAt("CRON_TZ=America/New_York 0 30 9 * * *", "Asia/Shanghai", 3, testNow)
		if !got.Valid {
			t.Fatalf("expected valid, got %q", got.Error)
		}
		if got.Timezone != "Asia/Shanghai" {
			t.Errorf("timezone=%q want=Asia/Shanghai (explicit should override)", got.Timezone)
		}
	})
}

func TestPreviewCron_Reboot(t *testing.T) {
	t.Run("@reboot 标记为启动触发且不产生虚假执行计划", func(t *testing.T) {
		got := previewCronAt("@reboot", "", 10, testNow)
		if !got.Valid {
			t.Fatalf("expected valid, got %q", got.Error)
		}
		if !got.Startup {
			t.Error("expected Startup=true for @reboot")
		}
		if len(got.NextRuns) != 0 {
			t.Errorf("next_runs=%d want=0 for @reboot", len(got.NextRuns))
		}
		if len(got.HeatmapCells) != 0 {
			t.Errorf("heatmap_cells=%d want=0 for @reboot", len(got.HeatmapCells))
		}
	})

	t.Run("带时区前缀的 @reboot", func(t *testing.T) {
		got := previewCronAt("@reboot", "Asia/Tokyo", 10, testNow)
		if !got.Valid {
			t.Fatalf("expected valid, got %q", got.Error)
		}
		if !got.Startup {
			t.Error("expected Startup=true for @reboot with timezone")
		}
		if len(got.NextRuns) != 0 {
			t.Errorf("next_runs=%d want=0 for @reboot", len(got.NextRuns))
		}
	})
}

func TestPreviewCron_Heatmap(t *testing.T) {
	t.Run("标准低频表达式", func(t *testing.T) {
		// 每天 09:30 一次，一周 7 次
		got := previewCronAt("0 30 9 * * *", "", 0, testNow)
		if !got.Valid {
			t.Fatalf("expected valid, got %q", got.Error)
		}
		// 应该只有一个 cell：9 点那列，每天都有
		total := 0
		for _, c := range got.HeatmapCells {
			total += c.Count
			if c.Hour != 9 {
				t.Errorf("unexpected cell hour=%d (expected only 9)", c.Hour)
			}
		}
		if total != 7 {
			t.Errorf("weekly total=%d want=7", total)
		}
		if got.Truncated {
			t.Errorf("should not be truncated for simple daily schedule")
		}
	})

	t.Run("高频表达式触发迭代上限", func(t *testing.T) {
		// 每秒一次，一周 604800 次，会远超 maxHeatmapIterations=2000
		got := previewCronAt("* * * * * *", "", 0, testNow)
		if !got.Valid {
			t.Fatalf("expected valid, got %q", got.Error)
		}
		if !got.Truncated {
			t.Errorf("expected Truncated=true for * * * * * *")
		}
		// cell 数应 <= 24*7=168
		if len(got.HeatmapCells) > 168 {
			t.Errorf("cells=%d should be <=168", len(got.HeatmapCells))
		}
	})

	t.Run("极低频（7 天内无触发）", func(t *testing.T) {
		// 每年 1 月 1 日 00:00，从 2026-04-20 算 7 天内肯定不到
		got := previewCronAt("0 0 0 1 1 *", "", 0, testNow)
		if !got.Valid {
			t.Fatalf("expected valid, got %q", got.Error)
		}
		if len(got.HeatmapCells) != 0 {
			t.Errorf("cells=%d want=0 for yearly schedule", len(got.HeatmapCells))
		}
	})
}

func TestPreviewCron_CountClamp(t *testing.T) {
	t.Run("count<=0 用默认", func(t *testing.T) {
		got := previewCronAt("0 * * * * *", "", 0, testNow)
		if len(got.NextRuns) != defaultNextRuns {
			t.Errorf("count=%d want=%d", len(got.NextRuns), defaultNextRuns)
		}
	})
	t.Run("count 超上限被钳制", func(t *testing.T) {
		got := previewCronAt("0 * * * * *", "", 1000, testNow)
		if len(got.NextRuns) != maxNextRunsRequested {
			t.Errorf("count=%d want=%d", len(got.NextRuns), maxNextRunsRequested)
		}
	})
	t.Run("count=5", func(t *testing.T) {
		got := previewCronAt("0 * * * * *", "", 5, testNow)
		if len(got.NextRuns) != 5 {
			t.Errorf("count=%d want=5", len(got.NextRuns))
		}
	})
}

func TestStripTimezonePrefix(t *testing.T) {
	tests := []struct {
		in       string
		wantBare string
		wantTZ   string
	}{
		{"0 * * * * *", "0 * * * * *", ""},
		{"CRON_TZ=UTC 0 * * * * *", "0 * * * * *", "UTC"},
		{"TZ=Asia/Shanghai 0 30 9 * * *", "0 30 9 * * *", "Asia/Shanghai"},
		{"CRON_TZ=UTC", "CRON_TZ=UTC", ""}, // 没空格不算合法前缀
	}
	for _, tc := range tests {
		gotBare, gotTZ := stripTimezonePrefix(tc.in)
		if gotBare != tc.wantBare || gotTZ != tc.wantTZ {
			t.Errorf("stripTimezonePrefix(%q) = (%q, %q), want (%q, %q)",
				tc.in, gotBare, gotTZ, tc.wantBare, tc.wantTZ)
		}
	}
}

// Benchmark：`* * * * * *` 应在毫秒级完成，防 DoS
func BenchmarkPreviewCron_HighFrequency(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = previewCronAt("* * * * * *", "", 10, testNow)
	}
}

func BenchmarkPreviewCron_Standard(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = previewCronAt("0 30 9 * * 1-5", "Asia/Shanghai", 10, testNow)
	}
}
