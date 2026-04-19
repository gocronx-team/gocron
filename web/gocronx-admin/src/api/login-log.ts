import request from '@/utils/http'

export interface LoginLogItem {
  id: number
  username: string
  ip: string
  created: string
}

export interface LoginLogListResult {
  total: number
  data: LoginLogItem[]
}

export function fetchLoginLogList(params: { page: number; page_size: number }) {
  return request.get<LoginLogListResult>({
    url: '/api/system/login-log',
    params
  })
}
