<template>
  <div class="tw-p-6 tw-space-y-6">
    <!-- Page header -->
    <div class="tw-flex tw-items-center tw-justify-between">
      <div>
        <h1 class="tw-text-2xl tw-font-bold tw-tracking-tight">
          {{ t('statistics.title') }}
        </h1>
        <p class="tw-text-sm tw-text-muted-foreground tw-mt-1">
          {{ t('statistics.last7DaysTrend') }}
        </p>
      </div>
      <Button
        variant="outline"
        size="sm"
        @click="refresh"
      >
        <RefreshCw class="tw-size-4 tw-mr-2" />
        {{ t('common.refresh') }}
      </Button>
    </div>

    <!-- KPI Cards -->
    <div class="tw-grid tw-grid-cols-1 sm:tw-grid-cols-2 lg:tw-grid-cols-4 tw-gap-4">
      <Card
        v-for="metric in metrics"
        :key="metric.key"
        class="tw-transition-transform hover:tw--translate-y-0.5"
      >
        <CardHeader class="tw-flex tw-flex-row tw-items-center tw-justify-between tw-space-y-0 tw-pb-2">
          <CardTitle class="tw-text-sm tw-font-medium tw-text-muted-foreground">
            {{ metric.label }}
          </CardTitle>
          <component
            :is="metric.icon"
            class="tw-size-4 tw-text-muted-foreground"
          />
        </CardHeader>
        <CardContent>
          <div class="tw-text-2xl tw-font-bold">
            {{ metric.value }}
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Trend Chart Card -->
    <Card>
      <CardHeader>
        <CardTitle class="tw-text-base tw-font-semibold">
          {{ t('statistics.last7DaysTrend') }}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <!-- SVG Line Chart (preserved from original, no external library) -->
        <div class="tw-overflow-x-auto">
          <svg
            class="tw-w-full tw-block"
            style="height: 240px; min-width: 480px;"
            viewBox="0 0 900 240"
            xmlns="http://www.w3.org/2000/svg"
          >
            <!-- Y-axis -->
            <line
              x1="70"
              y1="15"
              x2="70"
              y2="180"
              stroke="#909399"
              stroke-width="2"
            />
            <!-- X-axis -->
            <line
              x1="70"
              y1="180"
              x2="870"
              y2="180"
              stroke="#909399"
              stroke-width="2"
            />

            <!-- Y-axis ticks, labels, and grid lines -->
            <g
              v-for="i in 6"
              :key="'y-tick-' + i"
            >
              <line
                :x1="65"
                :y1="180 - (i - 1) * 33"
                :x2="70"
                :y2="180 - (i - 1) * 33"
                stroke="#909399"
                stroke-width="2"
              />
              <text
                :x="58"
                :y="180 - (i - 1) * 33 + 4"
                text-anchor="end"
                font-size="11"
                fill="#606266"
              >
                {{ Math.round((i - 1) * getMaxValue() / 5) }}
              </text>
              <!-- Grid line -->
              <line
                :x1="70"
                :y1="180 - (i - 1) * 33"
                :x2="870"
                :y2="180 - (i - 1) * 33"
                stroke="#e4e7ed"
                stroke-width="1"
                stroke-dasharray="5,5"
              />
            </g>

            <!-- X-axis ticks and date labels -->
            <g
              v-for="(item, index) in stats.chartData"
              :key="'x-tick-' + index"
            >
              <line
                :x1="getChartPointX(index)"
                :y1="180"
                :x2="getChartPointX(index)"
                :y2="185"
                stroke="#909399"
                stroke-width="2"
              />
              <text
                :x="getChartPointX(index)"
                :y="200"
                text-anchor="middle"
                font-size="11"
                fill="#606266"
              >
                {{ formatDate(item.date) }}
              </text>
            </g>

            <!-- Success polyline -->
            <polyline
              v-if="stats.chartData.length > 0"
              :points="getChartLinePoints('success')"
              fill="none"
              stroke="#22c55e"
              stroke-width="3"
              stroke-linecap="round"
              stroke-linejoin="round"
            />

            <!-- Failed polyline -->
            <polyline
              v-if="stats.chartData.length > 0"
              :points="getChartLinePoints('failed')"
              fill="none"
              stroke="#ef4444"
              stroke-width="3"
              stroke-linecap="round"
              stroke-linejoin="round"
            />

            <!-- Success data points -->
            <g
              v-for="(item, index) in stats.chartData"
              :key="'success-point-' + index"
            >
              <circle
                :cx="getChartPointX(index)"
                :cy="getChartPointY(item.success)"
                r="6"
                fill="#22c55e"
                stroke="#fff"
                stroke-width="2"
                class="tw-cursor-pointer"
              />
              <title>{{ item.date }}: {{ t('statistics.success') }} {{ item.success }}</title>
            </g>

            <!-- Failed data points -->
            <g
              v-for="(item, index) in stats.chartData"
              :key="'failed-point-' + index"
            >
              <circle
                :cx="getChartPointX(index)"
                :cy="getChartPointY(item.failed)"
                r="6"
                fill="#ef4444"
                stroke="#fff"
                stroke-width="2"
                class="tw-cursor-pointer"
              />
              <title>{{ item.date }}: {{ t('statistics.failed') }} {{ item.failed }}</title>
            </g>

            <!-- Y-axis label -->
            <text
              x="20"
              y="97"
              text-anchor="middle"
              font-size="12"
              fill="#606266"
              transform="rotate(-90, 20, 97)"
            >
              {{ t('statistics.executionCount') }}
            </text>

            <!-- X-axis label -->
            <text
              x="470"
              y="225"
              text-anchor="middle"
              font-size="12"
              fill="#606266"
            >
              {{ t('statistics.date') }}
            </text>
          </svg>
        </div>

        <!-- Legend -->
        <div class="tw-flex tw-justify-center tw-gap-8 tw-mt-3">
          <span class="tw-flex tw-items-center tw-gap-2 tw-text-sm tw-text-muted-foreground">
            <span class="tw-inline-block tw-w-3.5 tw-h-3.5 tw-rounded-sm tw-bg-green-500" />
            {{ t('statistics.success') }}
          </span>
          <span class="tw-flex tw-items-center tw-gap-2 tw-text-sm tw-text-muted-foreground">
            <span class="tw-inline-block tw-w-3.5 tw-h-3.5 tw-rounded-sm tw-bg-red-500" />
            {{ t('statistics.failed') }}
          </span>
        </div>
      </CardContent>
    </Card>

    <!-- Detailed Data Table Card -->
    <Card>
      <CardHeader>
        <CardTitle class="tw-text-base tw-font-semibold">
          {{ t('statistics.last7DaysTrend') }} — {{ t('statistics.detailedData') }}
        </CardTitle>
      </CardHeader>
      <CardContent class="tw-p-0">
        <div class="tw-overflow-x-auto">
          <table class="tw-w-full tw-text-sm">
            <thead>
              <tr class="tw-border-b tw-bg-muted/50">
                <th class="tw-px-4 tw-py-3 tw-text-left tw-font-medium tw-text-muted-foreground">
                  {{ t('common.date') }}
                </th>
                <th class="tw-px-4 tw-py-3 tw-text-left tw-font-medium tw-text-muted-foreground">
                  {{ t('statistics.total') }}
                </th>
                <th class="tw-px-4 tw-py-3 tw-text-left tw-font-medium tw-text-muted-foreground">
                  {{ t('statistics.success') }}
                </th>
                <th class="tw-px-4 tw-py-3 tw-text-left tw-font-medium tw-text-muted-foreground">
                  {{ t('statistics.failed') }}
                </th>
                <th class="tw-px-4 tw-py-3 tw-text-left tw-font-medium tw-text-muted-foreground">
                  {{ t('statistics.successRate') }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, index) in stats.last7Days"
                :key="index"
                class="tw-border-b tw-transition-colors hover:tw-bg-muted/30"
              >
                <td class="tw-px-4 tw-py-3 tw-text-foreground">
                  {{ row.date }}
                </td>
                <td class="tw-px-4 tw-py-3 tw-text-foreground">
                  {{ row.total }}
                </td>
                <td class="tw-px-4 tw-py-3">
                  <Badge variant="outline" class="tw-border-green-500 tw-text-green-600">
                    {{ row.success }}
                  </Badge>
                </td>
                <td class="tw-px-4 tw-py-3">
                  <Badge variant="outline" class="tw-border-red-500 tw-text-red-600">
                    {{ row.failed }}
                  </Badge>
                </td>
                <td class="tw-px-4 tw-py-3">
                  <div class="tw-flex tw-items-center tw-gap-2">
                    <!-- progress bar -->
                    <div class="tw-flex-1 tw-h-2 tw-rounded-full tw-bg-muted tw-overflow-hidden" style="min-width:80px;">
                      <div
                        class="tw-h-full tw-rounded-full tw-transition-all"
                        :style="{
                          width: calculateSuccessRate(row) + '%',
                          background: getProgressColor(calculateSuccessRate(row))
                        }"
                      />
                    </div>
                    <span class="tw-text-xs tw-tabular-nums tw-text-muted-foreground tw-w-8 tw-text-right">
                      {{ calculateSuccessRate(row) }}%
                    </span>
                  </div>
                </td>
              </tr>
              <tr v-if="stats.last7Days.length === 0">
                <td
                  colspan="5"
                  class="tw-px-4 tw-py-8 tw-text-center tw-text-muted-foreground"
                >
                  —
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  ListTodo,
  PlayCircle,
  CheckCircle2,
  XCircle,
  RefreshCw
} from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle
} from '@/components/ui/card'

