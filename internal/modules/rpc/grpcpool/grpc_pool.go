package grpcpool

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gocronx-team/gocron/internal/modules/ca"
	"github.com/gocronx-team/gocron/internal/modules/rpc/auth"
	"github.com/gocronx-team/gocron/internal/modules/rpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	backOffMaxDelay = 3 * time.Second
	dialTimeout     = 2 * time.Second
)

var (
	Pool = &GRPCPool{
		conns: make(map[string]*Client),
	}

	keepAliveParams = keepalive.ClientParameters{
		Time:                20 * time.Second,
		Timeout:             3 * time.Second,
		PermitWithoutStream: true,
	}
)

type Client struct {
	conn      *grpc.ClientConn
	rpcClient rpc.TaskClient
}

type GRPCPool struct {
	// map key格式 ip:port
	conns map[string]*Client
	mu    sync.RWMutex
}

func (p *GRPCPool) Get(addr string) (rpc.TaskClient, error) {
	p.mu.RLock()
	client, ok := p.conns[addr]
	p.mu.RUnlock()
	if ok {
		return client.rpcClient, nil
	}

	client, err := p.factory(addr)
	if err != nil {
		return nil, err
	}

	return client.rpcClient, nil
}

// 释放连接
func (p *GRPCPool) Release(addr string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	client, ok := p.conns[addr]
	if !ok {
		return
	}
	delete(p.conns, addr)
	client.conn.Close()
}

// 创建连接
func (p *GRPCPool) factory(addr string) (*Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	client, ok := p.conns[addr]
	if ok {
		return client, nil
	}
	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepAliveParams),
		grpc.WithBackoffMaxDelay(backOffMaxDelay),
	}

	// 默认启用 TLS
	enableTLS := true
	if !enableTLS {
		opts = append(opts, grpc.WithInsecure())
	} else {
		server := strings.Split(addr, ":")
		// 确保 CA 已初始化
		_ = ca.GetGlobalCA()
		
		// 使用统一的 CA 和客户端证书
		caCertPath := ca.GetCACertPath()
		clientCertPath, clientKeyPath := ca.GetClientCertPath()
		
		certificate := auth.Certificate{
			CAFile:     caCertPath,
			CertFile:   clientCertPath,
			KeyFile:    clientKeyPath,
			ServerName: server[0],
		}

		transportCreds, err := certificate.GetTransportCredsForClient()
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(transportCreds))
	}

	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, err
	}

	client = &Client{
		conn:      conn,
		rpcClient: rpc.NewTaskClient(conn),
	}

	p.conns[addr] = client

	return client, nil
}
