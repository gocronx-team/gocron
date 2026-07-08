// Package ai 提供服务端的智能体对话端点：在服务端跑一个 LLM 工具调用循环，
// 复用既有 MCP 工具回答运维问题（哪些任务失败了、某任务详情、主机列表等），
// 并以 SSE 流式把内容增量、工具调用、工具结果推送给前端。
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocronx-team/gocron/internal/mcp"
	"github.com/gocronx-team/gocron/internal/models"
	"github.com/gocronx-team/gocron/internal/modules/i18n"
	"github.com/gocronx-team/gocron/internal/modules/llm"
	"github.com/gocronx-team/gocron/internal/modules/logger"
	"github.com/gocronx-team/gocron/internal/modules/utils"
	"github.com/gocronx-team/gocron/internal/routers/base"
	"github.com/gocronx-team/gocron/internal/routers/user"
	"github.com/gocronx-team/gocron/internal/service"
)

const (
	// 多步工具流 + 本地慢模型每轮思考都很久，整轮对话给较宽裕的预算；
	// 用户可随时用前端「停止」按钮中断（会取消请求 ctx，连带停掉后端调用）。
	chatTimeout   = 10 * time.Minute
	maxIterations = 6
)

// systemPrompt 约束模型角色与行为：用提供的工具回答运维问题，简洁，不编造任务数据。
const systemPrompt = `You are the AI ops assistant embedded in gocron, a distributed cron task scheduler.
Users ask operational questions about scheduled tasks, their execution logs, and the hosts that run them.

About gocron (authoritative facts — rely on these, do NOT contradict them with outdated general knowledge):
- gocron is a lightweight distributed cron task scheduler (Go backend, Vue web UI).
- It DOES support MCP (Model Context Protocol): a built-in MCP server is exposed at the /mcp endpoint, authenticated with MCP access tokens managed under 系统管理 → MCP 密钥. (This very AI assistant is part of gocron's AI integration.) Never say gocron lacks MCP support.
- Task protocols are HTTP (trigger an HTTP request) and Shell (run a command on an execution node via a gRPC agent). There is NO "TCP protocol".
- Other features: distributed execution nodes (agents over gRPC), JWT + TOTP two-factor auth, notifications (Slack / email / webhook), task templates, tags, audit log, log retention, and AI helpers (natural-language-to-cron, failure diagnosis, and this chat).
- "任务模板 / templates" are reusable task presets stored in the system — when the user asks about their templates, call list_templates to fetch the real list; do NOT confuse them with cron shortcut descriptors.
- Supported databases: MySQL, PostgreSQL, or SQLite.
- Do not claim gocron lacks a capability unless you are certain; prefer the facts above.

Operating principles:
- Tool use: call a tool when you need live data (which tasks/hosts/templates exist, execution logs). For how-to / concept / "does gocron support X" questions, call search_docs to ground your answer — except for cron syntax and the facts listed above, which you may answer directly.
- IMPORTANT: call search_docs AT MOST ONCE per question. Then answer from whatever snippets it returns. Do NOT search again with reworded queries — if the docs don't fully cover the topic, give the best answer from what you found and briefly note that the detail may not be documented. Repeated searching is not allowed.
- To analyze why runs failed, call query_task_logs (status=0 for failed) and analyze the returned "result" (the execution output) yourself — keep tool calls to a minimum (prefer one query that returns the data you need, then answer).
- CRITICAL: never end your turn by only announcing an action (e.g. "let me check the tasks first"). In a single turn you must EITHER actually emit the tool call(s) you need, OR give the complete final answer. Do not stop after a preamble.
- When you do use tools, look up real data before concluding — never fabricate task names, ids, statuses, or log contents.
- Creating tasks: when the user clearly asks to create/add a scheduled task, call create_task with your best-guess fields (name, 6-field cron spec, protocol, command). This does NOT create it directly — it opens a pre-filled editable form for the user to review and confirm. Only admins can do this; for non-admins the tool will refuse. Do NOT claim a task was created — say the confirmation form is shown.
- NEVER invent HTTP API endpoints, URLs, request bodies, field names, curl examples, OR config-file sections (e.g. app.ini "[notify]"/"[webhook]" — these do not exist). To create a task, either call create_task (preferred when they ask you to make one) or describe the Web UI flow (任务管理 → 新建任务: 名称、命令、Cron 表达式、协议、可选执行节点). Notifications (email/Slack/webhook) are configured in the Web UI under 系统管理 → 通知配置, NOT in any config file. If search_docs returns nothing for a topic, say it may not be documented and point to the relevant Web UI page — do NOT guess config keys or file sections.
- For task-log execution status: 0 = failed, 1 = running, 2 = success (finished), 3 = cancelled.
- Language: reason (think) AND answer in the SAME language as the user's latest message. If the user writes in Chinese, your reasoning/thinking must also be in Simplified Chinese, not English.
- Be concise.

Cron syntax (IMPORTANT — gocron uses SECOND-level cron, not standard 5-field Unix cron):
- A spec has 6 space-separated fields: second minute hour day-of-month month day-of-week (seconds come FIRST).
- Day-of-week: 0 = Sunday, 1-5 = Mon-Fri, 6 = Saturday.
- When sub-second precision is not needed, the second field is 0. Examples:
  every minute "0 * * * * *"; every 5 minutes "0 */5 * * * *"; every 20 seconds "*/20 * * * * *";
  daily at 09:30 "0 30 9 * * *"; weekdays at 09:00 "0 0 9 * * 1-5"; 1st of month at 00:00 "0 0 0 1 * *".
- Shortcut descriptors are also supported: @yearly, @monthly, @weekly, @daily (midnight), @hourly,
  @every <duration> (e.g. "@every 30s", "@every 1m20s"), and @reboot (run once at startup).
- Do NOT describe gocron as using 5-field cron; that is incorrect for this system.`

