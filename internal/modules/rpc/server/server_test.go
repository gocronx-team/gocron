package server

import "testing"

func TestEnvSlice(t *testing.T) {
	if got := envSlice(nil); got != nil {
		t.Errorf("expected nil for empty map, got %v", got)
	}

	got := envSlice(map[string]string{"API_KEY": "v1"})
	if len(got) != 1 || got[0] != "API_KEY=v1" {
		t.Errorf("unexpected single-entry result: %v", got)
	}

	// 多元素:顺序无关,逐项验证 KEY=VALUE 格式
	multi := envSlice(map[string]string{"A": "1", "B": "2"})
	if len(multi) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(multi))
	}
	set := map[string]bool{multi[0]: true, multi[1]: true}
	if !set["A=1"] || !set["B=2"] {
		t.Errorf("unexpected pairs: %v", multi)
	}
}
