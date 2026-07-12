package auth

import (
	"context"
	"crypto/subtle"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TokenMetadataKey 是节点 RPC 共享令牌在 gRPC metadata 中的键名。
// 与 TLS 正交:即便未开启 TLS,只要调度器与节点配置了相同令牌,
// 明文通道上的每次调用也会被校验,可关闭默认无鉴权导致的远程命令执行风险。
const TokenMetadataKey = "x-gocron-token"

// TokenUnaryServerInterceptor 返回一个服务端一元拦截器,要求每次调用携带与
// token 相等的令牌(常量时间比较)。token 为空时不应安装该拦截器(保持旧行为)。
func TokenUnaryServerInterceptor(token string) grpc.UnaryServerInterceptor {
	want := []byte(token)
	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing rpc token")
		}
		values := md.Get(TokenMetadataKey)
		if len(values) == 0 || subtle.ConstantTimeCompare([]byte(values[0]), want) != 1 {
			return nil, status.Error(codes.Unauthenticated, "invalid rpc token")
		}
		return handler(ctx, req)
	}
}

// TokenUnaryClientInterceptor 返回一个客户端一元拦截器,为每次调用附加共享令牌。
// token 为空时不应安装该拦截器(保持旧行为)。
func TokenUnaryClientInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx = metadata.AppendToOutgoingContext(ctx, TokenMetadataKey, token)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
