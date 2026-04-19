import request from '@/utils/http'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface InstallParams {
  db_type: string
  db_host: string
  db_port: number
  db_username: string
  db_password: string
  db_name: string
  db_table_prefix: string
  admin_username: string
  admin_password: string
  confirm_admin_password: string
  admin_email: string
}

export interface InstallStatusResult {
  /** true if the system has already been installed */
  installed: boolean
}

// ── API functions ─────────────────────────────────────────────────────────────

/**
 * GET /api/install/status
 * Returns { code: 0, data: true|false } — data is the installed boolean.
 */
export function fetchInstallStatus() {
  return request.get<boolean>({
    url: '/api/install/status'
  })
}

/**
 * POST /api/install/store  (application/x-www-form-urlencoded)
 * gocron uses c.ShouldBind which reads form fields.
 */
export function fetchInstall(params: InstallParams) {
  const form = new URLSearchParams()
  form.append('db_type', params.db_type)
  form.append('db_host', params.db_host)
  form.append('db_port', String(params.db_port))
  form.append('db_username', params.db_username)
  form.append('db_password', params.db_password)
  form.append('db_name', params.db_name)
  form.append('db_table_prefix', params.db_table_prefix)
  form.append('admin_username', params.admin_username)
  form.append('admin_password', params.admin_password)
  form.append('confirm_admin_password', params.confirm_admin_password)
  form.append('admin_email', params.admin_email)

  return request.post<null>({
    url: '/api/install/store',
    data: form,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}
