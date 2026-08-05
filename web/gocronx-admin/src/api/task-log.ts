import request from '@/utils/http'
import { useUserStore } from '@/store/modules/user'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface TaskLogListParams {
  page: number
  page_size: number
  keyword?: string
  task_id?: number | string
  protocol?: number | string
  status?: number | string
  host_id?: number | string
  start_date?: string
  end_date?: string
}

export interface TaskLogListItem {
  id: number
  task_id: number
  /** Task name (backend field is `name`) */
  name: string
  host_id: number
  /** Raw command string (may contain HTML entities from old encoding) */
  command: string
  protocol: number
  status: number
  /** RFC3339 start time */
  start_time: string
  /** RFC3339 end time */
  end_time: string
  /** Hostname of the execution node */
  hostname: string
  /** Execution output text */
  output: string
  /** Execution result text (some versions use this field) */
  result: string
  /** Elapsed seconds */
  total_time: number
  retry_times: number
  spec: string
}

// ── API functions ─────────────────────────────────────────────────────────────

/**
 * GET /api/task/log  →  { total, data: TaskLogListItem[] }
 */
export function fetchTaskLogList(params: TaskLogListParams) {
  return request.get<{ total: number; data: TaskLogListItem[] }>({
    url: '/api/task/log',
    params
  })
}

/**
 * POST /api/task/log/clear  — clear all task logs (admin action)
 */
export function fetchTaskLogClear() {
  return request.post<null>({
    url: '/api/task/log/clear'
  })
}

/**
 * POST /api/task/log/stop  — terminate a running job
 */
export function fetchTaskLogStop(id: number, taskId: number) {
  return request.post<null>({
    url: '/api/task/log/stop',
    data: { id, task_id: taskId }
  })
}

export interface TaskLogStreamEvent {
  content: string
  seq: number
  status: number
  reset?: boolean
}

export async function streamTaskLog(
  id: number,
  seq: number,
  handlers: {
    onLog: (event: TaskLogStreamEvent) => void
    onDone: (event: TaskLogStreamEvent) => void
    onError: (message: string) => void
  },
  signal?: AbortSignal
): Promise<void> {
  const res = await fetch(`/api/task/log/${id}/stream?seq=${seq}`, {
    headers: {
      'Auth-Token': useUserStore().accessToken,
      'Accept-Language': useUserStore().language
    },
    signal
  })
  if (!res.ok || !res.body) throw new Error(`HTTP ${res.status}`)

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let eventType = ''
  let data = ''

  const dispatch = () => {
    if (!eventType || !data) return
    try {
      const event = JSON.parse(data) as TaskLogStreamEvent
      if (eventType === 'log') handlers.onLog(event)
      if (eventType === 'done') handlers.onDone(event)
    } catch {
      handlers.onError('invalid stream event')
    }
    eventType = ''
    data = ''
  }

  while (true) {
    const { value, done } = await reader.read()
    buffer += decoder.decode(value, { stream: !done })
    let newline = buffer.indexOf('\n')
    while (newline >= 0) {
      const line = buffer.slice(0, newline).replace(/\r$/, '')
      buffer = buffer.slice(newline + 1)
      if (line === '') dispatch()
      else if (line.startsWith('event:')) eventType = line.slice(6).trim()
      else if (line.startsWith('data:')) data += line.slice(5).trimStart()
      newline = buffer.indexOf('\n')
    }
    if (done) {
      dispatch()
      return
    }
  }
}
