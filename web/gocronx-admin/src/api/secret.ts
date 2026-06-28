import request from '@/utils/http'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface SecretItem {
  id: number
  name: string
  remark: string
  created_at: string
  updated_at: string
}

export interface SecretStoreParams {
  id?: number
  name: string
  /** Plaintext value; on edit, leave empty to keep the existing value. */
  value?: string
  remark?: string
}

// ── API functions ─────────────────────────────────────────────────────────────

/**
 * GET /api/secret  →  { total, data: SecretItem[] }  (value is never returned)
 */
export function fetchSecretList() {
  return request.get<{ total: number; data: SecretItem[] }>({
    url: '/api/secret'
  })
}

/**
 * POST /api/secret/store  (create or update)
 * Uses application/x-www-form-urlencoded — backend reads c.PostForm().
 */
export function storeSecret(params: SecretStoreParams) {
  const form = new URLSearchParams()
  if (params.id) form.append('id', String(params.id))
  form.append('name', params.name)
  if (params.value) form.append('value', params.value)
  if (params.remark !== undefined) form.append('remark', params.remark)

  return request.post<null>({
    url: '/api/secret/store',
    data: form,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}

/**
 * POST /api/secret/remove/:id
 */
export function removeSecret(id: number) {
  return request.post<null>({
    url: `/api/secret/remove/${id}`
  })
}
