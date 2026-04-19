<!-- Dashboard / Console — gocron KPI overview -->
<template>
  <div class="gocron-console">
    <!-- KPI cards -->
    <ElRow :gutter="20" class="kpi-row">
      <ElCol v-for="kpi in kpis" :key="kpi.key" :xs="24" :sm="12" :md="6">
        <div class="art-card kpi-card">
          <div class="kpi-icon-wrap" :style="{ background: kpi.color }">
            <ElIcon :size="22" color="#fff">
              <component :is="kpi.icon" />
            </ElIcon>
          </div>
          <div class="kpi-body">
            <ArtCountTo class="kpi-value" :target="kpi.value" :duration="900" />
            <div class="kpi-label">{{ kpi.label }}</div>
          </div>
        </div>
      </ElCol>
    </ElRow>

    <!-- Execution trend chart -->
    <div class="art-card chart-section">
      <div class="section-header">
        <span class="section-title">{{ t('dashboard.executionTrend') }}</span>
        <ElButton size="small" @click="load">{{ t('dashboard.refresh') }}</ElButton>
      </div>
      <ArtLineChart
        height="280px"
        :data="chartSeries"
        :xAxisData="chartDates"
        :showLegend="true"
        :showAreaColor="true"
        :loading="loading"
      />
    </div>

    <!-- Last-7-days detail table -->
    <div class="art-card table-section">
      <div class="section-header">
        <span class="section-title">{{ t('dashboard.last7Days') }}</span>
      </div>
      <ElTable :data="tableRows" border size="small" style="width: 100%">
        <ElTableColumn prop="date" :label="t('dashboard.colDate')" width="140" align="center" />
        <ElTableColumn prop="total" :label="t('dashboard.colTotal')" width="100" align="center" />
        <ElTableColumn prop="success" :label="t('dashboard.colSuccess')" width="100" align="center">
          <template #default="{ row }">
            <ElTag type="success" size="small">{{ row.success }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="failed" :label="t('dashboard.colFailed')" width="100" align="center">
          <template #default="{ row }">
            <ElTag type="danger" size="small">{{ row.failed }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn :label="t('dashboard.colSuccessRate')" align="center">
          <template #default="{ row }">
            <ElProgress
              :percentage="calcRate(row)"
              :color="rateColor(calcRate(row))"
              :stroke-width="8"
            />
          </template>
        </ElTableColumn>
      </ElTable>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { Document, CircleCheck, CircleClose, Monitor } from '@element-plus/icons-vue'
  import { fetchStatisticsOverview, type DayStats } from '@/api/statistics'
  import type { LineDataItem } from '@/types/component/chart'

  defineOptions({ name: 'Console' })

  const { t } = useI18n()

  // ── state ────────────────────────────────────────────────────────────────────
  const loading = ref(false)
  const totalTasks = ref(0)
  const last7Days = ref<DayStats[]>([])

  // ── derived ───────────────────────────────────────────────────────────────────
  /** Sum helpers over last-7-days rows */
  const sum = (field: keyof DayStats) =>
    last7Days.value.reduce((acc, r) => acc + (r[field] as number), 0)

  const total7DaysExec = computed(() => sum('total'))
  const total7DaysFailed = computed(() => sum('failed'))
  const total7DaysSuccess = computed(() => sum('success'))
  const successRate7Days = computed(() => {
    const t = total7DaysExec.value
    return t > 0 ? Math.round((total7DaysSuccess.value / t) * 1000) / 10 : 0
  })

  const kpis = computed(() => [
    {
      key: 'tasks',
      label: t('dashboard.taskCount'),
      value: totalTasks.value,
      color: '#409EFF',
      icon: Document
    },
    {
      key: 'executions',
      label: t('dashboard.taskLogCount'),
      value: total7DaysExec.value,
      color: '#67C23A',
      icon: CircleCheck
    },
    {
      key: 'failed',
      label: t('dashboard.failedCount'),
      value: total7DaysFailed.value,
      color: '#F56C6C',
      icon: CircleClose
    },
    {
      key: 'successRate',
      label: t('dashboard.successRate'),
      value: successRate7Days.value,
      color: '#E6A23C',
      icon: Monitor
    }
  ])

  /** Chart: chronological order (oldest → newest), two series: success / failed */
  const chartDates = computed(() => [...last7Days.value].reverse().map((r) => r.date))

  const chartSeries = computed<LineDataItem[]>(() => {
    const asc = [...last7Days.value].reverse()
    return [
      { name: t('dashboard.colSuccess'), data: asc.map((r) => r.success) },
      { name: t('dashboard.colFailed'), data: asc.map((r) => r.failed) }
    ]
  })

  /** Table: keep DESC order (newest first) */
  const tableRows = computed(() => last7Days.value)

  // ── helpers ────────────────────────────────────────────────────────────────
  const calcRate = (row: DayStats) =>
    row.total === 0 ? 0 : Math.round((row.success / row.total) * 100)

  const rateColor = (pct: number) => {
    if (pct >= 90) return '#67C23A'
    if (pct >= 70) return '#E6A23C'
    return '#F56C6C'
  }

  // ── data fetch ────────────────────────────────────────────────────────────
  async function load() {
    loading.value = true
    try {
      const data = await fetchStatisticsOverview()
      totalTasks.value = data?.total_tasks ?? 0
      last7Days.value = data?.last_7_days ?? []
    } finally {
      loading.value = false
    }
  }

  onMounted(load)
</script>

<style scoped>
  .gocron-console {
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  /* KPI row */
  .kpi-row {
    /* ElRow uses negative margin; reset so outer gap works */
  }

  .kpi-card {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 20px;
    margin-bottom: 0;
    min-height: 90px;
  }

  .kpi-icon-wrap {
    width: 48px;
    height: 48px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .kpi-body {
    flex: 1;
    min-width: 0;
  }

  .kpi-value {
    font-size: 26px;
    font-weight: 600;
    line-height: 1.2;
    display: block;
  }

  .kpi-label {
    margin-top: 4px;
    font-size: 13px;
    color: var(--el-text-color-regular);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* Shared card section */
  .chart-section,
  .table-section {
    padding: 20px;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .section-title {
    font-size: 15px;
    font-weight: 500;
    color: var(--el-text-color-primary);
  }
</style>
