package ai

import (
	"strings"
	"testing"
)

// captureEvents 收集 proposeCreateTask 发出的 SSE 事件,便于断言。
func captureEvents() (func(sseEvent), *[]sseEvent) {
	var events []sseEvent
	return func(ev sseEvent) { events = append(events, ev) }, &events
}

func findEvent(events []sseEvent, name string) (sseEvent, bool) {
	for _, e := range events {
		if e.event == name {
			return e, true
		}
	}
	return sseEvent{}, false
}

func TestProposeCreateTask(t *testing.T) {
	t.Run("non-admin refused, no proposal", func(t *testing.T) {
		send, events := captureEvents()
		msg := proposeCreateTask(`{"name":"x","spec":"0 * * * * *","protocol":1,"command":"http://x"}`, false, send, "tc1")
		if !strings.Contains(msg, "admin") {
			t.Errorf("expected admin-required message, got %q", msg)
		}
		if _, ok := findEvent(*events, "create_proposal"); ok {
			t.Error("non-admin must not get a create_proposal event")
		}
	})

	t.Run("missing command refused", func(t *testing.T) {
		send, events := captureEvents()
		msg := proposeCreateTask(`{"name":"x","spec":"0 * * * * *","protocol":1,"command":""}`, true, send, "tc")
		if !strings.Contains(strings.ToLower(msg), "missing") {
			t.Errorf("expected missing-field message, got %q", msg)
		}
		if _, ok := findEvent(*events, "create_proposal"); ok {
			t.Error("must not propose when command is empty")
		}
	})

	t.Run("invalid protocol refused", func(t *testing.T) {
		send, events := captureEvents()
		msg := proposeCreateTask(`{"name":"x","spec":"0 * * * * *","protocol":9,"command":"echo hi"}`, true, send, "tc")
		if !strings.Contains(strings.ToLower(msg), "protocol") {
			t.Errorf("expected protocol error, got %q", msg)
		}
		if _, ok := findEvent(*events, "create_proposal"); ok {
			t.Error("must not propose with invalid protocol")
		}
	})

	t.Run("valid HTTP defaults method to GET and emits proposal", func(t *testing.T) {
		send, events := captureEvents()
		msg := proposeCreateTask(`{"name":"ping","spec":"0 0 1 * * *","protocol":1,"command":"http://example.com"}`, true, send, "tc")
		if strings.Contains(strings.ToLower(msg), "created") && !strings.Contains(msg, "NOT create") && !strings.Contains(msg, "did NOT create") {
			t.Errorf("message should make clear the task was NOT created, got %q", msg)
		}
		ev, ok := findEvent(*events, "create_proposal")
		if !ok {
			t.Fatal("expected create_proposal event")
		}
		data, _ := ev.data.(map[string]any)
		if data["http_method"] != 1 {
			t.Errorf("http_method should default to 1 (GET), got %v", data["http_method"])
		}
		if data["name"] != "ping" {
			t.Errorf("name = %v, want ping", data["name"])
		}
	})
}
