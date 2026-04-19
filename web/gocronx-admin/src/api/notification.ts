import request from '@/utils/http'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface MailUser {
  id: number
  username: string
  email: string
}

export interface SlackChannel {
  id: number
  name: string
}

export interface WebhookUrl {
  id: number
  name: string
  url: string
}

export interface MailConfig {
  host: string
  port: number
  user: string
  password: string
  template: string
  mail_users: MailUser[]
}

export interface SlackConfig {
  url: string
  template: string
  channels: SlackChannel[]
}

export interface WebhookConfig {
  template: string
  webhook_urls: WebhookUrl[]
}

// ── Mail ──────────────────────────────────────────────────────────────────────

/**
 * GET /api/system/mail  →  full mail config including mail_users
 */
export function fetchMail() {
  return request.get<MailConfig>({
    url: '/api/system/mail'
  })
}

/**
 * POST /api/system/mail/update  — update SMTP settings
 */
export function updateMail(params: {
  host: string
  port: number
  user: string
  password: string
  template: string
}) {
  const form = new URLSearchParams()
  form.append('host', params.host)
  form.append('port', String(params.port))
  form.append('user', params.user)
  form.append('password', params.password)
  form.append('template', params.template)

  return request.post<null>({
    url: '/api/system/mail/update',
    data: form,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}

/**
 * POST /api/system/mail/user  — add a mail recipient
 */
export function createMailUser(params: { username: string; email: string }) {
  const form = new URLSearchParams()
  form.append('username', params.username)
  form.append('email', params.email)

  return request.post<null>({
    url: '/api/system/mail/user',
    data: form,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}

/**
 * POST /api/system/mail/user/remove/:id
 */
export function removeMailUser(id: number) {
  return request.post<null>({
    url: `/api/system/mail/user/remove/${id}`
  })
}

// ── Slack ─────────────────────────────────────────────────────────────────────

/**
 * GET /api/system/slack  →  full slack config including channels
 */
export function fetchSlack() {
  return request.get<SlackConfig>({
    url: '/api/system/slack'
  })
}

/**
 * POST /api/system/slack/update  — update Slack webhook URL and template
 */
export function updateSlack(params: { url: string; template: string }) {
  const form = new URLSearchParams()
  form.append('url', params.url)
  form.append('template', params.template)

  return request.post<null>({
    url: '/api/system/slack/update',
    data: form,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}

/**
 * POST /api/system/slack/channel  — add a Slack channel
 */
export function createSlackChannel(channel: string) {
  const form = new URLSearchParams()
  form.append('channel', channel)

  return request.post<null>({
    url: '/api/system/slack/channel',
    data: form,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}

/**
 * POST /api/system/slack/channel/remove/:id
 */
export function removeSlackChannel(id: number) {
  return request.post<null>({
    url: `/api/system/slack/channel/remove/${id}`
  })
}

// ── Webhook ───────────────────────────────────────────────────────────────────

/**
 * GET /api/system/webhook  →  full webhook config including webhook_urls
 */
export function fetchWebhook() {
  return request.get<WebhookConfig>({
    url: '/api/system/webhook'
  })
}

/**
 * POST /api/system/webhook/update  — update webhook template
 */
export function updateWebhook(params: { template: string }) {
  const form = new URLSearchParams()
  form.append('template', params.template)

  return request.post<null>({
    url: '/api/system/webhook/update',
    data: form,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}

/**
 * POST /api/system/webhook/url  — add a webhook URL
 */
export function createWebhookUrl(params: { name: string; url: string }) {
  const form = new URLSearchParams()
  form.append('name', params.name)
  form.append('url', params.url)

  return request.post<null>({
    url: '/api/system/webhook/url',
    data: form,
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
  })
}

/**
 * POST /api/system/webhook/url/remove/:id
 */
export function removeWebhookUrl(id: number) {
  return request.post<null>({
    url: `/api/system/webhook/url/remove/${id}`
  })
}
