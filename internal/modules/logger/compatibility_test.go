package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// 兼容性测试：确保API接口不变
func TestAPICompatibility(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	
	// 初始化logger
	var buf bytes.Buffer
	handler := newAsyncHandler(&buf, 10, 100*time.Millisecond)
	prevAsync := asyncLogWriter
	asyncLogWriter = handler
	t.Cleanup(func() {
		handler.close()
		asyncLogWriter = prevAsync
	})
	
	// 测试所有公开API是否正常工作
	t.Run("Info接口", func(t *testing.T) {
		Info("test")
		Infof("test %s", "format")
	})
	
	t.Run("Error接口", func(t *testing.T) {
		Error("test")
		Errorf("test %s", "format")
	})
	
	t.Run("Warn接口", func(t *testing.T) {
		Warn("test")
		Warnf("test %s", "format")
	})
	
	t.Run("Debug接口", func(t *testing.T) {
		Debug("test")
		Debugf("test %s", "format")
	})
}

// 测试日志输出格式不变
func TestLogFormatCompatibility(t *testing.T) {
	prevLogger := logger
	prevAsync := asyncLogWriter
	
	// 使用同步logger测试格式
	handler := newRecordingHandler()
	logger = slog.New(handler)
	asyncLogWriter = nil
	
	t.Cleanup(func() {
		logger = prevLogger
		asyncLogWriter = prevAsync
	})
	
	Info("test message")
	
	if len(handler.entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(handler.entries))
	}
	
	if handler.entries[0].msg != "test message" {
		t.Errorf("expected 'test message', got '%s'", handler.entries[0].msg)
	}
}

// 测试降级策略：异步失败时自动降级到同步
func TestFallbackToSync(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	
	prevLogger := logger
	prevAsync := asyncLogWriter
	
	logger = slog.New(handler)
	asyncLogWriter = nil // 模拟异步不可用
	
	t.Cleanup(func() {
		logger = prevLogger
		asyncLogWriter = prevAsync
	})
	
	Info("fallback test")
	
	if !strings.Contains(buf.String(), "fallback test") {
		t.Error("降级到同步日志失败")
	}
}

// 测试Close方法幂等性
func TestCloseIdempotent(t *testing.T) {
	var buf bytes.Buffer
	handler := newAsyncHandler(&buf, 10, 100*time.Millisecond)
	
	handler.log(slog.LevelInfo, "test")
	
	// 多次调用Close应该安全
	handler.close()
	handler.close()
	handler.close()
	
	// 验证日志已写入
	if buf.Len() == 0 {
		t.Error("日志未写入")
	}
}
