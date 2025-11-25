package logger

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"time"
)

// 异步日志批处理器
type asyncHandler struct {
	handler   slog.Handler
	logChan   chan *logRecord
	batchSize int
	flushTime time.Duration
	wg        sync.WaitGroup
	once      sync.Once
}

type logRecord struct {
	record slog.Record
}

// 创建异步处理器
func newAsyncHandler(writer io.Writer, batchSize int, flushTime time.Duration) *asyncHandler {
	h := &asyncHandler{
		handler:   slog.NewTextHandler(writer, &slog.HandlerOptions{Level: slog.LevelDebug}),
		logChan:   make(chan *logRecord, batchSize*2),
		batchSize: batchSize,
		flushTime: flushTime,
	}
	h.wg.Add(1)
	go h.worker()
	return h
}

// 后台批量写入
func (h *asyncHandler) worker() {
	defer h.wg.Done()
	batch := make([]*logRecord, 0, h.batchSize)
	ticker := time.NewTicker(h.flushTime)
	defer ticker.Stop()
	ctx := context.Background()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		for _, rec := range batch {
			h.handler.Handle(ctx, rec.record)
		}
		batch = batch[:0]
	}

	for {
		select {
		case record, ok := <-h.logChan:
			if !ok {
				flush()
				return
			}
			batch = append(batch, record)
			if len(batch) >= h.batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

// 写入日志（非阻塞）
func (h *asyncHandler) log(level slog.Level, msg string, args ...any) {
	rec := slog.NewRecord(time.Now(), level, msg, 0)
	// 使用对象池减少分配
	logRec := &logRecord{record: rec}
	select {
	case h.logChan <- logRec:
	default:
		// 队列满时直接同步写入（降级策略）
		h.handler.Handle(context.Background(), rec)
	}
}

// 优雅关闭
func (h *asyncHandler) close() {
	h.once.Do(func() {
		close(h.logChan)
		h.wg.Wait()
	})
}
