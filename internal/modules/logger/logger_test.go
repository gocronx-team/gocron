package logger

import (
	"strings"
	"testing"

	"github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
)

type logEntry struct {
	level string
	args  []interface{}
}

type recordingLogger struct {
	seelog.LoggerInterface
	entries    []logEntry
	flushCount int
}

func newRecordingLogger() *recordingLogger {
	return &recordingLogger{LoggerInterface: seelog.Disabled}
}

func (r *recordingLogger) record(level string, args ...interface{}) {
	cp := make([]interface{}, len(args))
	copy(cp, args)
	r.entries = append(r.entries, logEntry{level: level, args: cp})
}

func (r *recordingLogger) Debug(v ...interface{}) {
	r.record("debug", v...)
}

func (r *recordingLogger) Debugf(format string, params ...interface{}) {
	r.record("debugf", append([]interface{}{format}, params...)...)
}

func (r *recordingLogger) Info(v ...interface{}) {
	r.record("info", v...)
}

func (r *recordingLogger) Infof(format string, params ...interface{}) {
	r.record("infof", append([]interface{}{format}, params...)...)
}

func (r *recordingLogger) Warn(v ...interface{}) error {
	r.record("warn", v...)
	return nil
}

func (r *recordingLogger) Warnf(format string, params ...interface{}) error {
	r.record("warnf", append([]interface{}{format}, params...)...)
	return nil
}

func (r *recordingLogger) Error(v ...interface{}) error {
	r.record("error", v...)
	return nil
}

func (r *recordingLogger) Errorf(format string, params ...interface{}) error {
	r.record("errorf", append([]interface{}{format}, params...)...)
	return nil
}

func (r *recordingLogger) Critical(v ...interface{}) error {
	r.record("critical", v...)
	return nil
}

func (r *recordingLogger) Criticalf(format string, params ...interface{}) error {
	r.record("criticalf", append([]interface{}{format}, params...)...)
	return nil
}

func (r *recordingLogger) Flush() {
	r.flushCount++
}

func setupRecordingLogger(t *testing.T) *recordingLogger {
	t.Helper()
	prevLogger := logger
	rec := newRecordingLogger()
	logger = rec
	t.Cleanup(func() { logger = prevLogger })
	return rec
}

func TestDebugLoggingDependsOnGinMode(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	rec := setupRecordingLogger(t)
	Debug("release-mode")
	if hasLevel(rec.entries, "debug") {
		t.Fatalf("expected no debug entry in release mode, got %+v", rec.entries)
	}

	gin.SetMode(gin.DebugMode)
	rec = setupRecordingLogger(t)
	Debug("debug-mode")
	if !hasLevel(rec.entries, "debug") {
		t.Fatalf("expected debug entry in debug mode, got %+v", rec.entries)
	}
}

func TestInfoLogsAndFlushes(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	rec := setupRecordingLogger(t)
	Info("info-message")
	if !hasLevel(rec.entries, "info") {
		t.Fatalf("expected info entry, got %+v", rec.entries)
	}
	if rec.flushCount == 0 {
		t.Fatal("expected Flush to be called")
	}
}

func TestFatalLogsAndInvokesExit(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	rec := setupRecordingLogger(t)
	prevExit := exitFunc
	exitCalled := 0
	exitCode := 0
	exitFunc = func(code int) {
		exitCalled++
		exitCode = code
	}
	t.Cleanup(func() { exitFunc = prevExit })

	Fatal("fatal-message")
	if !hasLevel(rec.entries, "critical") {
		t.Fatalf("expected critical entry, got %+v", rec.entries)
	}
	if exitCalled != 1 || exitCode != 1 {
		t.Fatalf("expected exitFunc to be called once with code 1, got count=%d code=%d", exitCalled, exitCode)
	}
}

func TestGetLogConfigContainsConsoleAndFile(t *testing.T) {
	config := getLogConfig()
	if !strings.Contains(config, "<console />") {
		t.Fatalf("expected console output in config, got %s", config)
	}
	if !strings.Contains(config, "log/cron.log") {
		t.Fatalf("expected file output in config, got %s", config)
	}
}

func hasLevel(entries []logEntry, level string) bool {
	for _, entry := range entries {
		if entry.level == level {
			return true
		}
	}
	return false
}
