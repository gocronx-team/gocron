import request from '@/utils/http'

export interface AuditListParams {
  page: number
  page_size: number
  module?: string
  action?: string
  username?: string
  start_date?: string
  end_date?: string
}

export interface AuditListItem {
  id: number
  username: string
  module: string
  action: string
  target_id: number
  target_name: string
  detail: string
  ip: string
  created: string
}

export function fetchAuditList(params: AuditListParams) {
  return request.get<{ total: number; data: AuditListItem[] }>({
    url: '/api/audit',
    params
  })
}
