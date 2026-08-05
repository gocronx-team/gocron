package server

import (
	"bytes"
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gocronx-team/gocron/internal/modules/rpc/auth"
	pb "github.com/gocronx-team/gocron/internal/modules/rpc/proto"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type Server struct {
	pb.UnimplementedTaskServer
	taskContexts sync.Map // 存储正在运行的任务上下文
	taskOutputs  sync.Map // 存储任务输出
	stopChans    sync.Map // 存储停止通道
}

var keepAlivePolicy = keepalive.EnforcementPolicy{
	MinTime:             10 * time.Second,
	PermitWithoutStream: true,
}

var keepAliveParams = keepalive.ServerParameters{
	MaxConnectionIdle: 30 * time.Second,
	Time:              30 * time.Second,
	Timeout:           3 * time.Second,
}

func (s *Server) Run(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	return s.execute(ctx, req, nil)
}

// RunStream streams stdout/stderr and finishes with an optional error frame.
// The original Run method is retained for older schedulers.
func (s *Server) RunStream(req *pb.TaskRequest, stream grpc.ServerStreamingServer[pb.TaskResponse]) error {
	resp, err := s.execute(stream.Context(), req, func(chunk string) error {
		return stream.Send(&pb.TaskResponse{Output: chunk})
	})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return stream.Send(&pb.TaskResponse{Error: resp.Error})
	}
	return nil
}

func (s *Server) execute(ctx context.Context, req *pb.TaskRequest, onChunk func(string) error) (*pb.TaskResponse, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	// 清理 HTML 实体
	cleanedCmd := utils.CleanHTMLEntities(req.Command)

	// 检测是否是停止信号
	if cleanedCmd == "__STOP__" {
		if ch, ok := s.stopChans.LoadAndDelete(req.Id); ok {
			close(ch.(chan struct{}))
			return &pb.TaskResponse{}, nil
		}
		return &pb.TaskResponse{Error: "task not running"}, nil
	}

	// 使用任务超时创建独立的 context
	timeout := time.Duration(req.Timeout) * time.Second
	taskCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 存储任务上下文和输出 buffer
	outputBuf := &bytes.Buffer{}
	stopChan := make(chan struct{})
	s.taskContexts.Store(req.Id, cancel)
	s.taskOutputs.Store(req.Id, outputBuf)
	s.stopChans.Store(req.Id, stopChan)
	defer func() {
		s.taskContexts.Delete(req.Id)
		s.stopChans.Delete(req.Id)
		// 保留输出 5 秒，给 Stop 调用时间获取
		time.AfterFunc(5*time.Second, func() {
			s.taskOutputs.Delete(req.Id)
		})
	}()

	// 监听客户端取消或停止信号
	var wasStopped atomic.Bool
	go func() {
		select {
		case <-ctx.Done():
			cancel()
		case <-stopChan:
			wasStopped.Store(true)
			cancel()
		case <-taskCtx.Done():
		}
	}()

	// 执行命令,注入随请求下发的环境变量(机密等)
	output, execErr := utils.ExecShellWithEnvStream(taskCtx, cleanedCmd, envSlice(req.Env), onChunk)
	outputBuf.WriteString(output)

	resp := new(pb.TaskResponse)
	resp.Output = output
	if execErr != nil {
		// 如果是手动停止，使用特定的错误信息
		if wasStopped.Load() {
			resp.Error = "manual stop"
			log.Infof("[id: %d] Manually stopped (output %d bytes)", req.Id, len(output))
		} else {
			resp.Error = execErr.Error()
			log.Infof("[id: %d] Execution failed: %s (output %d bytes)", req.Id, execErr.Error(), len(output))
		}
	} else {
		resp.Error = ""
		log.Infof("[id: %d] Execution successful (output %d bytes)", req.Id, len(output))
	}

	return resp, nil
}

func Start(addr string, enableTLS bool, certificate auth.Certificate, token string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepAliveParams),
		grpc.KeepaliveEnforcementPolicy(keepAlivePolicy),
	}
	if enableTLS {
		tlsConfig, err := certificate.GetTLSConfigForServer()
		if err != nil {
			log.Fatal(err)
		}
		opt := grpc.Creds(credentials.NewTLS(tlsConfig))
		opts = append(opts, opt)
	}
	// 配置了共享令牌时,对每次调用强制校验(与 TLS 正交)。
	if token != "" {
		opts = append(opts,
			grpc.UnaryInterceptor(auth.TokenUnaryServerInterceptor(token)),
			grpc.StreamInterceptor(auth.TokenStreamServerInterceptor(token)),
		)
	} else if !enableTLS {
		log.Warn("gocron-node is running WITHOUT TLS and WITHOUT a shared token: " +
			"anyone able to reach this address can execute arbitrary commands. " +
			"Set -enable-tls or -token (env GOCRON_NODE_TOKEN) to secure it.")
	}
	server := grpc.NewServer(opts...)
	pb.RegisterTaskServer(server, &Server{})
	log.Infof("server listen on %s", addr)

	go func() {
		err = server.Serve(l)
		if err != nil {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		s := <-c
		log.Infoln("Received signal -- ", s)
		switch s {
		case syscall.SIGHUP:
			log.Infoln("Received terminal disconnect signal, ignoring")
		case syscall.SIGINT, syscall.SIGTERM:
			log.Info("Application preparing to exit")
			server.GracefulStop()
			return
		}
	}

}

// envSlice 把 proto 的 env map 转为 "KEY=VALUE" 形式的切片,供 exec.Cmd.Env 使用。
func envSlice(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}
