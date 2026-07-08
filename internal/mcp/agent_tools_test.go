package mcp

import (
	"strings"
	"testing"

	"github.com/gocronx-team/gocron/internal/models"
)

func TestAgentToolDefs(t *testing.T) {
	defs := AgentToolDefs()
	if len(defs) != 8 {
		t.Fatalf("expected 8 tool defs, got %d", len(defs))
	}
	names := map[string]bool{}
	for _, d := range defs {
		if d.Type != "function" {
			t.Errorf("tool %q type = %q", d.Function.Name, d.Type)
		}
		if len(d.Function.Parameters) == 0 {
			t.Errorf("tool %q has empty parameters schema", d.Function.Name)
		}
		names[d.Function.Name] = true
	}
	for _, want := range []string{"list_tasks", "get_task", "query_task_logs", "list_hosts", "run_task", "create_task", "list_templates", "search_docs"} {
		if !names[want] {
			t.Errorf("missing tool def %q", want)
		}
	}
}

func TestCallTool_ReadTool(t *testing.T) {
	defer setupTestDb(t)()
	created := seedTask(t, "nightly-backup", models.Enabled)

	out, err := CallTool("get_task", []byte(`{"id":`+itoa(created.Id)+`}`), false)
	if err != nil {
		t.Fatalf("CallTool(get_task): %v", err)
	}
	task, ok := out.(models.Task)
	if !ok {
		t.Fatalf("expected models.Task, got %T", out)
	}
	if task.Name != "nightly-backup" {
		t.Fatalf("unexpected task: %+v", task)
	}
}

func TestCallTool_ListTasksEmptyArgs(t *testing.T) {
	defer setupTestDb(t)()
	seedTask(t, "a", models.Enabled)

	out, err := CallTool("list_tasks", nil, false)
	if err != nil {
		t.Fatalf("CallTool(list_tasks): %v", err)
	}
	res, ok := out.(listTasksOutput)
	if !ok {
		t.Fatalf("expected listTasksOutput, got %T", out)
	}
	if res.Total != 1 {
		t.Fatalf("expected total 1, got %d", res.Total)
	}
}

func TestCallTool_RunTaskRequiresAdmin(t *testing.T) {
	defer setupTestDb(t)()
	created := seedTask(t, "manual", models.Enabled)

	_, err := CallTool("run_task", []byte(`{"id":`+itoa(created.Id)+`}`), false)
	if err != errAdminRequired {
		t.Fatalf("expected errAdminRequired for non-admin, got %v", err)
	}
}

func TestCallTool_UnknownName(t *testing.T) {
	_, err := CallTool("bogus", nil, true)
	if err == nil || !strings.Contains(err.Error(), "unknown tool") {
		t.Fatalf("expected unknown tool error, got %v", err)
	}
}

func TestCallTool_InvalidArgs(t *testing.T) {
	defer setupTestDb(t)()
	_, err := CallTool("get_task", []byte(`{not json}`), false)
	if err == nil || !strings.Contains(err.Error(), "invalid tool arguments") {
		t.Fatalf("expected invalid arguments error, got %v", err)
	}
}

// itoa 避免在测试里引入 strconv 仅为拼接一个小整数。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
