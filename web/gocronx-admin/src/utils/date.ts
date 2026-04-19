/**
 * Format a datetime string (ISO 8601 / RFC3339 or anything Date.parse accepts)
 * into `YYYY-MM-DD HH:mm:ss` in local time.
 *
 * gocron backend serializes time.Time as RFC3339 with TZ offset
 * (e.g. "2026-04-19T10:00:00+08:00"), which is ugly to display raw.
 */
export function formatDateTime(input: string | number | Date | null | undefined): string {
  if (!input) return ''
  const d = input instanceof Date ? input : new Date(input)
  if (isNaN(d.getTime())) return typeof input === 'string' ? input : ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return (
    `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ` +
    `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  )
}
