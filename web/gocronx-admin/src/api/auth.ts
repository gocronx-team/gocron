import request from '@/utils/http'

/**
 * 登录
 *
 * gocron backend: POST /api/user/login
 * Content-Type: application/x-www-form-urlencoded (c.PostForm() on the Go side)
 *
 * Success: { code: 0, message: "success", data: { token, uid, username, is_admin } }
 * 2FA required: { code: 0, message: "2fa_code_required", data: { require_2fa: true } }
 * Error: { code: <non-0>, message: "...", data: null }
 */
export function fetchLogin(params: Api.Auth.LoginParams) {
  const formBody = new URLSearchParams()
  formBody.append('username', params.username)
  formBody.append('password', params.password)
  if (params.two_factor_code) {
    formBody.append('two_factor_code', params.two_factor_code)
  }

  // We bypass the normal request helper here so we can inspect the raw
  // response body (including the 2FA intermediate state where code=0 but
  // data.require_2fa is true).  The response interceptor in http/index.ts
  // already normalises code 0 → pass-through, so we just return the raw
  // axios call and let the login page handle both outcomes.
  return request.post<Api.Auth.LoginResponse | Api.Auth.Login2FARequired>({
    url: '/api/user/login',
    data: formBody,
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    // Don't show automatic error toast — the login page renders errors inline
    showErrorMessage: false
  })
}

/**
 * 获取用户信息
 *
 * gocron has NO /api/user/info endpoint.
 * User info (uid / username / is_admin) is returned at login time and stored
 * in the Pinia user store.  This function reads it back from localStorage
 * (via the persisted store) so the route guard can call it on every page load
 * without making any HTTP request.
 *
 * Falls back to decoding the JWT claims directly if the store is empty.
 */
export function fetchGetUserInfo(): Promise<Api.Auth.UserInfo> {
  return Promise.resolve(resolveUserInfo())
}

function resolveUserInfo(): Api.Auth.UserInfo {
  // 1. Try to read from the persisted Pinia store (localStorage key managed
  //    by StorageKeyManager — the actual key is "sys-v{version}-user").
  //    We use a lazy import to avoid circular deps with the store module.
  try {
    // Find the right localStorage key: iterate all keys that look like user stores
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i)
      if (!key) continue
      if (!key.includes('user')) continue
      const raw = localStorage.getItem(key)
      if (!raw) continue
      const parsed = JSON.parse(raw)
      const info = parsed?.info
      if (info && info.userId) {
        return info as Api.Auth.UserInfo
      }
    }
  } catch {
    // fall through to JWT decode
  }

  // 2. Fall back: decode the JWT token claims stored in the same localStorage
  //    key under `accessToken`.
  try {
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i)
      if (!key) continue
      if (!key.includes('user')) continue
      const raw = localStorage.getItem(key)
      if (!raw) continue
      const parsed = JSON.parse(raw)
      const token: string = parsed?.accessToken || ''
      if (!token) continue

      const claims = decodeJwtClaims(token)
      if (claims) {
        return buildUserInfoFromClaims(claims)
      }
    }
  } catch {
    // fall through
  }

  // 3. Nothing found — return empty shell; the route guard will redirect to /login
  return { userId: 0, userName: '', isAdmin: 0, roles: [], buttons: [], email: '' }
}

/**
 * Decode the payload section of a JWT without verifying the signature.
 * The signature is validated server-side on every authenticated request.
 * gocron JWT claims: { uid, username, is_admin, exp, iat, issuer }
 */
function decodeJwtClaims(token: string): Record<string, any> | null {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) return null
    // base64url → base64 → string
    const payload = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const padded =
      payload + '=='.slice((payload.length + 3) % 4 > 0 ? ((payload.length + 3) % 4) - 1 : 2)
    const decoded = atob(padded)
    return JSON.parse(decoded)
  } catch {
    return null
  }
}

/**
 * Map gocron JWT claims to the UserInfo shape the template expects.
 */
function buildUserInfoFromClaims(claims: Record<string, any>): Api.Auth.UserInfo {
  const isAdmin: number = typeof claims.is_admin === 'number' ? claims.is_admin : 0
  return {
    userId: typeof claims.uid === 'number' ? claims.uid : 0,
    userName: typeof claims.username === 'string' ? claims.username : '',
    isAdmin,
    roles: isAdmin === 1 ? ['R_SUPER', 'R_ADMIN'] : ['R_USER'],
    buttons: [],
    email: ''
  }
}
