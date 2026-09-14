package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/gocronx-team/cron"
)

// Cron 预览：解析表达式，返回接下来 N 次执行时间 + 未来 7 天的执行分布热图。
// 设计要点：
// - 复用后端 cron 库（和实际调度一致），避免前端重复实现造成语法漂移。
// - 迭代有硬上限，防 `* * * * * *` 这类极端表达式 DoS 服务。
// - 非法表达式不抛错，返回 Valid=false，让前端能平滑展示。

const (
	maxHeatmapIterations = 2000
	heatmapWindowHours   = 24 * 7
	maxNextRunsRequested = 20
	defaultNextRuns      = 10
	maxSpecLength        = 128
)

type CronRun struct {
	Unix    int64  `json:"unix"`
	ISO     string `json:"iso"`
	Weekday int    `json:"weekday"` // 0=Sun..6=Sat
}

type HeatmapCell struct {
	Day   int `json:"day"`
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

type CronPreviewResult struct {
	Valid        bool          `json:"valid"`
	Error        string        `json:"error,omitempty"`
	Startup      bool          `json:"startup,omitempty"` // @reboot：调度器启动时执行一次，无时间表
	Timezone     string        `json:"timezone"`
	NowUnix      int64         `json:"now_unix"`
	NextRuns     []CronRun     `json:"next_runs"`
	HeatmapCells []HeatmapCell `json:"heatmap_cells"`
	Truncated    bool          `json:"truncated,omitempty"` // heatmap 达到迭代上限提前截断
}

// PreviewCron 以当前时间为基准计算预览。
func PreviewCron(spec, timezone string, count int) *CronPreviewResult {
	return previewCronAt(spec, timezone, count, time.Now())
}

// previewCronAt 暴露 now 参数用于单测（时间可注入）。
func previewCronAt(spec, timezone string, count int, now time.Time) *CronPreviewResult {
	// 1. 输入清洗
	spec = strings.TrimSpace(spec)
	result := &CronPreviewResult{
		Timezone: timezone,
		NowUnix:  now.Unix(),
	}
	if spec == "" {
		result.Error = "spec is required"
		return result
	}
	if strings.ContainsAny(spec, "\r\n") {
		result.Error = "spec must be single-line"
		return result
	}
	if len(spec) > maxSpecLength {
		result.Error = fmt.Sprintf("spec too long (max %d chars)", maxSpecLength)
		return result
	}

	if count <= 0 {
		count = defaultNextRuns
	}
	if count > maxNextRunsRequested {
		count = maxNextRunsRequested
	}

	// 2. 处理时区：如果显式传了 timezone，去掉 spec 里可能的前缀后再用它包
	finalSpec, effectiveTZ, tzErr := resolveSpecTimezone(spec, timezone)
	if tzErr != "" {
		result.Error = tzErr
		return result
	}
	result.Timezone = effectiveTZ

	// 3. 解析
	schedule, err := cron.ParseWithError(finalSpec)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Valid = true

	// @reboot 在调度器启动时执行一次，没有可预览的时间表。用裸 spec 判断以兼容
	// 时区前缀（CRON_TZ=... @reboot 会被解析库包装成 TZSchedule）。
	if bare, _ := stripTimezonePrefix(finalSpec); bare == "@reboot" {
		result.Startup = true
		return result
	}

	// 4. 接下来 N 次
	t := now
	for i := 0; i < count; i++ {
		next := schedule.Next(t)
		// 防死循环：Next 返回零值或不推进说明永不触发
		if next.IsZero() || !next.After(t) {
			break
		}
		result.NextRuns = append(result.NextRuns, CronRun{
			Unix:    next.Unix(),
			ISO:     next.Format(time.RFC3339),
			Weekday: int(next.Weekday()),
		})
		t = next
	}

	// 5. 未来 7 天分布（稀疏格式，只返 count>0 的格子）
	heatmapEnd := now.Add(time.Duration(heatmapWindowHours) * time.Hour)
	cells := make(map[[2]int]int)
	t = now
	for iter := 0; iter < maxHeatmapIterations; iter++ {
		next := schedule.Next(t)
		if next.IsZero() || !next.After(t) || next.After(heatmapEnd) {
			break
		}
		key := [2]int{int(next.Weekday()), next.Hour()}
		cells[key]++
		t = next
		if iter == maxHeatmapIterations-1 {
			// 还有一次就触顶，再看一眼是否真到窗口末
			if peek := schedule.Next(t); !peek.IsZero() && peek.After(t) && peek.Before(heatmapEnd) {
				result.Truncated = true
			}
		}
	}
	for key, c := range cells {
		result.HeatmapCells = append(result.HeatmapCells, HeatmapCell{
			Day: key[0], Hour: key[1], Count: c,
		})
	}

	return result
}

// resolveSpecTimezone 处理 spec 和 timezone 的组合：
// - 显式 timezone != "" : 剥除 spec 里已有的 CRON_TZ=/TZ= 前缀后，用显式 timezone 重新包
// - 显式 timezone == "" 且 spec 带前缀：保留原样，effectiveTZ 返回前缀里的 tz
// - 都没有：用服务器本地时区
func resolveSpecTimezone(spec, timezone string) (finalSpec, effectiveTZ, errMsg string) {
	bareSpec, prefixTZ := stripTimezonePrefix(spec)

	if timezone != "" {
		if _, err := time.LoadLocation(timezone); err != nil {
			return "", timezone, fmt.Sprintf("unknown timezone: %q", timezone)
		}
		return "CRON_TZ=" + timezone + " " + bareSpec, timezone, ""
	}

	if prefixTZ != "" {
		// spec 自带前缀，cron 库自己能解
		if _, err := time.LoadLocation(prefixTZ); err != nil {
			return "", prefixTZ, fmt.Sprintf("unknown timezone: %q", prefixTZ)
		}
		return spec, prefixTZ, ""
	}

	// 都没指定，用服务器本地
	return bareSpec, time.Local.String(), ""
}

// stripTimezonePrefix 返回 (去除前缀的 spec, 前缀中的 timezone)
// 无前缀时 timezone 为空。
func stripTimezonePrefix(spec string) (bareSpec, timezone string) {
	if !strings.HasPrefix(spec, "CRON_TZ=") && !strings.HasPrefix(spec, "TZ=") {
		return spec, ""
	}
	eqIdx := strings.IndexByte(spec, '=')
	rest := spec[eqIdx+1:]
	spIdx := strings.IndexByte(rest, ' ')
	if spIdx < 0 {
		return spec, ""
	}
	return strings.TrimSpace(rest[spIdx+1:]), rest[:spIdx]
}
