<template>
  <div class="task-list-page art-full-height">
    <!-- Filter card -->
    <ArtSearchBar v-model="filterForm" :items="filterItems" @search="handleSearch" @reset="handleReset" />

    <!-- Table card -->
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader :loading="loading" v-model:columns="columnChecks" @refresh="refreshData">
        <template #left>
          <span class="text-base font-medium">{{ t('menus.task.list') }}</span>
        </template>
        <template #right>
          <div class="header-btn-group">
            <ElButton type="success" @click="handleBatchEnable">
              {{ t('task.batchEnable') }}
            </ElButton>
            <ElButton type="warning" @click="handleBatchDisable">
              {{ t('task.batchDisable') }}
            </ElButton>
            <ElButton type="danger" @click="handleBatchRemove">
              {{ t('task.batchDelete') }}
            </ElButton>
            <ElButton type="primary" @click="toCreate">{{ t('task.addTask') }}</ElButton>
          </div>
        </template>
      </ArtTableHeader>

      <ArtTable
        ref="tableRef"
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @selection-change="handleSelectionChange"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, h, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useRouter } from 'vue-router'
  import { ElButton, ElMessage, ElMessageBox, ElSwitch, ElTag } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchTaskList,
    fetchTaskRemove,
    fetchTaskEnable,
    fetchTaskDisable,
    fetchTaskRunOnce,
    fetchBatchEnable,
    fetchBatchDisable,
    fetchBatchRemove,
    type TaskListItem
  } from '@/api/task'
  import { fetchHostList, type HostItem } from '@/api/host'
  import { formatDateTime } from '@/utils/date'

  defineOptions({ name: 'TaskList' })

  const { t } = useI18n()
  const router = useRouter()

  // ── Filter state ──────────────────────────────────────────────────────────────
  const filterForm = ref({
    name: '',
    host_id: '' as number | string,
    protocol: '' as number | string,
    status: '' as number | string,
    tag: ''
  })

  // ── Host dropdown options ─────────────────────────────────────────────────────
  const hostOptions = ref<HostItem[]>([])

  async function loadHostOptions() {
    try {
      const res = await fetchHostList({ page: 1, page_size: 999 })
      const list = (res as any)?.data ?? res
      hostOptions.value = Array.isArray(list) ? list : list?.data ?? []
    } catch {
      // ignore
    }
  }

  onMounted(() => {
    loadHostOptions()
  })

  // ── Filter items config ───────────────────────────────────────────────────────
  const filterItems = computed(() => [
    {
      label: t('task.name'),
      key: 'name',
      type: 'input',
      props: { placeholder: t('task.namePlaceholder'), clearable: true }
    },
    {
      label: t('task.host'),
      key: 'host_id',
      type: 'select',
      props: {
        placeholder: t('task.selectHost'),
        clearable: true,
        filterable: true,
        options: hostOptions.value.map((h) => ({
          label: `${h.alias} - ${h.name}:${h.port}`,
          value: h.id
        }))
      }
    },
    {
      label: t('task.protocol'),
      key: 'protocol',
      type: 'select',
      props: {
        placeholder: t('task.selectProtocol'),
        clearable: true,
        options: [
          { label: t('task.protocolHttp'), value: 1 },
          { label: t('task.protocolRpc'), value: 2 }
        ]
      }
    },
    {
      label: t('task.status'),
      key: 'status',
      type: 'select',
      props: {
        placeholder: t('task.selectStatus'),
        clearable: true,
        options: [
          { label: t('task.statusEnabled'), value: 1 },
          { label: t('task.statusDisabled'), value: 0 }
        ]
      }
    },
    {
      label: t('task.tag'),
      key: 'tag',
      type: 'input',
      props: { placeholder: t('task.tagPlaceholder'), clearable: true }
    }
  ])

  // ── Row selection ─────────────────────────────────────────────────────────────
  const tableRef = ref<any>(null)
  const selectedIds = ref<number[]>([])

  function handleSelectionChange(rows: TaskListItem[]) {
    selectedIds.value = (rows || []).filter((r) => r.level === 1).map((r) => r.id)
  }

  /**
   * Fallback in case @selection-change doesn't bubble up through ArtTable —
   * reads selection directly from the underlying ElTable ref. Called before
   * each batch action so the button works even if live tracking is off.
   */
  function readCurrentSelection(): number[] {
    const inner = tableRef.value?.elTableRef
    const rows = (inner?.getSelectionRows?.() || []) as TaskListItem[]
    return rows.filter((r) => r.level === 1).map((r) => r.id)
  }

  // ── Helpers ───────────────────────────────────────────────────────────────────
  function parseCronSpec(spec: string): { expr: string; tz: string } {
    if (!spec) return { expr: '', tz: '' }
    const match = spec.match(/^(?:CRON_TZ|TZ)=(\S+)\s+(.+)$/)
    if (match) return { tz: match[1], expr: match[2] }
    return { expr: spec, tz: '' }
  }

  function formatProtocol(row: TaskListItem): string {
    if (row.protocol === 2) return 'shell'
    if (row.http_method === 1) return 'http-get'
    return 'http-post'
  }

  // ── useTable ──────────────────────────────────────────────────────────────────
  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    searchParams,
    getData,
    refreshData,
    refreshRemove,
    handleSizeChange,
    handleCurrentChange,
    resetSearchParams
  } = useTable({
    core: {
      apiFn: fetchTaskList,
      apiParams: {
        page: 1,
        page_size: 20,
        name: '',
        host_id: '',
        protocol: '',
        status: '',
        tag: ''
      },
      paginationKey: {
        current: 'page',
        size: 'page_size'
      },
      columnsFactory: () => [
        { type: 'selection', width: 50, align: 'center' },
        {
          prop: 'id',
          label: t('task.id'),
          width: 70,
          align: 'center'
        },
        {
          prop: 'name',
          label: t('task.name'),
          minWidth: 140,
          align: 'center'
        },
        {
          prop: 'spec',
          label: t('task.spec'),
          minWidth: 160,
          align: 'center',
          formatter: (row: TaskListItem) => {
            const { expr, tz } = parseCronSpec(row.spec)
            if (!tz) return h('span', {}, expr)
            return h('div', {}, [
              h('div', {}, expr),
              h('div', { style: 'color:#909399;font-size:12px;line-height:1.4' }, tz)
            ])
          }
        },
        {
          prop: 'protocol',
          label: t('task.protocol'),
          width: 110,
          align: 'center',
          formatter: (row: TaskListItem) => {
            const label = formatProtocol(row)
            const type = row.protocol === 2 ? 'warning' : 'primary'
            return h(ElTag, { type, size: 'small' }, () => label)
          }
        },
        {
          prop: 'hosts',
          label: t('task.host'),
          minWidth: 160,
          align: 'center',
          formatter: (row: TaskListItem) => {
            const hosts = row.hosts || []
            if (hosts.length === 0) return h('span', { style: 'color:#c0c4cc' }, '-')
            return h(
              'div',
              {},
              hosts.map((h_) =>
                h(
                  'div',
                  { key: h_.host_id },
                  `${h_.alias} - ${h_.name}:${h_.port}`
                )
              )
            )
          }
        },
        {
          prop: 'status',
          label: t('task.status'),
          width: 100,
          align: 'center',
          formatter: (row: TaskListItem) => {
            if (row.level !== 1) return h('span', { style: 'color:#c0c4cc' }, '-')
            return h(ElSwitch, {
              modelValue: row.status === 1,
              activeValue: true,
              inactiveValue: false,
              'onUpdate:modelValue': (val: string | number | boolean) => handleStatusToggle(row, val)
            })
          }
        },
        {
          prop: 'tag',
          label: t('task.tag'),
          minWidth: 120,
          align: 'center',
          formatter: (row: TaskListItem) => {
            if (!row.tag) return h('span', { style: 'color:#c0c4cc' }, '-')
            const tags = row.tag.split(',').filter(Boolean)
            return h(
              'span',
              { style: 'display:inline-flex;flex-wrap:wrap;gap:4px;justify-content:center' },
              tags.map((tag) => h(ElTag, { size: 'small', key: tag }, () => tag))
            )
          }
        },
        {
          prop: 'next_run_time',
          label: t('task.nextRunTime'),
          width: 170,
          align: 'center',
          formatter: (row: TaskListItem) =>
            h('span', {}, formatDateTime(row.next_run_time) || '-')
        },
        {
          prop: 'action',
          label: t('task.actions'),
          width: 180,
          fixed: 'right',
          align: 'center',
          formatter: (row: TaskListItem) =>
            h(
              'div',
              {
                style:
                  'display:grid;grid-template-columns:1fr 1fr;gap:6px;justify-items:stretch;padding:4px 0'
              },
              [
                h(
                  ElButton,
                  {
                    type: 'primary',
                    size: 'small',
                    style: 'margin:0',
                    onClick: () => toEdit(row)
                  },
                  () => t('task.edit')
                ),
                h(
                  ElButton,
                  {
                    type: 'success',
                    size: 'small',
                    style: 'margin:0',
                    onClick: () => handleRunOnce(row)
                  },
                  () => t('task.runOnce')
                ),
                h(
                  ElButton,
                  {
                    type: 'info',
                    size: 'small',
                    style: 'margin:0',
                    onClick: () => toLog(row)
                  },
                  () => t('task.viewLog')
                ),
                h(
                  ElButton,
                  {
                    type: 'danger',
                    size: 'small',
                    style: 'margin:0',
                    onClick: () => handleRemove(row)
                  },
                  () => t('task.delete')
                )
              ]
            )
        }
      ]
    }
  })

  // ── Search / Reset ────────────────────────────────────────────────────────────
  function handleSearch() {
    Object.assign(searchParams, {
      name: filterForm.value.name || '',
      host_id: filterForm.value.host_id ?? '',
      protocol: filterForm.value.protocol ?? '',
      status: filterForm.value.status ?? '',
      tag: filterForm.value.tag || ''
    })
    getData()
  }

  function handleReset() {
    filterForm.value = { name: '', host_id: '', protocol: '', status: '', tag: '' }
    resetSearchParams()
  }

  // ── Navigation ────────────────────────────────────────────────────────────────
  function toCreate() {
    router.push('/task/create')
  }

  function toEdit(row: TaskListItem) {
    router.push(`/task/edit/${row.id}`)
  }

  function toLog(row: TaskListItem) {
    router.push(`/task/log?task_id=${row.id}`)
  }

  // ── Row actions ───────────────────────────────────────────────────────────────
  async function handleRunOnce(row: TaskListItem) {
    try {
      await ElMessageBox.confirm(
        t('task.confirmRunOnce', { name: row.name }),
        t('task.confirmTitle'),
        {
          confirmButtonText: t('task.confirm'),
          cancelButtonText: t('task.cancel'),
          type: 'warning',
          center: true
        }
      )
      await fetchTaskRunOnce(row.id)
      ElMessage.success(t('task.runOnceSuccess'))
    } catch (err: any) {
      if (err && err.message) ElMessage.error(String(err.message))
    }
  }

  async function handleStatusToggle(row: TaskListItem, val: string | number | boolean) {
    const enabled = Boolean(val)
    try {
      if (enabled) {
        await fetchTaskEnable(row.id)
        ElMessage.success(t('task.enableSuccess'))
      } else {
        await fetchTaskDisable(row.id)
        ElMessage.success(t('task.disableSuccess'))
      }
      refreshData()
    } catch (err: any) {
      if (err && err.message) ElMessage.error(String(err.message))
      // refresh to revert the optimistic switch state on failure
      refreshData()
    }
  }

  async function handleRemove(row: TaskListItem) {
    try {
      await ElMessageBox.confirm(
        t('task.confirmDelete', { name: row.name }),
        t('task.confirmTitle'),
        {
          confirmButtonText: t('task.confirm'),
          cancelButtonText: t('task.cancel'),
          type: 'warning',
          center: true
        }
      )
      await fetchTaskRemove(row.id)
      ElMessage.success(t('task.deleteSuccess'))
      refreshRemove()
    } catch (err: any) {
      if (err && err.message) ElMessage.error(String(err.message))
    }
  }

  // ── Batch actions ─────────────────────────────────────────────────────────────
  async function handleBatchEnable() {
    const ids = selectedIds.value.length > 0 ? selectedIds.value : readCurrentSelection()
    if (ids.length === 0) {
      ElMessage.warning(t('task.pleaseSelect'))
      return
    }
    try {
      await ElMessageBox.confirm(
        t('task.confirmBatchEnable', { count: ids.length }),
        t('task.confirmTitle'),
        {
          confirmButtonText: t('task.confirm'),
          cancelButtonText: t('task.cancel'),
          type: 'warning'
        }
      )
      await fetchBatchEnable(ids)
      ElMessage.success(t('task.enableSuccess'))
      selectedIds.value = []
      tableRef.value?.elTableRef?.clearSelection?.()
      refreshData()
    } catch (err: any) {
      if (err && err.message) ElMessage.error(String(err.message))
    }
  }

  async function handleBatchDisable() {
    const ids = selectedIds.value.length > 0 ? selectedIds.value : readCurrentSelection()
    if (ids.length === 0) {
      ElMessage.warning(t('task.pleaseSelect'))
      return
    }
    try {
      await ElMessageBox.confirm(
        t('task.confirmBatchDisable', { count: ids.length }),
        t('task.confirmTitle'),
        {
          confirmButtonText: t('task.confirm'),
          cancelButtonText: t('task.cancel'),
          type: 'warning'
        }
      )
      await fetchBatchDisable(ids)
      ElMessage.success(t('task.disableSuccess'))
      selectedIds.value = []
      tableRef.value?.elTableRef?.clearSelection?.()
      refreshData()
    } catch (err: any) {
      if (err && err.message) ElMessage.error(String(err.message))
    }
  }

  async function handleBatchRemove() {
    const ids = selectedIds.value.length > 0 ? selectedIds.value : readCurrentSelection()
    if (ids.length === 0) {
      ElMessage.warning(t('task.pleaseSelect'))
      return
    }
    try {
      await ElMessageBox.confirm(
        t('task.confirmBatchDelete', { count: ids.length }),
        t('task.confirmTitle'),
        {
          confirmButtonText: t('task.confirm'),
          cancelButtonText: t('task.cancel'),
          type: 'error'
        }
      )
      await fetchBatchRemove(ids)
      ElMessage.success(t('task.deleteSuccess'))
      selectedIds.value = []
      tableRef.value?.elTableRef?.clearSelection?.()
      refreshData()
    } catch (err: any) {
      if (err && err.message) ElMessage.error(String(err.message))
    }
  }
</script>

<style scoped>
  .task-list-page {
    display: flex;
    flex-direction: column;
  }

  .header-btn-group {
    display: inline-flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .header-btn-group :deep(.el-button) {
    min-width: 110px;
    margin-left: 0;
  }
</style>
