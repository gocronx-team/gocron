//go:build windows
// +build windows

package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type Result struct {
	output string
	err    error
}

// 执行shell命令，可设置执行超时时间
// 改进：将命令写入临时批处理文件执行，即使超时或被取消，也会返回已产生的输出
// ExecShell 执行命令(不注入额外环境变量)。
func ExecShell(ctx context.Context, command string) (string, error) {
	return ExecShellWithEnv(ctx, command, nil)
}

// ExecShellWithEnv 执行命令,并在父进程环境基础上追加 env(机密等),仅本次执行可见。
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
		_ = w.onChunk(ConvertEncoding(string(p)))
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
	// 清理可能存在的 HTML 实体编码,防止 &quot; 等导致命令执行失败
	// 例如: del &quot;C:\file.txt&quot; -> del "C:\file.txt"
	command = CleanHTMLEntities(command)

	// 将换行符统一替换为Windows风格的\r\n
	command = strings.ReplaceAll(command, "\r\n", "\n")
	command = strings.ReplaceAll(command, "\n", "\r\n")

	// 创建带时间戳的临时批处理文件名
	timestamp := time.Now().Format("20060102150405") // 年月日时分秒

	// 使用 os.CreateTemp 创建临时文件
	batFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("gocron_%s_*.bat", timestamp))
	if err != nil {
		return "", fmt.Errorf("创建临时批处理文件失败: %w", err)
	}
	defer os.Remove(batFile.Name()) // 确保函数退出时删除临时文件
	defer batFile.Close()

	// 将命令写入批处理文件
	content := "@echo off\r\n" + command

	// 使用 ANSI 编码 (GBK) 写入批处理文件
	gbkWriter := transform.NewWriter(batFile, simplifiedchinese.GBK.NewEncoder())
	_, err = io.WriteString(gbkWriter, content)

	if err != nil {
		return "", fmt.Errorf("写入批处理文件失败: %w", err)
	}

	// 关闭编码 writer,flush GBK 编码缓冲区中的剩余字节,否则末尾内容可能未写入导致 bat 被截断
	if err = gbkWriter.Close(); err != nil {
		return "", fmt.Errorf("刷新批处理文件失败: %w", err)
	}

	// 确保文件内容写入磁盘并显式关闭句柄
	err = batFile.Sync()
	if err != nil {
		return "", fmt.Errorf("同步批处理文件失败: %w", err)
	}
	_ = batFile.Close()

	// 使用 cmd.exe 执行批处理文件
	cmd := exec.CommandContext(ctx, "cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
		CmdLine:    `cmd /c "` + batFile.Name() + `"`,
	}
	cmd.Cancel = func() error {
		if cmd.Process != nil && cmd.Process.Pid > 0 {
			_ = exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
		}
		return nil
	}
	cmd.WaitDelay = 2 * time.Second

	// 注入额外环境变量(机密等),在父进程环境基础上追加,仅本次执行可见
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	// 设置工作目录为用户家目录，避免 getcwd 错误
	if homeDir, err := os.UserHomeDir(); err == nil {
		cmd.Dir = homeDir
	} else {
		cmd.Dir = os.TempDir()
	}

	writer := &streamWriter{onChunk: onChunk}
	cmd.Stdout = writer
	cmd.Stderr = writer

	runErr := cmd.Run()
	output := writer.String()

	if ctx.Err() != nil {
		return ConvertEncoding(output), errors.New("timeout killed")
	}

	return ConvertEncoding(output), runErr
}

func ConvertEncoding(outputGBK string) string {
	// windows平台编码为gbk，需转换为utf8才能入库
	outputUTF8, ok := GBK2UTF8(outputGBK)
	if ok {
		return outputUTF8
	}

	return outputGBK
}
