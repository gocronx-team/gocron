import request from '@/utils/http'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface TemplateListParams {
  page: number
  page_size: number
  name?: string
}

export interface TemplateListItem {
  id: number
  name: string
  spec: string
  protocol: number
  command: string
  description?: string
  category?: string
  tag?: string
  http_method?: number
  http_body?: string
  http_headers?: string
  success_pattern?: string
  timeout?: number
  multi?: number
  retry_times?: number
  retry_interval?: number
  timezone?: string
  notify_status?: number
  notify_type?: number
  notify_keyword?: string
  log_retention_days?: number
  is_builtin?: number
  created_at?: string
  updated_at?: string
  created_by?: string
}

export interface TemplateStoreParams {
  id?: number
  name: string
  spec: string
  protocol: number
  command: string
  description?: string
  category?: string
  tag?: string
  http_method?: number
  http_body?: string
  http_headers?: string
  success_pattern?: string
  timeout?: number
  multi?: number
  retry_times?: number
  retry_interval?: number
  timezone?: string
  notify_status?: number
  notify_type?: number
  notify_keyword?: string
  log_retention_days?: number
}

// ── API functions ─────────────────────────────────────────────────────────────

/**
 * GET /api/template  →  { total, data: TemplateListItem[] }
 */
export function fetchTemplateList(params: TemplateListParams) {
  return request.get<{ total: number; data: TemplateListItem[] }>({
    url: '/api/template',
    params
  })
}

/**
 * GET /api/template/:id  →  TemplateListItem
 */
export function fetchTemplateDetail(id: number) {
  return request.get<TemplateListItem>({
    url: `/api/template/${id}`
  })
}

/**
 * POST /api/template/store  (create or update)
 * Uses application/x-www-form-urlencoded — gocron uses c.PostForm()
 */
export function fetchTemplateStore(params: TemplateStoreParams) {
  const form = new URLSearchParams()
  if (params.id) form.append('id', String(params.id))
  form.append('name', params.name)
  form.append('spec', params.spec)
  form.append('protocol', String(params.protocol))
  form.append('command', params.command)
  if (params.description !== undefined) form.append('description', params.description)
  if (params.category !== undefined) form.append('category', params.category)
  if (params.tag !== undefined) form.append('tag', params.tag)
  if (params.http_method !== undefined) form.append('http_method', String(params.http_method))
  if (params.http_body !== undefined) form.append('http_body', params.http_body)
  if (params.http_headers !== undefined) form.append('http_headers', params.http_headers)
  if (params.success_pattern !== undefined) form.append('success_pattern', params.success_pattern)
  if (params.timeout !== undefined) form.append('timeout', String(params.timeout))
  if (params.multi !== undefined) form.append('multi', String(params.multi))
  if (params.retry_times !== undefined) form.append('retry_times', String(params.retry_times))
  if (params.retry_interval !== undefined)
    form.append('retry_interval', String(params.retry_interval))
  if (params.timezone !== undefined) form.append('timezone', params.timezone)
  if (params.notify_status !== undefined) form.append('notify_status', String(params.notify_status))
  if (params.notify_type !== undefined) form.append('notify_type', String(params.notify_type))
  if (params.notify_keyword !== undefined) form.append('notify_keyword', params.notify_keyword)
  if (params.log_retention_days !== undefined)
    form.append('log_retention_days', String(params.log_retention_days))

  return request.post<null>({
    url: '/api/template/store',
    data: form,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}

/**
 * POST /api/template/remove/:id
 */
export function fetchTemplateRemove(id: number) {
  return request.post<null>({
    url: `/api/template/remove/${id}`
  })
}

/**
 * POST /api/template/save-from-task
 * Clone the current task's scheduling/command fields into a new template.
 */
export function fetchTemplateSaveFromTask(params: {
  task_id: number
  name: string
  description?: string
  category?: string
}) {
  const form = new URLSearchParams()
  form.append('task_id', String(params.task_id))
  form.append('name', params.name)
  if (params.description !== undefined) form.append('description', params.description)
  if (params.category !== undefined) form.append('category', params.category)

  return request.post<null>({
    url: '/api/template/save-from-task',
    data: form,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}
