<template>
  <div class="audit-log-page art-full-height">
    <!-- Filter -->
    <ArtSearchBar
      v-model="filterForm"
      :items="filterItems"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- Table card -->
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader :loading="loading" v-model:columns="columnChecks" @refresh="refreshData">
        <template #left>
          <span class="text-base font-medium">{{ t('menus.system.auditLog') }}</span>
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

    <!-- Detail dialog -->
    <ElDialog
      v-model="dialogVisible"
      :title="t('audit.detailTitle')"
      width="640px"
      align-center
      destroy-on-close
    >
      <template v-if="detailRows.length > 0">
        <ElTable :data="detailRows" border size="small">
          <ElTableColumn prop="field" :label="t('audit.detailField')" width="160" />
          <ElTableColumn prop="old" :label="t('audit.detailBefore')" />
          <ElTableColumn width="40" align="center">
            <template #default>&rarr;</template>
          </ElTableColumn>
          <ElTableColumn prop="new" :label="t('audit.detailAfter')" />
        </ElTable>
      </template>
      <template v-else>
        <pre class="detail-raw">{{ rawDetail }}</pre>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, h } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElTag, ElButton, ElTableColumn, ElTable } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchAuditList, type AuditListItem } from '@/api/audit'
  import { formatDateTime } from '@/utils/date'
  import {
    AUDIT_MODULES,
    AUDIT_ACTIONS,
    MODULE_TAG_TYPES,
    ACTION_TAG_TYPES
  } from '@/enums/auditEnum'

  defineOptions({ name: 'AuditLog' })

  const { t } = useI18n()

  // ── i18n-derived option lists ────────────────────────────────────────────
  const moduleOptions = computed(() =>
    AUDIT_MODULES.map((m) => ({ value: m.value, label: t(m.labelKey) }))
  )
  const actionOptions = computed(() =>
    AUDIT_ACTIONS.map((a) => ({ value: a.value, label: t(a.labelKey) }))
  )

  // ── Filter state ─────────────────────────────────────────────────────────
  const filterForm = ref<Record<string, any>>({
    module: '',
    action: '',
    username: '',
    dateRange: [] as string[]
  })

  const filterItems = computed(() => [
    {
      label: t('audit.module'),
      key: 'module',
      type: 'select',
      props: { placeholder: t('audit.allModules'), clearable: true, options: moduleOptions.value }
    },
    {
      label: t('audit.action'),
      key: 'action',
      type: 'select',
      props: { placeholder: t('audit.allActions'), clearable: true, options: actionOptions.value }
    },
    {
      label: t('audit.username'),
      key: 'username',
      type: 'input',
      props: { placeholder: t('audit.usernamePlaceholder'), clearable: true }
    },
    {
      label: t('audit.dateRange'),
      key: 'dateRange',
      type: 'daterange',
      props: {
        type: 'daterange',
        valueFormat: 'YYYY-MM-DD',
        rangeSeparator: '-',
        startPlaceholder: t('audit.startDate'),
        endPlaceholder: t('audit.endDate')
      }
    }
  ])

  // ── Detail dialog ─────────────────────────────────────────────────────────
  const dialogVisible = ref(false)
  const detailRows = ref<{ field: string; old: string; new: string }[]>([])
  const rawDetail = ref('')

  function showDetail(row: AuditListItem) {
    const detail = row.detail || ''
    // Try to parse structured diff lines: "fieldName: oldValue → newValue"
    const lines = detail.split('\n').filter(Boolean)
    const parsed = lines.map((line) => {
      const arrowIdx = line.indexOf(' → ')
      if (arrowIdx === -1) return null
      const left = line.slice(0, arrowIdx)
      const newVal = line.slice(arrowIdx + 3)
      const colonIdx = left.indexOf(': ')
      if (colonIdx === -1) return null
      return {
        field: left.slice(0, colonIdx),
        old: left.slice(colonIdx + 2),
        new: newVal
      }
    })
    const structured = parsed.filter(Boolean) as { field: string; old: string; new: string }[]

    if (structured.length > 0) {
      detailRows.value = structured
      rawDetail.value = ''
    } else {
      detailRows.value = []
      rawDetail.value = detail
    }
    dialogVisible.value = true
  }

  // ── useTable ─────────────────────────────────────────────────────────────
  // gocron uses page / page_size — remap via paginationKey
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
      apiFn: fetchAuditList,
      apiParams: {
        page: 1,
        page_size: 20,
        module: '',
        action: '',
        username: '',
        start_date: '',
        end_date: ''
      },
      paginationKey: {
        current: 'page',
        size: 'page_size'
      },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        {
          prop: 'created',
          label: t('audit.colCreated'),
          width: 180,
          align: 'center',
          formatter: (row: AuditListItem) => formatDateTime(row.created)
        },
        {
          prop: 'username',
          label: t('audit.colUsername'),
          align: 'center'
        },
        {
          prop: 'module',
          label: t('audit.colModule'),
          width: 110,
          align: 'center',
          formatter: (row: AuditListItem) =>
            h(
              ElTag,
              { type: MODULE_TAG_TYPES[row.module] ?? 'info', size: 'small' },
              () => moduleOptions.value.find((m) => m.value === row.module)?.label ?? row.module
            )
        },
        {
          prop: 'action',
          label: t('audit.colAction'),
          width: 140,
          align: 'center',
          formatter: (row: AuditListItem) =>
            h(
              ElTag,
              { type: ACTION_TAG_TYPES[row.action] ?? 'info', size: 'small' },
              () => actionOptions.value.find((a) => a.value === row.action)?.label ?? row.action
            )
        },
        {
          prop: 'target',
          label: t('audit.colTarget'),
          align: 'center',
          formatter: (row: AuditListItem) =>
            h('span', {}, row.target_name || String(row.target_id || ''))
        },
        {
          prop: 'ip',
          label: t('audit.colIp'),
          align: 'center'
        },
        {
          prop: 'detail',
          label: t('audit.colDetail'),
          width: 120,
          align: 'center',
          formatter: (row: AuditListItem) =>
            row.detail
              ? h(
                  ElButton,
                  {
                    type: 'info',
                    size: 'small',
                    onClick: () => showDetail(row)
                  },
                  () => t('audit.viewDetail')
                )
              : h('span', {}, '-')
        }
      ]
    }
  })

  // ── Search / Reset ────────────────────────────────────────────────────────
  function handleSearch() {
    const dr = filterForm.value.dateRange
    const [startDate, endDate] = Array.isArray(dr) && dr.length === 2 ? dr : ['', '']

    Object.assign(searchParams, {
      module: filterForm.value.module || '',
      action: filterForm.value.action || '',
      username: filterForm.value.username || '',
      start_date: startDate,
      end_date: endDate
    })

    getData()
  }

  function handleReset() {
    filterForm.value = { module: '', action: '', username: '', dateRange: [] }
    resetSearchParams()
  }
</script>

<style scoped>
  .audit-log-page {
    display: flex;
    flex-direction: column;
  }

  .detail-raw {
    white-space: pre-wrap;
    word-break: break-all;
    font-family: monospace;
    font-size: 13px;
    background: var(--el-fill-color-light);
    padding: 12px;
    border-radius: 4px;
    max-height: 400px;
    overflow-y: auto;
  }
</style>
