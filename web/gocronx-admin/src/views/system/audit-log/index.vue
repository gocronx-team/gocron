<template>
  <div class="audit-log-page art-full-height">
    <!-- Filter card -->
    <ElCard shadow="never" class="mb-3">
      <ElForm :inline="true" :model="filterForm" @submit.prevent="handleSearch">
        <ElFormItem :label="t('audit.module')">
          <ElSelect
            v-model="filterForm.module"
            clearable
            style="width: 150px"
            :placeholder="t('audit.allModules')"
          >
            <ElOption
              v-for="item in moduleOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
        </ElFormItem>

        <ElFormItem :label="t('audit.action')">
          <ElSelect
            v-model="filterForm.action"
            clearable
            style="width: 170px"
            :placeholder="t('audit.allActions')"
          >
            <ElOption
              v-for="item in actionOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
        </ElFormItem>

        <ElFormItem :label="t('audit.username')">
          <ElInput
            v-model.trim="filterForm.username"
            clearable
            :placeholder="t('audit.usernamePlaceholder')"
            style="width: 160px"
          />
        </ElFormItem>

        <ElFormItem :label="t('audit.dateRange')">
          <ElDatePicker
            v-model="dateRange"
            type="daterange"
            value-format="YYYY-MM-DD"
            range-separator="-"
            :start-placeholder="t('audit.startDate')"
            :end-placeholder="t('audit.endDate')"
            style="width: 240px"
          />
        </ElFormItem>

        <ElFormItem>
          <ElButton type="primary" @click="handleSearch">{{ t('audit.search') }}</ElButton>
          <ElButton @click="handleReset">{{ t('audit.reset') }}</ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>

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
  const filterForm = ref({
    module: '',
    action: '',
    username: ''
  })
  const dateRange = ref<string[]>([])

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
          align: 'center'
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
    const [startDate, endDate] =
      dateRange.value && dateRange.value.length === 2 ? dateRange.value : ['', '']

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
    filterForm.value = { module: '', action: '', username: '' }
    dateRange.value = []
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
