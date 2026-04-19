import request from '@/utils/http'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface TaskLogListParams {
  page: number
  page_size: number
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
  task_name: string
  host_id: number
  host_name: string
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
