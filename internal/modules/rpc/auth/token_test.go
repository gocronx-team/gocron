package auth

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func okHandler(_ context.Context, _ interface{}) (interface{}, error) {
	return "ok", nil
}

func TestTokenUnaryServerInterceptor(t *testing.T) {
	const token = "s3cret-token"
	interceptor := TokenUnaryServerInterceptor(token)

	cases := []struct {
		name    string
		ctx     context.Context
		wantErr codes.Code // codes.OK 表示应放行
	}{
		{
			name:    "no metadata",
			ctx:     context.Background(),
			wantErr: codes.Unauthenticated,
		},
		{
			name:    "missing token key",
			ctx:     metadata.NewIncomingContext(context.Background(), metadata.Pairs("other", "x")),
			wantErr: codes.Unauthenticated,
		},
		{
			name:    "wrong token",
			ctx:     metadata.NewIncomingContext(context.Background(), metadata.Pairs(TokenMetadataKey, "nope")),
			wantErr: codes.Unauthenticated,
		},
		{
			name:    "correct token",
			ctx:     metadata.NewIncomingContext(context.Background(), metadata.Pairs(TokenMetadataKey, token)),
			wantErr: codes.OK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := interceptor(tc.ctx, nil, &grpc.UnaryServerInfo{}, okHandler)
			if tc.wantErr == codes.OK {
				if err != nil {
					t.Fatalf("expected pass, got error: %v", err)
				}
				if resp != "ok" {
					t.Fatalf("expected handler result, got %v", resp)
				}
				return
			}
			if status.Code(err) != tc.wantErr {
				t.Fatalf("expected code %v, got %v (err=%v)", tc.wantErr, status.Code(err), err)
			}
		})
	}
}

// TestTokenClientInterceptorAttachesToken 验证客户端拦截器把令牌写入 outgoing metadata,
// 且能被服务端拦截器接受(闭环)。
func TestTokenClientInterceptorAttachesToken(t *testing.T) {
	const token = "round-trip-token"
	clientIC := TokenUnaryClientInterceptor(token)
	serverIC := TokenUnaryServerInterceptor(token)

	// 客户端拦截器的 invoker:把 outgoing metadata 转成 incoming,交给服务端拦截器。
	invoker := func(ctx context.Context, _ string, _, _ interface{}, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			t.Fatal("expected outgoing metadata")
		}
		inCtx := metadata.NewIncomingContext(context.Background(), md)
		_, err := serverIC(inCtx, nil, &grpc.UnaryServerInfo{}, okHandler)
		return err
	}

	if err := clientIC(context.Background(), "/svc/M", nil, nil, nil, invoker); err != nil {
		t.Fatalf("round trip should succeed, got: %v", err)
	}
}
