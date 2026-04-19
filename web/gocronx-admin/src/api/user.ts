import request from '@/utils/http'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface UserListItem {
  id: number
  name: string
  email: string
  /** 0 = normal user, 1 = admin */
  is_admin: number
  /** 0 = disabled, 1 = enabled */
  status: number
  created: string
}

export interface UserListParams {
  page: number
  page_size: number
}

export interface UserStoreParams {
  id?: number
  name: string
  email: string
  password?: string
  is_admin?: number
}

export interface EditPasswordParams {
  id: number
  new_password: string
  confirm_new_password: string
}

export interface EditMyPasswordParams {
  old_password: string
  new_password: string
  confirm_new_password: string
}

// ── API functions ─────────────────────────────────────────────────────────────

/**
 * GET /api/user  →  { total, data: UserListItem[] }
 */
export function fetchUserList(params: UserListParams) {
  return request.get<{ total: number; data: UserListItem[] }>({
    url: '/api/user',
    params
  })
}

/**
 * GET /api/user/:id  →  UserListItem
 */
export function fetchUserDetail(id: number) {
  return request.get<UserListItem>({
    url: `/api/user/${id}`
  })
}

/**
 * POST /api/user/store  (create or update)
 * Uses application/x-www-form-urlencoded — gocron uses c.PostForm()
 */
export function fetchUserStore(params: UserStoreParams) {
  const form = new URLSearchParams()
  if (params.id) form.append('id', String(params.id))
  form.append('name', params.name)
  form.append('email', params.email)
  if (params.password) form.append('password', params.password)
  if (params.is_admin !== undefined) form.append('is_admin', String(params.is_admin))

  return request.post<null>({
    url: '/api/user/store',
    data: form,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}

/**
 * POST /api/user/remove/:id
 */
export function fetchUserRemove(id: number) {
  return request.post<null>({
    url: `/api/user/remove/${id}`
  })
}

/**
 * POST /api/user/enable/:id
 */
export function fetchUserEnable(id: number) {
  return request.post<null>({
    url: `/api/user/enable/${id}`
  })
}

/**
 * POST /api/user/disable/:id
 */
export function fetchUserDisable(id: number) {
  return request.post<null>({
    url: `/api/user/disable/${id}`
  })
}

/**
 * POST /api/user/editPassword/:id  (admin changes another user's password)
 */
export function fetchUserEditPassword(params: EditPasswordParams) {
  const form = new URLSearchParams()
  form.append('new_password', params.new_password)
  form.append('confirm_new_password', params.confirm_new_password)

  return request.post<null>({
    url: `/api/user/editPassword/${params.id}`,
    data: form,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}

/**
 * POST /api/user/editMyPassword  (current user changes own password)
 */
export function fetchUserEditMyPassword(params: EditMyPasswordParams) {
  const form = new URLSearchParams()
  form.append('old_password', params.old_password)
  form.append('new_password', params.new_password)
  form.append('confirm_new_password', params.confirm_new_password)

  return request.post<null>({
    url: '/api/user/editMyPassword',
    data: form,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}

/**
 * Alias for fetchUserEditMyPassword — task spec compatibility
 */
export function updateMyPassword(params: EditMyPasswordParams) {
  return fetchUserEditMyPassword(params)
}

// ── 2FA ───────────────────────────────────────────────────────────────────────

/**
 * GET /api/user/2fa/status  →  { enabled: boolean }
 */
export function get2FAStatus() {
  return request.get<{ enabled: boolean }>({ url: '/api/user/2fa/status' })
}

/**
 * GET /api/user/2fa/setup  →  { qr_code: string, secret: string }
 */
export function setup2FA() {
  return request.get<{ qr_code: string; secret: string }>({ url: '/api/user/2fa/setup' })
}

/**
 * POST /api/user/2fa/enable  body: secret + code (form-urlencoded)
 */
export function enable2FA(secret: string, code: string) {
  const body = new URLSearchParams()
  body.append('secret', secret)
  body.append('code', code)
  return request.post<null>({
    url: '/api/user/2fa/enable',
    data: body,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}

/**
 * POST /api/user/2fa/disable  body: code (form-urlencoded)
 */
export function disable2FA(code: string) {
  const body = new URLSearchParams()
  body.append('code', code)
  return request.post<null>({
    url: '/api/user/2fa/disable',
    data: body,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}
