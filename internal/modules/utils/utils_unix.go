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

	// 根据当前系统环境检测 bash 路径，执行脚本文件
	scriptPath := tmpFile.Name()
	bashPath := detectBashPath()
	cmd := exec.Command(bashPath, scriptPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
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

	// 使用管道实时捕获输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	// 用于收集输出
	var outputBuffer bytes.Buffer
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 启动命令
	if err := cmd.Start(); err != nil {
		return "", err
	}

	// 实时读取 stdout 和 stderr
	wg.Add(2)
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				mu.Lock()
				outputBuffer.Write(buf[:n])
				mu.Unlock()
			}
			if err != nil {
				break
			}
		}
	}()
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				mu.Lock()
				outputBuffer.Write(buf[:n])
				mu.Unlock()
			}
			if err != nil {
				break
			}
		}
	}()

	// 等待命令完成或超时
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		// 超时或被取消，尝试优雅终止
		if cmd.Process != nil && cmd.Process.Pid > 0 {
			// 先发送 SIGTERM，给进程清理的机会
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)

			// 等待 2 秒，看进程是否自行退出
			timer := time.NewTimer(2 * time.Second)
			select {
			case <-done:
				timer.Stop()
			case <-timer.C:
				// 进程仍未退出，强制杀死
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				<-done // 等待 Wait() 返回
			}
		}

		// 等待 IO 读取完成
		wg.Wait()

		// 返回已捕获的输出和错误信息
		mu.Lock()
		output := outputBuffer.String()
		mu.Unlock()
		return output, errors.New("timeout killed")

	case err := <-done:
		// 命令正常完成
		wg.Wait()
		mu.Lock()
		output := outputBuffer.String()
		mu.Unlock()
		return output, err
	}
}