// ChatMessage 是请求中的一条对话消息。
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Messages []ChatMessage `json:"messages"`
}

// sseEvent 是发送给前端的一条 SSE 事件。
type sseEvent struct {
	event string
	data  any
}

// Chat 运行一个有界的 LLM 工具调用循环，并以 SSE 流式推送结果。
// 事件契约：
//   - reasoning    {"content": "<delta>"}                       思考过程增量（思考型模型）
//   - message      {"content": "<delta>"}                       内容增量
//   - tool_call    {"id","name","arguments"}                    模型决定调用工具
//   - tool_result  {"id","name","ok": true|false}               工具执行完成（不回传结果体）
//   - confirm_required {"task_id","task_name"}                   模型请求执行任务，需用户确认（不自动执行）
//   - create_proposal {name,spec,protocol,command,http_method,timeout}  模型建议新建任务，前端弹出可编辑确认表单（不自动创建）
//   - error        {"message": "<msg>"}                          运行期错误
//   - done         {}                                           始终最后发送
//
// 请求校验在写入 SSE 头之前完成，校验失败/未配置时以普通 JSON 错误响应返回，
// 与应用其余接口保持一致。
func Chat(c *gin.Context) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}
	if len(req.Messages) == 0 || strings.TrimSpace(req.Messages[len(req.Messages)-1].Content) == "" {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}

	client, err := llm.FromSettings()
	if err != nil {
		base.RespondError(c, i18n.T(c, "llm_not_configured"))
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	sendEvent := func(ev sseEvent) {
		payload, err := json.Marshal(ev.data)
		if err != nil {
			payload = []byte(`{}`)
		}
		_, _ = fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", ev.event, payload)
		c.Writer.Flush()
	}

	messages := buildMessages(req.Messages)
	isAdmin := user.IsAdmin(c)
	tools := mcp.AgentToolDefs()

	ctx, cancel := context.WithTimeout(c.Request.Context(), chatTimeout)
	defer cancel()

	defer sendEvent(sseEvent{event: "done", data: map[string]any{}})

	searchCount := 0 // search_docs 只真正执行一次，防止模型换词反复重搜拖慢

	for i := 0; i < maxIterations; i++ {
		msg, err := client.ChatStream(ctx, messages, tools,
			func(delta string) {
				sendEvent(sseEvent{event: "message", data: map[string]string{"content": delta}})
			},
			func(delta string) {
				sendEvent(sseEvent{event: "reasoning", data: map[string]string{"content": delta}})
			})
		if err != nil {
			logger.Errorf("AI对话#调用LLM失败#轮次%d#%s", i, err)
			// 仅向管理员透出 provider 真实错误(便于定位配置);普通用户给通用文案,
			// 避免把 LLM 端点地址/内网错误等基础设施细节泄露给非管理员。
			// err 已在 llm client 层对 api_key 脱敏。
			msg := i18n.T(c, "ai_chat_failed")
			if isAdmin {
				msg += ": " + err.Error()
			}
			sendEvent(sseEvent{event: "error", data: map[string]string{"message": msg}})
			return
		}
		logger.Infof("AI对话#轮次%d#内容长度%d#工具数%d", i, len(msg.Content), len(msg.ToolCalls))

		// 没有工具调用：模型已通过 message 事件流出终答。
		if len(msg.ToolCalls) == 0 {
			return
		}

		messages = append(messages, msg)
		for _, tc := range msg.ToolCalls {
			sendEvent(sseEvent{event: "tool_call", data: map[string]string{
				"id":        tc.ID,
				"name":      tc.Function.Name,
				"arguments": tc.Function.Arguments,
			}})

			// run_task 不在 agent 内直接执行：发 confirm_required 让用户在前端确认，
			// 真正的执行走 /api/ai/run-task（管理员 + 审计）。防止提示注入误触发执行。
			if tc.Function.Name == "run_task" {
				content := proposeRunTask(tc.Function.Arguments, isAdmin, sendEvent, tc.ID)
				messages = append(messages, llm.Message{Role: "tool", ToolCallID: tc.ID, Content: content})
				continue
			}

			// create_task 同理:不直接落库,发 create_proposal 让用户在前端可编辑确认表单里确认,
			// 真正的创建走既有 /api/task/store(管理员 + 审计 + 名称唯一校验)。
			if tc.Function.Name == "create_task" {
				content := proposeCreateTask(tc.Function.Arguments, isAdmin, sendEvent, tc.ID)
				messages = append(messages, llm.Message{Role: "tool", ToolCallID: tc.ID, Content: content})
				continue
			}

			// search_docs 只真正检索一次：第二次起直接催模型作答，不再重搜（避免反复搜拖慢）。
			if tc.Function.Name == "search_docs" {
				searchCount++
				if searchCount > 1 {
					sendEvent(sseEvent{event: "tool_result", data: map[string]any{"id": tc.ID, "name": "search_docs", "ok": true}})
					messages = append(messages, llm.Message{
						Role:       "tool",
						ToolCallID: tc.ID,
						Content:    "You already searched the docs once. Do NOT search again — answer the user now using the earlier search results.",
					})
					continue
				}
			}

			result, terr := safeCallTool(tc.Function.Name, tc.Function.Arguments, isAdmin)
			if terr != nil {
				logger.Errorf("AI对话#工具失败#%s#args=%s#%s", tc.Function.Name, tc.Function.Arguments, terr)
			} else {
				logger.Infof("AI对话#工具成功#%s", tc.Function.Name)
			}
			sendEvent(sseEvent{event: "tool_result", data: map[string]any{
				"id":   tc.ID,
				"name": tc.Function.Name,
				"ok":   terr == nil,
			}})

			messages = append(messages, llm.Message{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    toolResultContent(result, terr),
			})
		}
	}

	// 达到最大轮次仍未给出终答（模型一直在调工具/反复打转）。
	logger.Errorf("AI对话#达到最大轮次%d仍无终答", maxIterations)
	sendEvent(sseEvent{event: "error", data: map[string]string{"message": i18n.T(c, "ai_chat_max_iterations")}})
}

