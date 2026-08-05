package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/status"

	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/rpc/grpcpool"
	pb "github.com/gocronx-team/gocron/internal/modules/rpc/proto"
	"google.golang.org/grpc/codes"
)

var (
	taskCtxMap        sync.Map                        // 存储任务执行的 context.CancelFunc
	ErrManualStop     = errors.New("rpc_manual_stop") // 特殊错误标识，用于判断是否手动停止
	ErrTaskNotRunning = errors.New("rpc_task_not_running")
)

const maxStreamOutputBytes = 1 << 20

// errRPCUnavailable 在调用时翻译，以遵循启动时配置的服务端默认语言
// （不能用包级变量，否则会在配置加载前就固化为中文）。
func errRPCUnavailable() error {
	return errors.New(i18n.Translate("rpc_unavailable"))
}

func generateTaskUniqueKey(ip string, port int, id int64) string {
	return fmt.Sprintf("%s:%d:%d", ip, port, id)
}

func Stop(ip string, port int, id int64) error {
	addr := fmt.Sprintf("%s:%d", ip, port)
	c, err := grpcpool.Pool.Get(addr)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := c.Run(ctx, &pb.TaskRequest{Command: "__STOP__", Id: id})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		if resp.Error == "task not running" {
			return ErrTaskNotRunning
		}
		return errors.New(resp.Error)
	}
	return nil
}

func Exec(ip string, port int, taskReq *pb.TaskRequest) (string, error) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("panic#rpc/client.go:Exec#", err)
		}
	}()
	addr := fmt.Sprintf("%s:%d", ip, port)
	c, err := grpcpool.Pool.Get(addr)
	if err != nil {
		return "", err
	}
	if taskReq.Timeout <= 0 || taskReq.Timeout > 86400 {
		taskReq.Timeout = 86400
	}
	timeout := time.Duration(taskReq.Timeout) * time.Second
	// RPC context: 比任务超时多5秒，给服务端时间清理进程并返回输出
	ctx, cancel := context.WithTimeout(context.Background(), timeout+5*time.Second)
	defer cancel()

	taskUniqueKey := generateTaskUniqueKey(ip, port, taskReq.Id)
	taskCtxMap.Store(taskUniqueKey, cancel)
	defer taskCtxMap.Delete(taskUniqueKey)

	resp, err := c.Run(ctx, taskReq)

	// 处理响应：即使有错误，也要返回已产生的输出
	if err != nil {
		if resp != nil && resp.Output != "" {
			return resp.Output, parseGRPCErrorOnly(err)
		}
		return parseGRPCError(err)
	}

	if resp.Error == "" {
		return resp.Output, nil
	}

	// 检查是否是手动停止
	if resp.Error == "manual stop" {
		return resp.Output, ErrManualStop
	}

	return resp.Output, errors.New(resp.Error)
}

// ExecStream executes a task through the server-streaming RPC and invokes
// onChunk for each stdout/stderr fragment. Older nodes are detected through
// UNIMPLEMENTED and transparently use the original unary RPC.
//
// Once a streaming response has been received we never retry through Run: the
// remote process has started and a retry could execute the task twice.
func ExecStream(ip string, port int, taskReq *pb.TaskRequest, onChunk func(string)) (string, error) {
	addr := fmt.Sprintf("%s:%d", ip, port)
	c, err := grpcpool.Pool.Get(addr)
	if err != nil {
		return "", err
	}
	if taskReq.Timeout <= 0 || taskReq.Timeout > 86400 {
		taskReq.Timeout = 86400
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(taskReq.Timeout)*time.Second+5*time.Second)
	defer cancel()

	taskUniqueKey := generateTaskUniqueKey(ip, port, taskReq.Id)
	taskCtxMap.Store(taskUniqueKey, cancel)
	defer taskCtxMap.Delete(taskUniqueKey)

	stream, err := c.RunStream(ctx, taskReq)
	if err != nil {
		if status.Code(err) == codes.Unimplemented {
			return execUnaryWithClient(ctx, c, taskReq, onChunk)
		}
		return parseGRPCError(err)
	}

	var output strings.Builder
	received := false
	for {
		resp, recvErr := stream.Recv()
		if recvErr == io.EOF {
			return output.String(), nil
		}
		if recvErr != nil {
			// Some gRPC servers surface UNIMPLEMENTED on the first Recv rather
			// than when the stream is created. Falling back is safe only before
			// any response was observed.
			if !received && status.Code(recvErr) == codes.Unimplemented {
				return execUnaryWithClient(ctx, c, taskReq, onChunk)
			}
			return output.String(), parseGRPCErrorOnly(recvErr)
		}
		received = true
		if resp.Output != "" {
			if remaining := maxStreamOutputBytes - output.Len(); remaining > 0 {
				chunk := resp.Output
				if len(chunk) > remaining {
					chunk = chunk[:remaining]
				}
				output.WriteString(chunk)
			}
			if onChunk != nil {
				onChunk(resp.Output)
			}
		}
		if resp.Error != "" {
			if resp.Error == "manual stop" {
				return output.String(), ErrManualStop
			}
			return output.String(), errors.New(resp.Error)
		}
	}
}

func execUnaryWithClient(ctx context.Context, c pb.TaskClient, taskReq *pb.TaskRequest, onChunk func(string)) (string, error) {
	resp, err := c.Run(ctx, taskReq)
	if resp != nil && resp.Output != "" && onChunk != nil {
		onChunk(resp.Output)
	}
	if err != nil {
		if resp != nil {
			return resp.Output, parseGRPCErrorOnly(err)
		}
		return parseGRPCError(err)
	}
	if resp.Error == "" {
		return resp.Output, nil
	}
	if resp.Error == "manual stop" {
		return resp.Output, ErrManualStop
	}
	return resp.Output, errors.New(resp.Error)
}

func parseGRPCError(err error) (string, error) {
	switch status.Code(err) {
	case codes.Unavailable:
		return "", errRPCUnavailable()
	case codes.DeadlineExceeded:
		return "", errors.New(i18n.Translate("rpc_timeout"))
	case codes.Canceled:
		return "", ErrManualStop
	}
	return "", err
}

// parseGRPCErrorOnly 只返回错误，不返回输出
func parseGRPCErrorOnly(err error) error {
	switch status.Code(err) {
	case codes.Unavailable:
		return errRPCUnavailable()
	case codes.DeadlineExceeded:
		return errors.New(i18n.Translate("rpc_timeout"))
	case codes.Canceled:
		return ErrManualStop
	}
	return err
}
