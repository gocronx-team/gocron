import request from '@/utils/http'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface HostListParams {
  page: number
  page_size: number
  id?: number | string
  name?: string
}

export interface HostItem {
  id: number
  name: string
  alias: string
  port: number
  remark?: string
  created: string
}

export interface HostStoreParams {
  id?: number
  name: string
  alias?: string
  port: number
  remark?: string
}

export interface AgentTokenResult {
  token: string
  expires_at: string
  install_cmd: string
}

// ── API functions ─────────────────────────────────────────────────────────────

/**
 * GET /api/host  →  { total, data: HostItem[] }
 */
export function fetchHostList(params: HostListParams) {
  return request.get<{ total: number; data: HostItem[] }>({
    url: '/api/host',
    params
  })
}

/**
 * GET /api/host/:id  →  HostItem
 */
export function fetchHostDetail(id: number) {
  return request.get<HostItem>({
    url: `/api/host/${id}`
  })
}

/**
 * POST /api/host/store  (create or update)
 * Uses application/x-www-form-urlencoded — gocron uses c.PostForm()
 */
export function saveHost(params: HostStoreParams) {
  const form = new URLSearchParams()
  if (params.id) form.append('id', String(params.id))
  form.append('name', params.name)
  if (params.alias !== undefined) form.append('alias', params.alias)
  form.append('port', String(params.port))
  if (params.remark !== undefined) form.append('remark', params.remark)

  return request.post<null>({
    url: '/api/host/store',
    data: form,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}

/**
 * GET /api/host/ping/:id  →  ping result
 */
export function pingHost(id: number) {
  return request.get<any>({
    url: `/api/host/ping/${id}`
  })
}

/**
 * POST /api/host/remove/:id
 */
export function removeHost(id: number) {
  return request.post<null>({
    url: `/api/host/remove/${id}`
  })
}

/**
 * POST /api/agent/generate-token  →  { token, expires_at, install_cmd }
 */
export function generateAgentToken() {
  return request.post<AgentTokenResult>({
    url: '/api/agent/generate-token'
  })
}