// safeCallTool 执行工具调用并兜底 panic，避免单个工具异常中断整个 SSE 流。
func safeCallTool(name, args string, isAdmin bool) (result any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("tool %s panicked: %v", name, r)
		}
	}()
	return mcp.CallTool(name, []byte(args), isAdmin)
}

// proposeRunTask 处理模型的 run_task 请求：不执行，只校验并向前端发 confirm_required，
// 返回给模型的 tool 结果说明"未执行、需用户确认"。返回值作为该 tool 调用的结果文本。
func proposeRunTask(args string, isAdmin bool, sendEvent func(sseEvent), toolCallID string) string {
	emitResult := func(ok bool) {
		sendEvent(sseEvent{event: "tool_result", data: map[string]any{"id": toolCallID, "name": "run_task", "ok": ok}})
	}
	if !isAdmin {
		emitResult(false)
		return "Permission denied: running a task requires an admin account. Not executed."
	}
	var in struct {
		Id int `json:"id"`
	}
	_ = json.Unmarshal([]byte(args), &in)
	if in.Id <= 0 {
		emitResult(false)
		return "Invalid task id. Not executed."
	}
	task, err := new(models.Task).Detail(in.Id)
	if err != nil || task.Id <= 0 {
		emitResult(false)
		return fmt.Sprintf("Task %d not found. Not executed.", in.Id)
	}
	emitResult(true)
	sendEvent(sseEvent{event: "confirm_required", data: map[string]any{"task_id": task.Id, "task_name": task.Name}})
	return fmt.Sprintf("Task '%s' (id %d) was NOT executed. Running a task needs explicit user confirmation; tell the user to click the confirm button if they want to run it.", task.Name, task.Id)
}

