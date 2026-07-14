package base

import "testing"

// TestParseStatusFilter 锁定状态过滤解析:空/非法 -> -1(不过滤),否则原样返回枚举值。
// 这是 issue #236 的回归测试:此前 handler 对状态值多减了 1,导致
// 任务列表(启用=1/禁用=0)与任务日志(失败=0/运行中=1/成功=2/取消=3)
// 的过滤结果与所选状态错位。
func TestParseStatusFilter(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", -1},    // 未选择 -> 不过滤
		{"   ", -1}, // 纯空白 -> 不过滤
		{"abc", -1}, // 非数字 -> 不过滤
		{"-1", -1},  // 负数 -> 不过滤
		{"-5", -1},  // 负数 -> 不过滤
		{"0", 0},    // 禁用 / 失败
		{"1", 1},    // 启用 / 运行中
		{"2", 2},    // 成功
		{"3", 3},    // 取消
		{" 2 ", 2},  // 容忍首尾空白
	}
	for _, tc := range cases {
		if got := ParseStatusFilter(tc.in); got != tc.want {
			t.Errorf("ParseStatusFilter(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
