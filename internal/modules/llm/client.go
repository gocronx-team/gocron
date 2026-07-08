// Package llm 提供一个最小的 OpenAI 兼容 Chat Completions 客户端，
// 用于自然语言转 cron、失败日志诊断等产品内 AI 功能。
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// 本地大模型（如 MLX 上的 27B/35B）首次推理较慢，给较宽松的超时。
const defaultTimeout = 120 * time.Second

var (
	// ErrNotConfigured 表示 LLM 未启用或配置不完整。
	ErrNotConfigured = errors.New("llm not configured")
	// ErrEmptyResponse 表示模型返回了空内容。
	ErrEmptyResponse = errors.New("llm returned empty response")
)

// Client 是一个最小的 OpenAI 兼容 Chat Completions 客户端。
type Client struct {
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
	// streamHTTP 专用于流式：不设总超时（http.Client.Timeout 会把"读完整个响应体"也计时，
	// 慢模型的长流式输出会被中途掐断）。流式的取消/超时完全交给调用方传入的 ctx 控制。
	streamHTTP *http.Client
}

// New 创建客户端。baseURL 形如 https://api.openai.com/v1。
func New(baseURL, apiKey, model string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:     strings.TrimSpace(apiKey),
		model:      strings.TrimSpace(model),
		streamHTTP: &http.Client{},
		http:       &http.Client{Timeout: defaultTimeout},
	}
}

