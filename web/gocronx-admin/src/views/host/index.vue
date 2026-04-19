<template>
  <div class="host-list-page art-full-height">
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
          <span class="text-base font-medium">{{ t('menus.host.list') }}</span>
        </template>
        <template #right>
          <ElButton type="primary" @click="toCreate">{{ t('host.addNode') }}</ElButton>
          <ElButton type="success" @click="showAutoRegister">{{ t('host.autoRegister') }}</ElButton>
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

    <!-- Auto-register dialog -->
    <ElDialog
      v-model="registerDialogVisible"
      :title="t('host.autoRegister')"
      width="680px"
      align-center
      destroy-on-close
    >
      <div v-if="agentTokenData">
        <ElAlert
          :title="t('host.installTip')"
          type="info"
          :closable="false"
          style="margin-bottom: 16px"
          show-icon
        />

        <div style="margin-bottom: 12px">
          <div class="install-label">{{ t('host.installCmd') }}</div>
          <pre class="install-pre">{{ agentTokenData.install_cmd }}</pre>
          <div style="text-align: right; margin-top: 8px">
            <ElButton type="primary" size="small" @click="copyInstallCmd">
              {{ t('host.copy') }}
            </ElButton>
          </div>
        </div>

        <ElDivider />

        <ElDescriptions :column="1" border size="small">
          <ElDescriptionsItem :label="t('host.tokenExpires')">
            <ElTag type="warning" effect="plain">{{ formatDateTime(agentTokenData.expires_at) }}</ElTag>
          </ElDescriptionsItem>
          <ElDescriptionsItem :label="t('host.tokenUsage')">
            <span style="color: var(--el-color-success)">{{ t('host.tokenReusable') }}</span>
          </ElDescriptionsItem>
        </ElDescriptions>
      </div>

      <div v-else style="text-align: center; padding: 32px 0">
        <ElIcon class="is-loading" :size="28">
          <Loading />
        </ElIcon>
        <p style="margin-top: 12px; color: var(--el-text-color-secondary)">{{ t('host.loading') }}</p>
      </div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, h } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useRouter } from 'vue-router'
  import { ElButton, ElMessage, ElMessageBox, ElTag, ElIcon } from 'element-plus'
  import { Loading } from '@element-plus/icons-vue'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchHostList,
    pingHost,
    removeHost,
    generateAgentToken,
    type HostItem,
    type AgentTokenResult
  } from '@/api/host'
  import { formatDateTime } from '@/utils/date'

  defineOptions({ name: 'HostList' })

  const { t } = useI18n()
  const router = useRouter()

  // ── Filter state ─────────────────────────────────────────────────────────────
  const filterForm = ref<Record<string, any>>({ id: '', name: '' })

  const filterItems = computed(() => [
    {
      label: 'ID',
      key: 'id',
      type: 'input',
      props: { placeholder: t('host.idPlaceholder'), clearable: true }
    },
    {
      label: t('host.name'),
      key: 'name',
      type: 'input',
      props: { placeholder: t('host.namePlaceholder'), clearable: true }
    }
  ])

  // ── Auto-register dialog state ────────────────────────────────────────────────
  const registerDialogVisible = ref(false)
  const agentTokenData = ref<AgentTokenResult | null>(null)
  // Token cache: reuse if not yet expired
  const cachedToken = ref<AgentTokenResult | null>(null)
  const cachedTokenExpires = ref<Date | null>(null)

  // ── useTable ─────────────────────────────────────────────────────────────────
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
      apiFn: fetchHostList,
      apiParams: {
        page: 1,
        page_size: 20,
        id: '',
        name: ''
      },
      paginationKey: {
        current: 'page',
        size: 'page_size'
      },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#', align: 'center' },
        {
          prop: 'id',
          label: t('host.id'),
          width: 80,
          align: 'center'
        },
        {
          prop: 'name',
          label: t('host.name'),
          align: 'center'
        },
        {
          prop: 'alias',
          label: t('host.alias'),
          align: 'center'
        },
        {
          prop: 'port',
          label: t('host.port'),
          width: 90,
          align: 'center'
        },
        {
          prop: 'remark',
          label: t('host.remark'),
          align: 'center',
          formatter: (row: HostItem) => h('span', {}, row.remark || '-')
        },
        {
          prop: 'created',
          label: t('host.createdAt'),
          width: 180,
          align: 'center',
          formatter: (row: HostItem) => formatDateTime(row.created)
        },
        {
          prop: 'action',
          label: t('host.operation'),
          width: 240,
          fixed: 'right',
          align: 'center',
          formatter: (row: HostItem) =>
            h('span', { style: 'display:inline-flex;gap:6px;' }, [
              h(
                ElButton,
                {
                  type: 'info',
                  size: 'small',
                  onClick: () => handlePing(row)
                },
                () => t('host.ping')
              ),
              h(
                ElButton,
                {
                  type: 'primary',
                  size: 'small',
                  onClick: () => toEdit(row)
                },
                () => t('host.edit')
              ),
              h(
                ElButton,
                {
                  type: 'danger',
                  size: 'small',
                  onClick: () => handleRemove(row)
                },
                () => t('host.delete')
              )
            ])
        }
      ]
    }
  })

  // ── Actions ───────────────────────────────────────────────────────────────────
  function handleSearch() {
    Object.assign(searchParams, {
      id: filterForm.value.id || '',
      name: filterForm.value.name || ''
    })
    getData()
  }

  function handleReset() {
    filterForm.value = { id: '', name: '' }
    resetSearchParams()
  }

  async function handlePing(row: HostItem) {
    try {
      await pingHost(row.id)
      ElMessage.success(t('host.pingSuccess'))
    } catch {
      ElMessage.error(t('host.pingFailed'))
    }
  }

  function toCreate() {
    router.push('/host/create')
  }

  function toEdit(row: HostItem) {
    router.push(`/host/edit/${row.id}`)
  }

  async function handleRemove(row: HostItem) {
    try {
      await ElMessageBox.confirm(t('host.confirmDelete'), t('host.confirmTitle'), {
        confirmButtonText: t('host.confirm'),
        cancelButtonText: t('host.cancel'),
        type: 'warning',
        center: true
      })
      await removeHost(row.id)
      ElMessage.success(t('host.deleteSuccess'))
      refreshRemove()
    } catch (err: any) {
      if (err !== 'cancel' && err?.message !== 'cancel') {
        // ElMessageBox cancel is caught here too — only show error for real failures
        if (err && err.message) {
          ElMessage.error(t('host.deleteFailed'))
        }
      }
    }
  }

  async function showAutoRegister() {
    registerDialogVisible.value = true
    agentTokenData.value = null

    // Reuse cached token if still valid
    const now = new Date()
    if (cachedToken.value && cachedTokenExpires.value && now < cachedTokenExpires.value) {
      agentTokenData.value = cachedToken.value
      return
    }

    try {
      const res = await generateAgentToken()
      // res may be wrapped by the http util; unwrap if needed
      const tokenData: AgentTokenResult = (res as any)?.data ?? res
      agentTokenData.value = tokenData
      cachedToken.value = tokenData
      cachedTokenExpires.value = new Date(tokenData.expires_at)
    } catch {
      ElMessage.error(t('host.tokenFailed'))
      registerDialogVisible.value = false
    }
  }

  function copyInstallCmd() {
    if (!agentTokenData.value?.install_cmd) return
    navigator.clipboard
      .writeText(agentTokenData.value.install_cmd)
      .then(() => ElMessage.success(t('host.copySuccess')))
      .catch(() => ElMessage.error(t('host.copyFailed')))
  }
</script>

<style scoped>
  .host-list-page {
    display: flex;
    flex-direction: column;
  }

  .install-label {
    font-size: 13px;
    color: var(--el-text-color-secondary);
    margin-bottom: 6px;
  }

  .install-pre {
    white-space: pre-wrap;
    word-break: break-all;
    font-family: monospace;
    font-size: 13px;
    background: var(--el-fill-color-light);
    padding: 12px;
    border-radius: 4px;
    margin: 0;
    max-height: 160px;
    overflow-y: auto;
  }
</style>
