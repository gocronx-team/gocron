import request from '@/utils/http'
import { useUserStore } from '@/store/modules/user'

// 把当前界面语言（zh/en）作为 Accept-Language 传给后端，
// 让 AI 的输出语言与界面切换保持一致（后端 GetLocale 据此选 prompt 语言）。
function langHeader() {
  return { 'Accept-Language': useUserStore().language }
}

// ── LLM config ──────────────────────────────────────────────────────────────

export interface LLMConfig {
  enable: boolean
  base_url: string
  model: string
  api_key_set: boolean
}

export interface LLMConfigUpdate {
  enable: boolean
  base_url: string
  api_key: string // 留空表示不修改
  model: string
}

/** GET /api/system/llm  →  LLMConfig (never returns the key) */
export function fetchLLMConfig() {
  return request.get<LLMConfig>({ url: '/api/system/llm' })
}

/** POST /api/system/llm/update */
export function updateLLMConfig(data: LLMConfigUpdate) {
  return request.post<null>({ url: '/api/system/llm/update', data })
}

/**
 * POST /api/system/llm/test — 用当前表单配置发一句极短对话验证连通性。
 * api_key 留空则用已保存的值。失败时后端把 provider 真实错误带在 message 里(已脱敏)。
 */
export function testLLMConfig(data: LLMConfigUpdate) {
  return request.post<null>({ url: '/api/system/llm/test', data, timeout: AI_TIMEOUT })
}

// LLM 推理（尤其本地大模型）较慢，这两个接口单独放宽超时，避免前端 15s 默认超时提前断开。
const AI_TIMEOUT = 120000

// ── NL → cron ─────────────────────────────────────────────────────────────────

export interface NlToCronResult {
  spec: string
  preview: {
    valid: boolean
    error?: string
    next_runs?: { text: string }[]
  }
}

/** POST /api/task/nl-to-cron */
export function nlToCron(text: string, timezone?: string) {
  return request.post<NlToCronResult>({
    url: '/api/task/nl-to-cron',
    data: { text, timezone: timezone || '' },
    timeout: AI_TIMEOUT,
    headers: langHeader()
  })
}

// ── Failure log diagnosis ─────────────────────────────────────────────────────

export interface DiagnoseResult {
  root_cause: string
  suggestions: string[]
}

/** POST /api/task/log/diagnose/:id */
export function diagnoseLog(id: number) {
  return request.post<DiagnoseResult>({
    url: `/api/task/log/diagnose/${id}`,
    timeout: AI_TIMEOUT,
    headers: langHeader()
  })
}

// ── Ops chat ────────────────────────────────────────────────────────────────

export type AiChatMessage = {
  role: 'user' | 'assistant'
  content: string
}

export interface ToolCall {
  id: string
  name: string
  arguments: string
}

export interface ToolResult {
  id: string
  name: string
  ok: boolean
}

/** AI 建议新建的任务预览(用户可在前端编辑后再确认创建)。 */
export interface CreateProposal {
  name: string
  spec: string
  protocol: number
  command: string
  http_method: number
  http_body: string
  http_headers: string
  success_pattern: string
  timeout: number
  multi: number
  retry_times: number
  retry_interval: number
  tag: string
  remark: string
  log_retention_days: number
}

export interface StreamAiChatHandlers {
  onMessage: (delta: string) => void
  onReasoning: (delta: string) => void
  onToolCall: (t: ToolCall) => void
  onToolResult: (t: ToolResult) => void
  onConfirmRequired: (t: { taskId: number; taskName: string }) => void
  onCreateProposal: (p: CreateProposal) => void
  onError: (msg: string) => void
  onDone: () => void
}

/** POST /api/ai/run-task/:id — 用户在聊天里确认后真正执行任务（仅管理员，后端写审计）。 */
export function confirmRunTask(id: number) {
  return request.post<null>({ url: `/api/ai/run-task/${id}` })
}