// proposeCreateTask 处理模型的 create_task 请求:不落库,只校验并向前端发 create_proposal
// (携带 AI 预填的任务字段),由用户在可编辑表单里确认;真正创建走 /api/task/store。
func proposeCreateTask(args string, isAdmin bool, sendEvent func(sseEvent), toolCallID string) string {
	emitResult := func(ok bool) {
		sendEvent(sseEvent{event: "tool_result", data: map[string]any{"id": toolCallID, "name": "create_task", "ok": ok}})
	}
	if !isAdmin {
		emitResult(false)
		return "Permission denied: creating a task requires an admin account. Nothing was created."
	}
	var in struct {
		Name             string `json:"name"`
		Spec             string `json:"spec"`
		Protocol         int    `json:"protocol"`
		Command          string `json:"command"`
		HttpMethod       int    `json:"http_method"`
		HttpBody         string `json:"http_body"`
		HttpHeaders      string `json:"http_headers"`
		SuccessPattern   string `json:"success_pattern"`
		Timeout          int    `json:"timeout"`
		Multi            int    `json:"multi"`
		RetryTimes       int    `json:"retry_times"`
		RetryInterval    int    `json:"retry_interval"`
		Tag              string `json:"tag"`
		Remark           string `json:"remark"`
		LogRetentionDays int    `json:"log_retention_days"`
	}
	_ = json.Unmarshal([]byte(args), &in)
	in.Name = strings.TrimSpace(in.Name)
	in.Command = strings.TrimSpace(in.Command)
	if in.Name == "" || in.Command == "" {
		emitResult(false)
		return "Missing task name or command. Nothing was created; ask the user for the missing detail."
	}
	if in.Protocol != int(models.TaskHTTP) && in.Protocol != int(models.TaskRPC) {
		emitResult(false)
		return "Invalid protocol (must be 1 for HTTP or 2 for Shell). Nothing was created."
	}
	// HTTP 默认 GET
	if in.Protocol == int(models.TaskHTTP) && in.HttpMethod != int(models.TaskHTTPMethodGet) && in.HttpMethod != int(models.TaskHttpMethodPost) {
		in.HttpMethod = int(models.TaskHTTPMethodGet)
	}
	if in.Multi != 0 && in.Multi != 1 {
		in.Multi = 0
	}

	emitResult(true)
	sendEvent(sseEvent{event: "create_proposal", data: map[string]any{
		"name":               in.Name,
		"spec":               strings.TrimSpace(in.Spec),
		"protocol":           in.Protocol,
		"command":            in.Command,
		"http_method":        in.HttpMethod,
		"http_body":          in.HttpBody,
		"http_headers":       in.HttpHeaders,
		"success_pattern":    in.SuccessPattern,
		"timeout":            in.Timeout,
		"multi":              in.Multi,
		"retry_times":        in.RetryTimes,
		"retry_interval":     in.RetryInterval,
		"tag":                in.Tag,
		"remark":             in.Remark,
		"log_retention_days": in.LogRetentionDays,
	}})
	return fmt.Sprintf("Proposed a new task '%s' but did NOT create it. A pre-filled confirmation form is now shown to the user; tell them to review/edit and click create. Do not claim the task was created.", in.Name)
}

// RunTask 是用户在聊天里点「确认执行」后真正触发任务的端点：仅管理员可用，且写审计。
// 路由 POST /api/ai/run-task/:id（不在 urlAuth 普通用户白名单内，故默认仅管理员可达）。
func RunTask(c *gin.Context) {
	if !user.IsAdmin(c) {
		base.RespondError(c, i18n.T(c, "unauthorized"))
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		base.RespondError(c, i18n.T(c, "param_error"))
		return
	}
	task, err := new(models.Task).Detail(id)
	if err != nil || task.Id <= 0 {
		base.RespondError(c, i18n.T(c, "get_task_detail_failed"))
		return
	}
	task.Spec = i18n.T(c, "manual_run")
	service.ServiceTask.Run(task)
	writeRunAudit(c, task.Id, task.Name)
	base.RespondSuccess(c, i18n.T(c, "task_started_check_log"), nil)
}

// writeRunAudit 为 AI 确认执行的任务写一条审计记录（GET /api/task/run 不经审计中间件，这里显式补）。
func writeRunAudit(c *gin.Context, id int, name string) {
	log := &models.AuditLog{
		Username:   user.Username(c),
		Ip:         utils.ClientIP(c),
		Module:     "task",
		Action:     "run",
		TargetId:   id,
		TargetName: name,
		Detail:     "AI 助手确认执行",
	}
	if _, err := log.Create(); err != nil {
		logger.Warnf("AI对话#执行任务审计写入失败#%s", err)
	}
}

// buildMessages 在用户消息前注入系统提示词（含当前服务器时间）。
func buildMessages(in []ChatMessage) []llm.Message {
	prompt := systemPrompt + fmt.Sprintf("\n\nCurrent server time: %s", time.Now().Format("2006-01-02 15:04:05 MST"))
	out := make([]llm.Message, 0, len(in)+1)
	out = append(out, llm.Message{Role: "system", Content: prompt})
	for _, m := range in {
		out = append(out, llm.Message{Role: m.Role, Content: m.Content})
	}
	return out
}

// toolResultContent 把工具结果或错误信息序列化为 tool 消息内容。
func toolResultContent(result any, err error) string {
	if err != nil {
		return err.Error()
	}
	encoded, mErr := json.Marshal(result)
	if mErr != nil {
		return mErr.Error()
	}
	return string(encoded)
}
