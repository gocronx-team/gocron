package server

import (
	"context"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	pb "github.com/gocronx-team/gocron/internal/modules/rpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Opt-in real binary contract test. Only loopback, disposable node processes are used.
func TestPreviousNodeBinaryContract(t *testing.T) {
	for _, key := range []string{"GOCRON_TEST_PREVIOUS_NODE", "GOCRON_TEST_CURRENT_NODE"} {
		t.Run(key, func(t *testing.T) {
			binary := os.Getenv(key)
			if binary == "" {
				t.Skip("node binary not configured")
			}
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			addr := listener.Addr().String()
			_ = listener.Close()
			cmd := exec.Command(binary, "-s", addr, "-allow-root")
			cmd.Env = append(os.Environ(), "GOCRON_NODE_TOKEN=")
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				_ = cmd.Process.Kill()
				_ = cmd.Wait()
			})
			conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = conn.Close() }()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			// Only the original unary method and fields: the previous scheduler's
			// wire contract must remain accepted by the upgraded node too.
			result, err := pb.NewTaskClient(conn).Run(ctx, &pb.TaskRequest{
				Id: 1, Command: "printf previous-wire-ok", Timeout: 2,
			}, grpc.WaitForReady(true))
			if err != nil {
				t.Fatal(err)
			}
			if result.Output != "previous-wire-ok" || result.Error != "" {
				t.Fatalf("result=%v", result)
			}
		})
	}
}