import statisticsApi from '@/api/statistics'

const { t } = useI18n()

const stats = ref({
  totalTasks: 0,
  todayExecutions: 0,
  successRate: 0,
  failedCount: 0,
  last7Days: [],
  chartData: []
})

// KPI metric definitions
const metrics = computed(() => [
  {
    key: 'totalTasks',
    label: t('statistics.totalTasks'),
    value: stats.value.totalTasks,
    icon: ListTodo
  },
  {
    key: 'todayExecutions',
    label: t('statistics.last7DaysExecutions'),
    value: stats.value.todayExecutions,
    icon: PlayCircle
  },
  {
    key: 'successRate',
    label: t('statistics.successRate'),
    value: stats.value.successRate + '%',
    icon: CheckCircle2
  },
  {
    key: 'failedCount',
    label: t('statistics.failedCount'),
    value: stats.value.failedCount,
    icon: XCircle
  }
])

// Fetch statistics from API
const fetchStatistics = () => {
  statisticsApi.getOverview((data) => {
    if (data) {
      const last7Days = data.last_7_days || []

      const total7DaysSuccess = last7Days.reduce((sum, item) => sum + item.success, 0)
      const total7DaysFailed = last7Days.reduce((sum, item) => sum + item.failed, 0)
      const total7DaysExecutions = last7Days.reduce((sum, item) => sum + item.total, 0)

      let successRate7Days = 0
      if (total7DaysExecutions > 0) {
        successRate7Days = Math.round((total7DaysSuccess / total7DaysExecutions) * 1000) / 10
      }

      stats.value = {
        totalTasks: data.total_tasks || 0,
        todayExecutions: total7DaysExecutions,
        successRate: successRate7Days,
        failedCount: total7DaysFailed,
        last7Days: last7Days,
        chartData: [...last7Days].reverse()
      }
    }
  })
}

