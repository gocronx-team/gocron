//go:build !windows
// +build !windows

package utils

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestExecShellSuccess(t *testing.T) {
	ctx := context.Background()
	output, err := ExecShell(ctx, "echo 'hello world'")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !strings.Contains(output, "hello world") {
		t.Fatalf("Expected output to contain 'hello world', got: %s", output)
	}
}

func TestExecShellTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 运行一个会产生输出然后睡眠的命令
	output, err := ExecShell(ctx, "echo 'partial output'; sleep 1; echo 'should not see this'")

	if err == nil {
		t.Fatal("Expected timeout error")
	}
	if err.Error() != "timeout killed" {
		t.Fatalf("Expected 'timeout killed' error, got: %v", err)
	}
	if !strings.Contains(output, "partial output") {
		t.Fatalf("Expected partial output to contain 'partial output', got: %s", output)
	}
	if strings.Contains(output, "should not see this") {
		t.Fatalf("Should not contain output after timeout, got: %s", output)
	}
}

func TestExecShellCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// 启动一个长时间运行的命令
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel() // 手动取消
	}()

	output, err := ExecShell(ctx, "echo 'before cancel'; sleep 1; echo 'after cancel'")

	if err == nil {
		t.Fatal("Expected cancel error")
	}
	if err.Error() != "timeout killed" {
		t.Fatalf("Expected 'timeout killed' error, got: %v", err)
	}
	if !strings.Contains(output, "before cancel") {
		t.Fatalf("Expected partial output to contain 'before cancel', got: %s", output)
	}
}

func TestExecShellCommandError(t *testing.T) {
	ctx := context.Background()
	output, err := ExecShell(ctx, "nonexistentcommand")

	if err == nil {
		t.Fatal("Expected command error")
	}
	// 应该有错误输出
	if output == "" {
		t.Fatal("Expected some error output")
	}
}

func TestDetectBashPath(t *testing.T) {
	path := detectBashPath()

	// 路径不能为空
	if path == "" {
		t.Fatal("Expected non-empty bash path")
	}

	// 路径应该以 "bash" 结尾
	if !strings.HasSuffix(path, "bash") {
		t.Fatalf("Expected path to end with 'bash', got: %s", path)
	}

	// 验证路径指向一个存在的文件，而非目录
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Bash path does not exist: %s, error: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("Bash path is a directory, not a file: %s", path)
	}

	// 验证该路径确实可以执行 shell 命令
	cmd := exec.Command(path, "-c", "echo 'bash_ok'")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to execute bash at %s: %v", path, err)
	}
	if !strings.Contains(string(output), "bash_ok") {
		t.Fatalf("Expected output to contain 'bash_ok', got: %s", string(output))
	}
}
