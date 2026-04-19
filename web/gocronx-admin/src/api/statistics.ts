import request from '@/utils/http'

/**
 * A single day's execution data returned by the backend in last_7_days array.
 * date: "YYYY-MM-DD", total/success/failed: integer counts.
 */
export interface DayStats {
  date: string
  total: number
  success: number
  failed: number
}

/**
 * Overview response shape from GET /api/statistics/overview.
 * Mirrors the old frontend's stats.value mapping in web/vue/src/pages/statistics/index.vue.
 */
export interface StatisticsOverview {
  /** Total registered tasks */
  total_tasks: number
  /** Last-7-days rows, newest first (DESC) */
  last_7_days: DayStats[]
}

export function fetchStatisticsOverview() {
  return request.get<StatisticsOverview>({ url: '/api/statistics/overview' })
}
