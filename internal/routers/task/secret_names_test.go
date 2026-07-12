package task

import "testing"

func TestNormalizeSecretNames(t *testing.T) {
	tests := []struct {
		raw    string
		want   string
		wantOk bool
	}{
		// 空输入:合法,表示注入全部机密(兼容旧任务)
		{"", "", true},
		{"   ", "", true},
		{",, ,", "", true},
		// 常规
		{"API_KEY", "API_KEY", true},
		{"API_KEY,DB_PASSWORD", "API_KEY,DB_PASSWORD", true},
		// 去空白、去重
		{" API_KEY , DB_PASSWORD ,API_KEY,", "API_KEY,DB_PASSWORD", true},
		// 非法环境变量名
		{"1BAD", "", false},
		{"GOOD,has space", "", false},
		{"A-B", "", false},
		{"$(cmd)", "", false},
	}
	for _, tt := range tests {
		got, ok := normalizeSecretNames(tt.raw)
		if got != tt.want || ok != tt.wantOk {
			t.Errorf("normalizeSecretNames(%q) = (%q, %v), want (%q, %v)",
				tt.raw, got, ok, tt.want, tt.wantOk)
		}
	}
}
