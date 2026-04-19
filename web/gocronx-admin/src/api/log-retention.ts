import request from '@/utils/http'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface LogRetentionData {
  /** Days to retain task logs (0 = unlimited) */
  days: number
  /** Daily cleanup time in HH:mm format */
  cleanup_time: string
  /** Maximum log file size in MB (0 = unlimited) */
  file_size_limit: number
}

// ── API functions ─────────────────────────────────────────────────────────────

/**
 * GET /api/system/log-retention
 * Returns current log retention settings.
 */
export function fetchLogRetention() {
  return request.get<LogRetentionData>({
    url: '/api/system/log-retention'
  })
}

/**
 * POST /api/system/log-retention
 * Updates log retention settings.
 */
export function updateLogRetention(params: LogRetentionData) {
  return request.post<null>({
    url: '/api/system/log-retention',
    data: params
  })
}
