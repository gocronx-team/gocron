package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestAsyncLoggerPerformance(t *testing.T) {
	var buf bytes.Buffer
	handler := newAsyncHandler(&buf, 50, 50*time.Millisecond)
	defer handler.close()

	// 写入1000条日志
	start := time.Now()
	for i := 0; i < 1000; i++ {
		handler.log(slog.LevelInfo, "test message")
	}
	handler.close()
	elapsed := time.Since(start)

	t.Logf("写入1000条日志耗时: %v", elapsed)

	// 验证日志已写入
	if buf.Len() == 0 {
		t.Fatal("日志未写入")
	}
}

func TestAsyncLoggerBatchFlush(t *testing.T) {
	var buf bytes.Buffer
	handler := newAsyncHandler(&buf, 10, 100*time.Millisecond)
	defer handler.close()

	// 写入5条日志（小于批量大小）
	for i := 0; i < 5; i++ {
		handler.log(slog.LevelInfo, "test")
	}

	// 等待定时刷新
	time.Sleep(150 * time.Millisecond)
	handler.close()

	// 验证日志已刷新
	if buf.Len() == 0 {
		t.Fatal("定时刷新失败")
	}
}

func TestAsyncLoggerFullBatch(t *testing.T) {
	var buf bytes.Buffer
	handler := newAsyncHandler(&buf, 10, 1*time.Second)
	defer handler.close()

	// 写入10条日志（等于批量大小）
	for i := 0; i < 10; i++ {
		handler.log(slog.LevelInfo, "test")
	}

	// 短暂等待批量写入
	time.Sleep(50 * time.Millisecond)

	// 验证日志已写入
	if buf.Len() == 0 {
		t.Fatal("批量写入失败")
	}
}

func TestAsyncLoggerClose(t *testing.T) {
	var buf bytes.Buffer
	handler := newAsyncHandler(&buf, 100, 1*time.Second)

	// 写入日志后立即关闭
	handler.log(slog.LevelInfo, "test message")
	handler.close()

	// 验证关闭时刷新了所有日志
	if !strings.Contains(buf.String(), "test message") {
		t.Fatal("关闭时未刷新日志")
	}
}

// 性能对比测试
func BenchmarkSyncLogger(b *testing.B) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0))
	}
}

func BenchmarkAsyncLogger(b *testing.B) {
	var buf bytes.Buffer
	handler := newAsyncHandler(&buf, 50, 100*time.Millisecond)
	defer handler.close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.log(slog.LevelInfo, "test")
	}
}
