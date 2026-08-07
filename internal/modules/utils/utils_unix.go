//go:build !windows
// +build !windows

package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Result struct {
	output string
	err    error
}

// detectBashPath 根据当前系统环境检测 bash 路径。
func detectBashPath() string {
	if path, err := exec.LookPath("bash"); err == nil {
		return path
	}
	if path, err := exec.LookPath("sh"); err == nil {
		return path
	}
	return "/bin/bash"
}

// 执行shell命令，可设置执行超时时间
// 改进：将命令写入临时脚本执行，即使超时或被取消，也会返回已产生的输出
// ExecShell 执行 shell 命令(不注入额外环境变量)。
func ExecShell(ctx context.Context, command string) (string, error) {
	return ExecShellWithEnv(ctx, command, nil)
}

// ExecShellWithEnv 执行 shell 命令,并在父进程环境基础上追加 env(机密等),
// 仅本次执行可见。
func ExecShellWithEnv(ctx context.Context, command string, env []string) (string, error) {
	return ExecShellWithEnvStream(ctx, command, env, nil)
}

type streamWriter struct {
	mu      sync.Mutex
	buf     bytes.Buffer
	onChunk func(string) error
}

func (w *streamWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf.Write(p)
	if w.onChunk != nil {
		_ = w.onChunk(string(p))
	}
	return len(p), nil
}

func (w *streamWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

// ExecShellWithEnvStream executes a command and reports stdout/stderr chunks
// as they arrive. onChunk is serialized across the two pipes.
func ExecShellWithEnvStream(ctx context.Context, command string, env []string, onChunk func(string) error) (string, error) {
	// 清理可能存在的 HTML 实体编码
	command = CleanHTMLEntities(command)
	// 将换行符统一替换为Unix风格的\n
	command = strings.ReplaceAll(command, "\r\n", "\n")

	// 使用系统临时目录
	tmpDir := os.TempDir()
	timestamp := time.Now().Format("20060102150405")
	scriptPattern := fmt.Sprintf("gocron_%s_*.sh", timestamp)

	tmpFile, err := os.CreateTemp(tmpDir, scriptPattern)
	if err != nil {
		return "", fmt.Errorf("创建临时脚本文件失败: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // 执行完毕后删除临时文件
	defer tmpFile.Close()

	// 将命令写入临时文件
	_, err = tmpFile.WriteString(command)
	if err != nil {
		return "", fmt.Errorf("写入脚本内容失败: %w", err)
	}

	// 确保文件写入磁盘
	err = tmpFile.Sync()
	if err != nil {
		return "", fmt.Errorf("同步文件失败: %w", err)
	}

	// 给脚本文件添加执行权限
	err = os.Chmod(tmpFile.Name(), 0700)
	if err != nil {
		return "", fmt.Errorf("设置脚本执行权限失败: %w", err)
	}

	// 显式关闭写句柄，避免 Unix/Linux 下写句柄未关闭导致的进程执行异常
	_ = tmpFile.Close()

	// 根据当前系统环境检测 bash 路径，执行脚本文件
	scriptPath := tmpFile.Name()
	bashPath := detectBashPath()
	cmd := exec.CommandContext(ctx, bashPath, scriptPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Cancel = func() error {
		if cmd.Process != nil && cmd.Process.Pid > 0 {
			// 先发送 SIGTERM 给进程组
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
			// 延迟发送 SIGKILL 确保子进程完全退出
			go func(pid int) {
				time.Sleep(300 * time.Millisecond)
				_ = syscall.Kill(-pid, syscall.SIGKILL)
			}(cmd.Process.Pid)
		}
		return nil
	}
	// 避免子进程未释放标准流描述符时一直阻塞
	cmd.WaitDelay = 2 * time.Second

	// 注入额外环境变量(机密等),在父进程环境基础上追加,仅本次执行可见
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	// 设置工作目录为用户家目录，避免 getcwd 错误
	if homeDir, err := os.UserHomeDir(); err == nil {
		cmd.Dir = homeDir
	} else {
		cmd.Dir = tmpDir
	}

	writer := &streamWriter{onChunk: onChunk}
	cmd.Stdout = writer
	cmd.Stderr = writer

	runErr := cmd.Run()
	output := writer.String()

	if ctx.Err() != nil {
		return output, errors.New("timeout killed")
	}

	return output, runErr
}
