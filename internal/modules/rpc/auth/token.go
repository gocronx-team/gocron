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

// TokenStreamServerInterceptor applies the same shared-token check to
// streaming RPCs.
func TokenStreamServerInterceptor(token string) grpc.StreamServerInterceptor {
	want := []byte(token)
	return func(srv interface{}, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(stream.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "missing rpc token")
		}
		values := md.Get(TokenMetadataKey)
		if len(values) == 0 || subtle.ConstantTimeCompare([]byte(values[0]), want) != 1 {
			return status.Error(codes.Unauthenticated, "invalid rpc token")
		}
		return handler(srv, stream)
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

// TokenStreamClientInterceptor is the streaming equivalent of
// TokenUnaryClientInterceptor. A separate interceptor is required because
// gRPC does not run unary interceptors for server-streaming methods.
func TokenStreamClientInterceptor(token string) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		ctx = metadata.AppendToOutgoingContext(ctx, TokenMetadataKey, token)
		return streamer(ctx, desc, cc, method, opts...)
	}
}