// redact 从对外暴露的错误信息里抹掉 api_key,避免上游把 token 回显到错误里被透传泄露。
func (c *Client) redact(s string) string {
	if c.apiKey == "" {
		return s
	}
	return strings.ReplaceAll(s, c.apiKey, "***")
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Message 是 OpenAI 兼容的对话消息，支持工具调用（tool calling）。
// content 始终序列化（assistant/tool 消息可能内容为空但仍需出现该字段）；
// tool_calls 仅在非空时序列化。
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// ToolCall 表示模型请求调用某个工具。
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall 中的 Arguments 按 OpenAI 规范是一段 JSON 编码的字符串。
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Tool 是暴露给模型的一个可调用工具的定义。
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction 描述工具的名称、用途与参数 JSON Schema。
type ToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Chat 发送一次单轮对话，返回模型回复文本。
// 调用方应通过 ctx 控制取消/超时；客户端自身也带有默认超时。
func (c *Client) Chat(ctx context.Context, system, user string) (string, error) {
	// 请求发出前校验 base_url(收敛 SSRF 面:仅允许带 host 的 http/https)。
	if err := ValidateBaseURL(c.baseURL); err != nil {
		return "", err
	}
	reqBody := chatRequest{
		Model:       c.model,
		Temperature: 0.2,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("call llm: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("decode response (status %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil && parsed.Error.Message != "" {
			return "", fmt.Errorf("llm error (status %d): %s", resp.StatusCode, c.redact(parsed.Error.Message))
		}
		return "", fmt.Errorf("llm http status %d", resp.StatusCode)
	}
	if len(parsed.Choices) == 0 {
		return "", ErrEmptyResponse
	}
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return "", ErrEmptyResponse
	}
	return content, nil
}

type toolChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
	Temperature float64   `json:"temperature"`
}

type toolChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ChatWithTools 发送一次带工具定义的多轮对话，返回 choices[0].message。
// 返回的 Message 可能包含 Content 和/或 ToolCalls，由调用方据此驱动工具调用循环。
// 与 Chat 复用相同的鉴权头与错误处理；choices 为空时返回 ErrEmptyResponse。
func (c *Client) ChatWithTools(ctx context.Context, messages []Message, tools []Tool) (Message, error) {
	reqBody := toolChatRequest{
		Model:       c.model,
		Messages:    messages,
		Tools:       tools,
		Temperature: 0.2,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return Message{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return Message{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("call llm: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Message{}, fmt.Errorf("read response: %w", err)
	}

	var parsed toolChatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return Message{}, fmt.Errorf("decode response (status %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil && parsed.Error.Message != "" {
			return Message{}, fmt.Errorf("llm error (status %d): %s", resp.StatusCode, parsed.Error.Message)
		}
		return Message{}, fmt.Errorf("llm http status %d", resp.StatusCode)
	}
	if len(parsed.Choices) == 0 {
		return Message{}, ErrEmptyResponse
	}
	return parsed.Choices[0].Message, nil
}

type streamChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
	Temperature float64   `json:"temperature"`
	Stream      bool      `json:"stream"`
}

// streamChunk 是 OpenAI 兼容流式分片（SSE data 行）的结构。
type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
			// ReasoningContent 是「思考型」模型（如 Qwen3 thinking、DeepSeek-R1）的推理过程增量，
			// 与最终答案 Content 分开下发；需单独流式展示，否则思考期间前端会长时间无任何输出。
			ReasoningContent string `json:"reasoning_content"`
			ToolCalls        []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// streamScannerBuffer 是扫描器的最大单行缓冲（部分模型一行 data 可能很长）。
const streamScannerBuffer = 1 << 20 // 1MB

// ChatStream 发送一次带工具定义的流式对话，按 SSE 分片累积内容与工具调用。
// 每段非空 content 增量回调 onContent；每段 reasoning_content（思考型模型的推理过程）回调 onReasoning。
// 两个回调均可为 nil。返回的 Message 仅含最终答案 Content 与按 index 组装的 ToolCalls
// （reasoning 是过程性内容，不计入 Content、不回灌历史）。HTTP 非 200 时先读体返回错误。
func (c *Client) ChatStream(ctx context.Context, messages []Message, tools []Tool, onContent, onReasoning func(delta string)) (Message, error) {
	// 请求发出前校验 base_url(收敛 SSRF 面:仅允许带 host 的 http/https)。
	if err := ValidateBaseURL(c.baseURL); err != nil {
		return Message{}, err
	}
	reqBody := streamChatRequest{
		Model:       c.model,
		Messages:    messages,
		Tools:       tools,
		Temperature: 0.2,
		Stream:      true,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return Message{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return Message{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.streamHTTP.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("call llm: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var parsed chatResponse
		if json.Unmarshal(body, &parsed) == nil && parsed.Error != nil && parsed.Error.Message != "" {
			return Message{}, fmt.Errorf("llm error (status %d): %s", resp.StatusCode, c.redact(parsed.Error.Message))
		}
		return Message{}, fmt.Errorf("llm http status %d", resp.StatusCode)
	}

	var content strings.Builder
	// toolCalls 按 index 累积：首个分片带 id+name，后续分片拼接 arguments。
	toolCalls := make(map[int]*ToolCall)

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), streamScannerBuffer)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return Message{}, err
		}
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			// 跳过无法解析的分片（如保活注释），不中断整体流。
			continue
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.ReasoningContent != "" && onReasoning != nil {
				onReasoning(choice.Delta.ReasoningContent)
			}
			if choice.Delta.Content != "" {
				content.WriteString(choice.Delta.Content)
				if onContent != nil {
					onContent(choice.Delta.Content)
				}
			}
			for _, tc := range choice.Delta.ToolCalls {
				item, ok := toolCalls[tc.Index]
				if !ok {
					item = &ToolCall{Type: "function"}
					toolCalls[tc.Index] = item
				}
				if tc.ID != "" {
					item.ID = tc.ID
				}
				if tc.Type != "" {
					item.Type = tc.Type
				}
				if tc.Function.Name != "" {
					item.Function.Name = tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					item.Function.Arguments += tc.Function.Arguments
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return Message{}, fmt.Errorf("read stream: %w", err)
	}

	indexes := make([]int, 0, len(toolCalls))
	for idx := range toolCalls {
		indexes = append(indexes, idx)
	}
	sort.Ints(indexes)
	assembled := make([]ToolCall, 0, len(indexes))
	for _, idx := range indexes {
		assembled = append(assembled, *toolCalls[idx])
	}

	return Message{Role: "assistant", Content: content.String(), ToolCalls: assembled}, nil
}
