<template>
  <div class="login-log-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData" />

      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { fetchLoginLogList, type LoginLogItem } from '@/api/login-log'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'LoginLog' })

  const { t } = useI18n()

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchLoginLogList,
      apiParams: {
        page: 1,
        page_size: 20
      },
      // gocron API uses page / page_size instead of current / size
      paginationKey: {
        current: 'page',
        size: 'page_size'
      },
      columnsFactory: () => [
        { type: 'index', width: 60, label: t('loginLog.index') },
        { prop: 'username', label: t('loginLog.username') },
        { prop: 'ip', label: t('loginLog.ip') },
        {
          prop: 'created',
          label: t('loginLog.loginTime'),
          formatter: (row: LoginLogItem) => row.created
        }
      ]
    }
  })
</script>
