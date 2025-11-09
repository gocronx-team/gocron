package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/Tang-RoseChild/mahonia"
)

func TestRandAuthToken(t *testing.T) {
	token := RandAuthToken()
	if len(token) != 64 {
		t.Fatalf("expected length 64, got %d", len(token))
	}
	if matched := regexp.MustCompile(`^[0-9a-f]+$`).MatchString(token); !matched {
		t.Fatalf("token should be hex, got %s", token)
	}
}

func TestRandString(t *testing.T) {
	tests := []struct {
		name   string
		length int64
	}{
		{"zero", 0},
		{"positive", 16},
	}
	charset := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandString(tt.length)
			if int64(len(got)) != tt.length {
				t.Fatalf("expected length %d, got %d", tt.length, len(got))
			}
			for _, c := range got {
				if !strings.ContainsRune(charset, c) {
					t.Fatalf("unexpected rune %q in result %q", c, got)
				}
			}
		})
	}
}

func TestMd5(t *testing.T) {
	got := Md5("gocron")
	const expect = "9a34de944ae472434f79c0eb612ca724"
	if got != expect {
		t.Fatalf("expected %s, got %s", expect, got)
	}
}

func TestRandNumber(t *testing.T) {
	const max = 10
	for i := 0; i < 100; i++ {
		n := RandNumber(max)
		if n < 0 || n >= max {
			t.Fatalf("number out of range: %d", n)
		}
	}
}

func TestGBK2UTF8(t *testing.T) {
	encoder := mahonia.NewEncoder("gbk")
	gbkStr := encoder.ConvertString("你好")
	utf8Str, ok := GBK2UTF8(gbkStr)
	if !ok {
		t.Fatal("expected conversion success")
	}
	if utf8Str != "你好" {
		t.Fatalf("expected 你好, got %s", utf8Str)
	}
}

func TestReplaceStrings(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		if got := ReplaceStrings("", []string{"a"}, []string{"b"}); got != "" {
			t.Fatalf("expected empty string, got %s", got)
		}
	})
	t.Run("length mismatch", func(t *testing.T) {
		original := "foo"
		if got := ReplaceStrings(original, []string{"f"}, []string{"b", "c"}); got != original {
			t.Fatalf("expected original string, got %s", got)
		}
	})
	t.Run("replace success", func(t *testing.T) {
		input := "a\nb\tc\""
		got := ReplaceStrings(input, []string{"\n", "\t", "\""}, []string{"N", "T", "Q"})
		if got != "aNbTcQ" {
			t.Fatalf("unexpected replace result %s", got)
		}
	})
}

func TestInStringSlice(t *testing.T) {
	if !InStringSlice([]string{" foo ", "bar"}, "foo") {
		t.Fatal("expected to find trimmed element")
	}
	if InStringSlice([]string{"foo"}, "bar") {
		t.Fatal("did not expect to find missing element")
	}
}

func TestEscapeJson(t *testing.T) {
	input := "line1\n\"quote\"\t\\slash"
	got := EscapeJson(input)
	expect := "line1\\n\\\"quote\\\"\\t\\\\slash"
	if got != expect {
		t.Fatalf("expected %s, got %s", expect, got)
	}
}

func TestFileExist(t *testing.T) {
	tempDir := t.TempDir()
	file := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if !FileExist(file) {
		t.Fatal("expected file to exist")
	}
	if FileExist(filepath.Join(tempDir, "missing.txt")) {
		t.Fatal("expected missing file to return false")
	}
}

func TestFormatAppVersion(t *testing.T) {
	info, err := FormatAppVersion("1.2.3", "abcdef", "2024-01-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, expect := range []string{"1.2.3", "abcdef", "2024-01-01", runtime.Version(), runtime.GOOS + "/" + runtime.GOARCH} {
		if !strings.Contains(info, expect) {
			t.Fatalf("expected output to contain %s, got %s", expect, info)
		}
	}
}

func TestPanicToError(t *testing.T) {
	t.Run("panic captured", func(t *testing.T) {
		err := PanicToError(func() {
			panic("boom")
		})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "boom") {
			t.Fatalf("expected error to contain panic message, got %v", err)
		}
	})
	t.Run("no panic", func(t *testing.T) {
		if err := PanicToError(func() {}); err != nil {
			t.Fatalf("did not expect error, got %v", err)
		}
	})
}

func TestPanicTrace(t *testing.T) {
	trace := PanicTrace("boom")
	if !strings.Contains(trace, "panic:") || !strings.Contains(trace, "boom") {
		t.Fatalf("unexpected panic trace: %s", trace)
	}
}
