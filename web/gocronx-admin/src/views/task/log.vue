<template>
  <div class="task-log-page art-full-height">
    <!-- Filter card -->
    <ArtSearchBar v-model="filterForm" :items="filterItems" @search="handleSearch" @reset="handleReset" />

    <!-- Table card -->
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader :loading="loading" v-model:columns="columnChecks" @refresh="refreshData">
        <template #left>
          <span class="text-base font-medium">{{ t('task.log.title') }}</span>
        </template>
        <template #right>
          <ElButton type="danger" @click="handleClearAll">{{ t('task.log.clearAll') }}</ElButton>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <!-- Output dialog -->
    <ElDialog
      v-model="outputDialogVisible"
      :title="t('task.log.outputTitle')"
      width="680px"
      align-center
      destroy-on-close
    >
      <div v-if="currentLog">
        <div v-if="currentLog.hostname" style="margin-bottom: 12px">
          <strong>{{ t('task.log.colHost') }}:</strong>
          <!-- eslint-disable-next-line vue/no-v-html -->
          <pre class="log-pre" v-html="currentLog.hostname" />
        </div>
        <div style="margin-bottom: 12px">
          <strong>{{ t('task.name') }}:</strong>
          <pre class="log-pre">{{ currentLog.command }}</pre>
        </div>
        <div>
          <strong>{{ t('task.log.colOutput') }}:</strong>
          <pre class="log-pre">{{ currentLog.output || currentLog.result || t('task.log.noOutput') }}</pre>
        </div>
      </div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, h, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useRoute, useRouter } from 'vue-router'
  import { ElButton, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchTaskLogList, fetchTaskLogClear, fetchTaskLogStop, type TaskLogListItem } from '@/api/task-log'
  import { fetchTaskList } from '@/api/task'
  import { fetchHostList } from '@/api/host'
  import { formatDateTime } from '@/utils/date'

  defineOptions({ name: 'TaskLog' })

  const { t } = useI18n()
  const route = useRoute()
  const router = useRouter()

  // ── Status options (derived from i18n) ────────────────────────────────────
  const statusOptions = computed(() => [
    { value: '0', label: t('task.log.statusFailed') },
    { value: '1', label: t('task.log.statusRunning') },
    { value: '2', label: t('task.log.statusSuccess') },
    { value: '3', label: t('task.log.statusCancelled') }
  ])

  // ── Filter state ──────────────────────────────────────────────────────────
  const filterForm = ref<Record<string, any>>({
    task_id: '' as string | number,
    host_id: '' as string | number,
    status: '',
    protocol: '',
    date_range: [] as string[]
  })

  // ── Task select options ───────────────────────────────────────────────────
  const taskOptions = ref<{ value: number; label: string }[]>([])

  // ── Host dropdown ─────────────────────────────────────────────────────────
  const hostOptions = ref<{ value: number; label: string }[]>([])

  async function loadHosts() {
    try {
      const res = await fetchHostList({ page: 1, page_size: 200 })
      const list = (res as any)?.data ?? (res as any) ?? []
      hostOptions.value = (Array.isArray(list) ? list : []).map((item: any) => ({
        value: item.id,
        label: item.name || item.alias || String(item.id)
      }))
    } catch {
      hostOptions.value = []
    }
  }

  // ── Filter items config ───────────────────────────────────────────────────
  const filterItems = computed(() => [
    {
      label: t('task.name'),
      key: 'task_id',
      type: 'select',
      props: {
        placeholder: t('task.log.selectTask'),
        clearable: true,
        filterable: true,
        options: taskOptions.value
      }
    },
    {
      label: t('task.log.colHost'),
      key: 'host_id',
      type: 'select',
      props: {
        placeholder: t('task.log.selectHost'),
        clearable: true,
        options: hostOptions.value
      }
    },
    {
      label: t('task.log.colStatus'),
      key: 'status',
      type: 'select',
      props: {
        placeholder: t('task.log.selectStatus'),
        clearable: true,
        options: statusOptions.value
      }
    },
    {
      label: t('task.log.colProtocol'),
      key: 'protocol',
      type: 'select',
      props: {
        placeholder: t('task.log.selectProtocol'),
        clearable: true,
        options: [
          { label: 'HTTP', value: '1' },
          { label: 'Shell (RPC)', value: '2' }
        ]
      }
    },
    {
      label: t('task.log.dateRange'),
      key: 'date_range',
      type: 'daterange',
      props: {
        type: 'daterange',
        valueFormat: 'YYYY-MM-DD',
        rangeSeparator: '-',
        startPlaceholder: t('task.log.startDate'),
        endPlaceholder: t('task.log.endDate')
      }
    }
  ])

  // ── Output dialog ─────────────────────────────────────────────────────────
  const outputDialogVisible = ref(false)
  const currentLog = ref<TaskLogListItem | null>(null)

  function showOutput(row: TaskLogListItem) {
    // Decode HTML entities in command
    let cmd = row.command || ''
    cmd = cmd
      .replace(/&quot;/g, '"')
      .replace(/&apos;/g, "'")
      .replace(/&#39;/g, "'")
      .replace(/&lt;/g, '<')
      .replace(/&gt;/g, '>')
      .replace(/&amp;/g, '&')
    currentLog.value = { ...row, command: cmd }
    outputDialogVisible.value = true
  }

  // ── Format helpers ────────────────────────────────────────────────────────
  // Backend stores task-log hostname as "alias - name<br>alias2 - name2<br>".
  // For the table cell, split on <br>, drop empties, join with ", ".
  function formatHostList(raw: string): string {
    if (!raw) return '-'
    const parts = raw.split(/<br\s*\/?>/i).map((s) => s.trim()).filter(Boolean)
    return parts.length > 0 ? parts.join(', ') : '-'
  }

  function formatDuration(seconds: number): string {
    const s = seconds > 0 ? seconds : 1
    if (s < 60) return `${s}s`
    const m = Math.floor(s / 60)
    const rem = s % 60
    if (m < 60) return rem > 0 ? `${m}m ${rem}s` : `${m}m`
    const h = Math.floor(m / 60)
    const rm = m % 60
    return rm > 0 ? `${h}h ${rm}m` : `${h}h`
  }

  function statusTagType(status: number): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
    if (status === 0) return 'danger'
    if (status === 1) return 'warning'
    if (status === 2) return 'success'
    if (status === 3) return 'info'
    return 'info'
  }

  function statusLabel(status: number): string {
    if (status === 0) return t('task.log.statusFailed')
    if (status === 1) return t('task.log.statusRunning')
    if (status === 2) return t('task.log.statusSuccess')
    if (status === 3) return t('task.log.statusCancelled')
    return String(status)
  }

  function protocolLabel(protocol: number): string {
    if (protocol === 1) return 'HTTP'
    return 'Shell (RPC)'
  }

  function protocolTagType(protocol: number): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
    return protocol === 1 ? 'primary' : 'success'
  }

  // ── useTable ──────────────────────────────────────────────────────────────
  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    searchParams,
    getData,
    refreshData,
    handleSizeChange,
    handleCurrentChange,
    resetSearchParams
  } = useTable({
    core: {
      apiFn: fetchTaskLogList,
      apiParams: {
        page: 1,
        page_size: 20,
        task_id: '',
        protocol: '',
        status: '',
        host_id: '',
        start_date: '',
        end_date: ''
      },
      paginationKey: {
        current: 'page',
        size: 'page_size'
      },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#', align: 'center' },
        {
          prop: 'id',
          label: 'ID',
          width: 80,
          align: 'center'
        },
        {
          prop: 'task_name',
          label: t('task.log.colTaskName'),
          align: 'center',
          formatter: (row: TaskLogListItem) =>
            h(
              'a',
              {
                style: 'color: var(--el-color-primary); cursor: pointer;',
                onClick: () => router.push(`/task/edit/${row.task_id}`)
              },
              row.task_name || String(row.task_id)
            )
        },
        {
          prop: 'host_name',
          label: t('task.log.colHost'),
          align: 'center',
          formatter: (row: TaskLogListItem) =>
            h('span', {}, formatHostList(row.host_name || row.hostname || ''))
        },
        {
          prop: 'protocol',
          label: t('task.log.colProtocol'),
          width: 120,
          align: 'center',
          formatter: (row: TaskLogListItem) =>
            h(
              ElTag,
              { type: protocolTagType(row.protocol), size: 'small' },
              () => protocolLabel(row.protocol)
            )
        },
        {
          prop: 'status',
          label: t('task.log.colStatus'),
          width: 110,
          align: 'center',
          formatter: (row: TaskLogListItem) =>
            h(
              ElTag,
              { type: statusTagType(row.status), size: 'small' },
              () => statusLabel(row.status)
            )
        },
        {
          prop: 'start_time',
          label: t('task.log.colStartTime'),
          width: 180,
          align: 'center',
          formatter: (row: TaskLogListItem) => formatDateTime(row.start_time)
        },
        {
          prop: 'total_time',
          label: t('task.log.colDuration'),
          width: 100,
          align: 'center',
          formatter: (row: TaskLogListItem) => formatDuration(row.total_time)
        },
        {
          prop: 'action',
          label: t('task.log.colOutput'),
          width: 170,
          fixed: 'right',
          align: 'center',
          formatter: (row: TaskLogListItem) => {
            const btns = []

            // View output: available for finished runs (failed=0, success=2, cancelled=3)
            if (row.status === 0 || row.status === 2 || row.status === 3) {
              btns.push(
                h(
                  ElButton,
                  {
                    type: row.status === 2 ? 'success' : row.status === 0 ? 'warning' : 'info',
                    size: 'small',
                    style: 'margin-right: 4px',
                    onClick: () => showOutput(row)
                  },
                  () => t('task.log.viewOutput')
                )
              )
            }

            // Kill: only for running shell (RPC) jobs
            if (row.status === 1 && row.protocol === 2) {
              btns.push(
                h(
                  ElButton,
                  {
                    type: 'danger',
                    size: 'small',
                    onClick: () => handleKill(row)
                  },
                  () => t('task.log.kill')
                )
              )
            }

            return h('span', { style: 'display:inline-flex;gap:4px;flex-wrap:wrap;justify-content:center;' }, btns)
          }
        }
      ]
    }
  })

  // ── Search / Reset ────────────────────────────────────────────────────────
  function handleSearch() {
    const dr = filterForm.value.date_range
    const [startDate, endDate] = Array.isArray(dr) && dr.length === 2 ? dr : ['', '']

    Object.assign(searchParams, {
      task_id: filterForm.value.task_id || '',
      host_id: filterForm.value.host_id || '',
      status: filterForm.value.status || '',
      protocol: filterForm.value.protocol || '',
      start_date: startDate,
      end_date: endDate
    })

    getData()
  }

  function handleReset() {
    filterForm.value = { task_id: '', host_id: '', status: '', protocol: '', date_range: [] }
    resetSearchParams()
  }

  // ── Kill running job ──────────────────────────────────────────────────────
  async function handleKill(row: TaskLogListItem) {
    try {
      await fetchTaskLogStop(row.id, row.task_id)
      ElMessage.success(t('task.log.killSuccess'))
      refreshData()
    } catch {
      // error already handled by http interceptor
    }
  }

  // ── Clear all logs ────────────────────────────────────────────────────────
  async function handleClearAll() {
    try {
      await ElMessageBox.confirm(t('task.log.confirmClear'), t('task.log.clearAll'), {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning',
        center: true
      })
      await fetchTaskLogClear()
      ElMessage.success(t('task.log.clearSuccess'))
      refreshData()
    } catch (err: any) {
      if (err !== 'cancel' && err?.message !== 'cancel') {
        if (err && err.message) {
          // error already handled by http interceptor
        }
      }
    }
  }

  // ── Mount: load hosts + tasks, pre-fill task_id from route query ─────────
  onMounted(async () => {
    await loadHosts()

    // Always load task list so the task select is populated
    try {
      const res = await fetchTaskList({ page: 1, page_size: 200 })
      const list = (res as any)?.data ?? (res as any) ?? []
      taskOptions.value = (Array.isArray(list) ? list : []).map((item: any) => ({
        value: item.id,
        label: item.name
      }))
    } catch {
      taskOptions.value = []
    }

    const queryTaskId = route.query.task_id
    if (queryTaskId) {
      const numId = Number(queryTaskId)
      filterForm.value.task_id = numId
      Object.assign(searchParams, { task_id: numId })
      getData()
    }
  })
</script>

<style scoped>
  .task-log-page {
    display: flex;
    flex-direction: column;
  }

  .log-pre {
    white-space: pre-wrap;
    word-break: break-all;
    font-family: monospace;
    font-size: 13px;
    background: #2d2d2d;
    color: #f0f0f0;
    padding: 12px;
    border-radius: 4px;
    margin: 6px 0 0 0;
    max-height: 300px;
    overflow-y: auto;
  }
</style>
