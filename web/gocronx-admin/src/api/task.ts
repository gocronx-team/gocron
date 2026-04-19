import request from '@/utils/http'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface TaskListParams {
  page: number
  page_size: number
  name?: string
  host_id?: number | string
  protocol?: number | string
  status?: number | string
  tag?: string
}

export interface TaskHostRef {
  host_id: number
  alias: string
  name: string
  port: number
}

export interface TaskListItem {
  id: number
  name: string
  spec: string
  protocol: number
  http_method: number
  http_body?: string
  http_headers?: string
  success_pattern?: string
  command: string
  timeout: number
  multi: number
  retry_times: number
  retry_interval: number
  remark: string
  tag: string
  status: number
  level: number
  dependency_status?: number
  dependency_task_id?: string
  notify_status?: number
  notify_type?: number
  notify_keyword?: string
  notify_receiver_id?: string
  log_retention_days?: number
  next_run_time: string
  created: string
  hosts: TaskHostRef[]
}

export interface TaskStoreParams {
  id?: number
  name: string
  spec: string
  protocol: number
  command: string
  timeout?: number
  multi?: number
  retry_times?: number
  retry_interval?: number
  remark?: string
  tag?: string
  host_id?: string | number | number[]
  http_method?: number
  http_body?: string
  http_headers?: string
  success_pattern?: string
  level?: number
  dependency_status?: number
  dependency_task_id?: string
  notify_status?: number
  notify_type?: number
  notify_keyword?: string
  notify_receiver_id?: string
  log_retention_days?: number
}

// ── API functions ─────────────────────────────────────────────────────────────

/**
 * GET /api/task  →  { total, data: TaskListItem[] }
 */
export function fetchTaskList(params: TaskListParams) {
  return request.get<{ total: number; data: TaskListItem[] }>({
    url: '/api/task',
    params
  })
}

/**
 * GET /api/task/:id  →  TaskListItem
 */
export function fetchTaskDetail(id: number) {
  return request.get<TaskListItem>({
    url: `/api/task/${id}`
  })
}

/**
 * POST /api/task/store  (create or update)
 */
export function fetchTaskStore(data: TaskStoreParams) {
  const form = new URLSearchParams()
  Object.entries(data).forEach(([key, val]) => {
    if (val !== undefined && val !== null) {
      if (Array.isArray(val)) {
        val.forEach((v) => form.append(key, String(v)))
      } else {
        form.append(key, String(val))
      }
    }
  })
  return request.post<null>({
    url: '/api/task/store',
    data: form,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}

/**
 * POST /api/task/remove/:id
 */
export function fetchTaskRemove(id: number) {
  return request.post<null>({
    url: `/api/task/remove/${id}`
  })
}

/**
 * POST /api/task/enable/:id
 */
export function fetchTaskEnable(id: number) {
  return request.post<null>({
    url: `/api/task/enable/${id}`
  })
}

/**
 * POST /api/task/disable/:id
 */
export function fetchTaskDisable(id: number) {
  return request.post<null>({
    url: `/api/task/disable/${id}`
  })
}

/**
 * GET /api/task/run/:id  — trigger immediate run
 */
export function fetchTaskRunOnce(id: number) {
  return request.get<null>({
    url: `/api/task/run/${id}`,
    params: { _t: Date.now() }
  })
}

/**
 * GET /api/task/tags  →  string[]
 */
export function fetchTaskTags() {
  return request.get<string[]>({
    url: '/api/task/tags'
  })
}

/**
 * POST /api/task/cron-preview  →  { next_times: string[], heatmap: ... }
 */
export function fetchCronPreview(params: { spec: string; timezone?: string; count?: number }) {
  return request.post<{ next_times: string[] }>({
    url: '/api/task/cron-preview',
    data: params
  })
}

/**
 * POST /api/task/batch-enable  →  null
 */
export function fetchBatchEnable(ids: number[]) {
  return request.post<null>({
    url: '/api/task/batch-enable',
    data: { ids }
  })
}

/**
 * POST /api/task/batch-disable  →  null
 */
export function fetchBatchDisable(ids: number[]) {
  return request.post<null>({
    url: '/api/task/batch-disable',
    data: { ids }
  })
}

/**
 * POST /api/task/batch-remove  →  null
 */
export function fetchBatchRemove(ids: number[]) {
  return request.post<null>({
    url: '/api/task/batch-remove',
    data: { ids }
  })
}
