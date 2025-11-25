package logger

import (
	"bytes"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"
)

// 高并发场景测试
func BenchmarkConcurrentSync(b *testing.B) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			handler.Handle(nil, slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0))
		}
	})
}

func BenchmarkConcurrentAsync(b *testing.B) {
	var buf bytes.Buffer
	handler := newAsyncHandler(&buf, 50, 100*time.Millisecond)
	defer handler.close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			handler.log(slog.LevelInfo, "test")
		}
	})
}

// 真实场景：模拟任务执行中的日志写入
func TestRealWorldScenario(t *testing.T) {
	tests := []struct {
		name        string
		taskNum     int
		logsPerTask int
	}{
		{"10任务x10日志", 10, 10},
		{"100任务x10日志", 100, 10},
		{"100任务x100日志", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name+"-同步", func(t *testing.T) {
			var buf bytes.Buffer
			handler := slog.NewTextHandler(&buf, nil)

			start := time.Now()
			var wg sync.WaitGroup
			for i := 0; i < tt.taskNum; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < tt.logsPerTask; j++ {
						handler.Handle(nil, slog.NewRecord(time.Now(), slog.LevelInfo, "task executing", 0))
					}
				}()
			}
			wg.Wait()
			elapsed := time.Since(start)
			t.Logf("同步日志耗时: %v", elapsed)
		})

		t.Run(tt.name+"-异步", func(t *testing.T) {
			var buf bytes.Buffer
			handler := newAsyncHandler(&buf, 50, 100*time.Millisecond)
			defer handler.close()

			start := time.Now()
			var wg sync.WaitGroup
			for i := 0; i < tt.taskNum; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < tt.logsPerTask; j++ {
						handler.log(slog.LevelInfo, "task executing")
					}
				}()
			}
			wg.Wait()
			elapsed := time.Since(start)
			t.Logf("异步日志耗时: %v", elapsed)
		})
	}
}

// 吞吐量测试
func TestThroughput(t *testing.T) {
	duration := 1 * time.Second

	t.Run("同步吞吐量", func(t *testing.T) {
		var buf bytes.Buffer
		handler := slog.NewTextHandler(&buf, nil)

		count := 0
		done := make(chan bool)

		go func() {
			time.Sleep(duration)
			done <- true
		}()

		for {
			select {
			case <-done:
				t.Logf("同步日志 1秒内写入: %d 条", count)
				return
			default:
				handler.Handle(nil, slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0))
				count++
			}
		}
	})

	t.Run("异步吞吐量", func(t *testing.T) {
		var buf bytes.Buffer
		handler := newAsyncHandler(&buf, 50, 100*time.Millisecond)
		defer handler.close()

		count := 0
		done := make(chan bool)

		go func() {
			time.Sleep(duration)
			done <- true
		}()

		for {
			select {
			case <-done:
				t.Logf("异步日志 1秒内写入: %d 条", count)
				return
			default:
				handler.log(slog.LevelInfo, "test")
				count++
			}
		}
	})
}

// 测试写入真实文件的性能差异
func BenchmarkRealFileSync(b *testing.B) {
	writer := io.Discard
	handler := slog.NewTextHandler(writer, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(nil, slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0))
	}
}

func BenchmarkRealFileAsync(b *testing.B) {
	writer := io.Discard
	handler := newAsyncHandler(writer, 50, 100*time.Millisecond)
	defer handler.close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.log(slog.LevelInfo, "test message")
	}
}