// Calculate per-row success rate (0–100)
const calculateSuccessRate = (row) => {
  if (row.total === 0) return 0
  return Math.round((row.success / row.total) * 100)
}

// Progress bar color thresholds
const getProgressColor = (percentage) => {
  if (percentage >= 90) return '#22c55e'
  if (percentage >= 70) return '#f59e0b'
  return '#ef4444'
}

// Chart helpers
const getMaxValue = () => {
  if (stats.value.chartData.length === 0) return 1
  const allValues = stats.value.chartData.flatMap((item) => [item.success, item.failed])
  return Math.max(...allValues, 1)
}

const getChartPointX = (index) => {
  const totalDays = stats.value.chartData.length
  if (totalDays <= 1) return 470
  const chartWidth = 800
  const spacing = chartWidth / (totalDays - 1)
  return 70 + spacing * index
}

const getChartPointY = (value) => {
  const maxValue = getMaxValue()
  if (maxValue === 0) return 180
  const chartHeight = 165
  const ratio = value / maxValue
  return 180 - ratio * chartHeight
}

const getChartLinePoints = (type) => {
  return stats.value.chartData
    .map((item, index) => {
      const x = getChartPointX(index)
      const y = getChartPointY(type === 'success' ? item.success : item.failed)
      return `${x},${y}`
    })
    .join(' ')
}

const formatDate = (dateStr) => {
  const date = new Date(dateStr)
  return `${date.getMonth() + 1}/${date.getDate()}`
}

const refresh = () => {
  fetchStatistics()
}

onMounted(() => {
  fetchStatistics()
})
</script>
