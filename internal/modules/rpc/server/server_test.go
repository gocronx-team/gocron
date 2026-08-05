package server

import (
	"context"
	"strings"
	"testing"

	pb "github.com/gocronx-team/gocron/internal/modules/rpc/proto"
	"google.golang.org/grpc/metadata"
)

type captureTaskStream struct {
	ctx       context.Context
	responses []*pb.TaskResponse
}

func (s *captureTaskStream) Send(resp *pb.TaskResponse) error {
	s.responses = append(s.responses, resp)
	return nil
}
func (s *captureTaskStream) SetHeader(metadata.MD) error  { return nil }
func (s *captureTaskStream) SendHeader(metadata.MD) error { return nil }
func (s *captureTaskStream) SetTrailer(metadata.MD)       {}
func (s *captureTaskStream) Context() context.Context     { return s.ctx }
func (s *captureTaskStream) SendMsg(interface{}) error    { return nil }
func (s *captureTaskStream) RecvMsg(interface{}) error    { return nil }

func TestEnvSlice(t *testing.T) {
	if got := envSlice(nil); got != nil {
		t.Errorf("expected nil for empty map, got %v", got)
	}

	got := envSlice(map[string]string{"API_KEY": "v1"})
	if len(got) != 1 || got[0] != "API_KEY=v1" {
		t.Errorf("unexpected single-entry result: %v", got)
	}

	// 多元素:顺序无关,逐项验证 KEY=VALUE 格式
	multi := envSlice(map[string]string{"A": "1", "B": "2"})
	if len(multi) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(multi))
	}
	set := map[string]bool{multi[0]: true, multi[1]: true}
	if !set["A=1"] || !set["B=2"] {
		t.Errorf("unexpected pairs: %v", multi)
	}
}

func TestRunStreamReturnsIncrementalOutput(t *testing.T) {
	stream := &captureTaskStream{ctx: context.Background()}
	err := new(Server).RunStream(&pb.TaskRequest{
		Id: 1, Timeout: 5, Command: "printf first; sleep 0.1; printf second",
	}, stream)
	if err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	for _, resp := range stream.responses {
		output.WriteString(resp.Output)
		if resp.Error != "" {
			t.Fatalf("unexpected error frame: %q", resp.Error)
		}
	}
	if output.String() != "firstsecond" {
		t.Fatalf("streamed output = %q", output.String())
	}
}

func TestStopAcknowledgesOnlyRunningTask(t *testing.T) {
	s := new(Server)
	stop := make(chan struct{})
	s.stopChans.Store(int64(42), stop)

	resp, err := s.Run(context.Background(), &pb.TaskRequest{Id: 42, Command: "__STOP__"})
	if err != nil || resp.Error != "" {
		t.Fatalf("first stop = (%v, %v), want acknowledgement", resp, err)
	}
	select {
	case <-stop:
	default:
		t.Fatal("stop channel was not closed")
	}

	resp, err = s.Run(context.Background(), &pb.TaskRequest{Id: 42, Command: "__STOP__"})
	if err != nil || resp.Error != "task not running" {
		t.Fatalf("second stop = (%v, %v), want task-not-running response", resp, err)
	}
}
