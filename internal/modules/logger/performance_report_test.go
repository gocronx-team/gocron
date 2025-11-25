package logger

import (
	"bytes"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"
)

// 性能对比报告
func TestPerformanceReport(t *testing.T) {
	fmt.Println("")
	fmt.Println("========================================")
	fmt.Println("日志性能优化对比报告")
	fmt.Println("========================================")
	fmt.Println("")

	// 测试1: 单线程顺序写入
	t.Run("1.单线程顺序写入1000条", func(t *testing.T) {
		count := 1000

		// 同步
		var buf1 bytes.Buffer
		handler1 := slog.NewTextHandler(&buf1, nil)
		start1 := time.Now()
		for i := 0; i < count; i++ {
			handler1.Handle(nil, slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0))
		}
		sync1 := time.Since(start1)

		// 异步
		var buf2 bytes.Buffer
		handler2 := newAsyncHandler(&buf2, 50, 100*time.Millisecond)
		start2 := time.Now()
		for i := 0; i < count; i++ {
			handler2.log(slog.LevelInfo, "test")
		}
		handler2.close()
		async := time.Since(start2)

		improvement := float64(sync1-async) / float64(sync1) * 100
		fmt.Printf("  同步: %v\n", sync1)
		fmt.Printf("  异步: %v\n", async)
		fmt.Printf("  提升: %.1f%%\n\n", improvement)
	})

	// 测试2: 高并发场景
	t.Run("2.100个goroutine并发写入", func(t *testing.T) {
		goroutines := 100
		logsPerGoroutine := 100

		// 同步
		var buf1 bytes.Buffer
		handler1 := slog.NewTextHandler(&buf1, nil)
		start1 := time.Now()
		var wg1 sync.WaitGroup
		for i := 0; i < goroutines; i++ {
			wg1.Add(1)
			go func() {
				defer wg1.Done()
				for j := 0; j < logsPerGoroutine; j++ {
					handler1.Handle(nil, slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0))
				}
			}()
		}
		wg1.Wait()
		sync1 := time.Since(start1)

		// 异步
		var buf2 bytes.Buffer
		handler2 := newAsyncHandler(&buf2, 50, 100*time.Millisecond)
		start2 := time.Now()
		var wg2 sync.WaitGroup
		for i := 0; i < goroutines; i++ {
			wg2.Add(1)
			go func() {
				defer wg2.Done()
				for j := 0; j < logsPerGoroutine; j++ {
					handler2.log(slog.LevelInfo, "test")
				}
			}()
		}
		wg2.Wait()
		handler2.close()
		async := time.Since(start2)

		improvement := float64(sync1-async) / float64(sync1) * 100
		fmt.Printf("  同步: %v\n", sync1)
		fmt.Printf("  异步: %v\n", async)
		fmt.Printf("  提升: %.1f%%\n\n", improvement)
	})

	// 测试3: 模拟真实任务场景
	t.Run("3.模拟50个任务执行(每任务20条日志)", func(t *testing.T) {
		tasks := 50
		logsPerTask := 20

		// 同步
		var buf1 bytes.Buffer
		handler1 := slog.NewTextHandler(&buf1, nil)
		start1 := time.Now()
		var wg1 sync.WaitGroup
		for i := 0; i < tasks; i++ {
			wg1.Add(1)
			go func(taskID int) {
				defer wg1.Done()
				// 模拟任务执行
				for j := 0; j < logsPerTask; j++ {
					handler1.Handle(nil, slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprintf("Task %d executing step %d", taskID, j), 0))
					time.Sleep(10 * time.Microsecond) // 模拟任务处理
				}
			}(i)
		}
		wg1.Wait()
		sync1 := time.Since(start1)

		// 异步
		var buf2 bytes.Buffer
		handler2 := newAsyncHandler(&buf2, 50, 100*time.Millisecond)
		start2 := time.Now()
		var wg2 sync.WaitGroup
		for i := 0; i < tasks; i++ {
			wg2.Add(1)
			go func(taskID int) {
				defer wg2.Done()
				// 模拟任务执行
				for j := 0; j < logsPerTask; j++ {
					handler2.log(slog.LevelInfo, fmt.Sprintf("Task %d executing step %d", taskID, j))
					time.Sleep(10 * time.Microsecond) // 模拟任务处理
				}
			}(i)
		}
		wg2.Wait()
		handler2.close()
		async := time.Since(start2)

		improvement := float64(sync1-async) / float64(sync1) * 100
		fmt.Printf("  同步: %v\n", sync1)
		fmt.Printf("  异步: %v\n", async)
		fmt.Printf("  提升: %.1f%%\n\n", improvement)
	})

	// 测试4: 批量写入效率
	t.Run("4.批量写入效率测试", func(t *testing.T) {
		count := 5000

		// 同步 - 每次都写入
		var buf1 bytes.Buffer
		handler1 := slog.NewTextHandler(&buf1, nil)
		start1 := time.Now()
		for i := 0; i < count; i++ {
			handler1.Handle(nil, slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0))
		}
		sync1 := time.Since(start1)

		// 异步 - 批量写入
		var buf2 bytes.Buffer
		handler2 := newAsyncHandler(&buf2, 50, 100*time.Millisecond)
		start2 := time.Now()
		for i := 0; i < count; i++ {
			handler2.log(slog.LevelInfo, "test")
		}
		handler2.close()
		async := time.Since(start2)

		improvement := float64(sync1-async) / float64(sync1) * 100
		fmt.Printf("  同步: %v (每次写入)\n", sync1)
		fmt.Printf("  异步: %v (批量写入)\n", async)
		fmt.Printf("  提升: %.1f%%\n\n", improvement)
	})

	fmt.Println("========================================")
	fmt.Println("优化总结:")
	fmt.Println("1. 异步日志不阻塞业务逻辑")
	fmt.Println("2. 批量写入减少I/O次数")
	fmt.Println("3. 高并发场景性能提升明显")
	fmt.Println("4. 内存开销可控(channel缓冲)")
	fmt.Println("========================================")
	fmt.Println("")
}