/**
 * POST /api/ai/chat — stream the assistant reply via Server-Sent Events.
 *
 * 用原生 fetch + ReadableStream 读取 `text/event-stream`。每个事件形如
 * `event: <type>\ndata: <json>\n\n`，我们按行解析并在空行处 flush 一个完整事件，
 * 跨 chunk 的半截事件由 buffer 暂存，只有读到完整行才消费，避免 JSON 解析到一半。
 *
 * gocron 用 `Auth-Token` 头携带令牌（不是 Authorization: Bearer），值取自 user store。
 */
export async function streamAiChat(
  messages: AiChatMessage[],
  handlers: StreamAiChatHandlers,
  signal?: AbortSignal
): Promise<void> {
  let done = false
  const finish = (): void => {
    if (done) return
    done = true
    handlers.onDone()
  }

  let res: Response
  try {
    res = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Auth-Token': useUserStore().accessToken,
        'Accept-Language': useUserStore().language
      },
      body: JSON.stringify({ messages }),
      signal
    })
  } catch (e) {
    if (signal?.aborted) {
      finish()
      return
    }
    handlers.onError((e as Error)?.message || 'network error')
    finish()
    return
  }

  if (!res.ok || !res.body) {
    handlers.onError(`HTTP ${res.status}`)
    finish()
    return
  }

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let eventType = ''
  let dataLines: string[] = []

  const dispatch = (): void => {
    if (!eventType || dataLines.length === 0) {
      eventType = ''
      dataLines = []
      return
    }
    let data: Record<string, unknown> = {}
    try {
      data = JSON.parse(dataLines.join('\n'))
    } catch {
      eventType = ''
      dataLines = []
      return
    }
    switch (eventType) {
      case 'message':
        handlers.onMessage(String(data.content ?? ''))
        break
      case 'reasoning':
        handlers.onReasoning(String(data.content ?? ''))
        break
      case 'tool_call':
        handlers.onToolCall({
          id: String(data.id ?? ''),
          name: String(data.name ?? ''),
          arguments: String(data.arguments ?? '')
        })
        break
      case 'tool_result':
        handlers.onToolResult({
          id: String(data.id ?? ''),
          name: String(data.name ?? ''),
          ok: data.ok === true
        })
        break
      case 'confirm_required':
        handlers.onConfirmRequired({
          taskId: Number(data.task_id ?? 0),
          taskName: String(data.task_name ?? '')
        })
        break
      case 'create_proposal':
        handlers.onCreateProposal({
          name: String(data.name ?? ''),
          spec: String(data.spec ?? ''),
          protocol: Number(data.protocol ?? 1),
          command: String(data.command ?? ''),
          http_method: Number(data.http_method ?? 1),
          http_body: String(data.http_body ?? ''),
          http_headers: String(data.http_headers ?? ''),
          success_pattern: String(data.success_pattern ?? ''),
          timeout: Number(data.timeout ?? 0),
          multi: Number(data.multi ?? 0),
          retry_times: Number(data.retry_times ?? 0),
          retry_interval: Number(data.retry_interval ?? 0),
          tag: String(data.tag ?? ''),
          remark: String(data.remark ?? ''),
          log_retention_days: Number(data.log_retention_days ?? 0)
        })
        break
      case 'error':
        handlers.onError(String(data.message ?? 'error'))
        break
      case 'done':
        finish()
        break
    }
    eventType = ''
    dataLines = []
  }

  const processLine = (line: string): void => {
    if (line.startsWith('event:')) {
      eventType = line.slice(6).trim()
    } else if (line.startsWith('data:')) {
      dataLines.push(line.slice(5).replace(/^ /, ''))
    } else if (line === '') {
      dispatch()
    }
  }

  try {
    for (;;) {
      const { done: streamDone, value } = await reader.read()
      if (streamDone) break
      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() ?? ''
      for (const line of lines) processLine(line.replace(/\r$/, ''))
    }
    buffer += decoder.decode()
    for (const line of buffer.split('\n')) processLine(line.replace(/\r$/, ''))
    dispatch()
  } catch (e) {
    if (!signal?.aborted) handlers.onError((e as Error)?.message || 'stream error')
  } finally {
    finish()
  }
}
