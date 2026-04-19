<template>
  <div class="template-list-page art-full-height">
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
          <span class="text-base font-medium">{{ t('menus.template.list') }}</span>
        </template>
        <template #right>
          <ElButton type="primary" @click="toCreate">{{ t('template.addTemplate') }}</ElButton>
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
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, h } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useRouter } from 'vue-router'
  import { ElButton, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchTemplateList,
    fetchTemplateRemove,
    type TemplateListItem
  } from '@/api/template'
  import { formatDateTime } from '@/utils/date'

  defineOptions({ name: 'TemplateList' })

  const { t } = useI18n()
  const router = useRouter()

  // ── Filter state ─────────────────────────────────────────────────────────────
  const filterForm = ref<Record<string, any>>({ name: '' })

  const filterItems = computed(() => [
    {
      label: t('template.name'),
      key: 'name',
      type: 'input',
      props: { placeholder: t('template.namePlaceholder'), clearable: true }
    }
  ])

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
      apiFn: fetchTemplateList,
      apiParams: {
        page: 1,
        page_size: 20,
        name: ''
      },
      paginationKey: {
        current: 'page',
        size: 'page_size'
      },
      columnsFactory: () => [
        {
          prop: 'id',
          label: t('template.id'),
          width: 80,
          align: 'center'
        },
        {
          prop: 'name',
          label: t('template.name'),
          minWidth: 160,
          align: 'center'
        },
        {
          prop: 'spec',
          label: t('template.spec'),
          minWidth: 160,
          align: 'center'
        },
        {
          prop: 'protocol',
          label: t('template.protocol'),
          width: 90,
          align: 'center',
          formatter: (row: TemplateListItem) =>
            h(
              ElTag,
              { type: row.protocol === 1 ? 'primary' : 'success', size: 'small' },
              () => (row.protocol === 1 ? t('template.protocolHttp') : t('template.protocolRpc'))
            )
        },
        {
          prop: 'created',
          label: t('template.created'),
          width: 180,
          align: 'center',
          formatter: (row: TemplateListItem) => formatDateTime(row.created ?? '')
        },
        {
          prop: 'action',
          label: t('template.actions'),
          width: 180,
          fixed: 'right',
          align: 'center',
          formatter: (row: TemplateListItem) =>
            h('span', { style: 'display:inline-flex;gap:6px;' }, [
              h(
                ElButton,
                {
                  type: 'primary',
                  size: 'small',
                  onClick: () => toEdit(row)
                },
                () => t('template.edit')
              ),
              h(
                ElButton,
                {
                  type: 'danger',
                  size: 'small',
                  onClick: () => handleRemove(row)
                },
                () => t('template.delete')
              )
            ])
        }
      ]
    }
  })

  // ── Actions ───────────────────────────────────────────────────────────────────
  function handleSearch() {
    Object.assign(searchParams, { name: filterForm.value.name || '' })
    getData()
  }

  function handleReset() {
    filterForm.value = { name: '' }
    resetSearchParams()
  }

  function toCreate() {
    router.push('/template/create')
  }

  function toEdit(row: TemplateListItem) {
    router.push(`/template/edit/${row.id}`)
  }

  async function handleRemove(row: TemplateListItem) {
    try {
      await ElMessageBox.confirm(
        t('template.confirmDelete', { name: row.name }),
        t('template.confirmTitle'),
        {
          confirmButtonText: t('template.confirm'),
          cancelButtonText: t('template.cancel'),
          type: 'warning',
          center: true
        }
      )
      await fetchTemplateRemove(row.id)
      ElMessage.success(t('template.deleteSuccess'))
      refreshRemove()
    } catch (err: any) {
      if (err !== 'cancel' && err?.message !== 'cancel') {
        if (err && err.message) {
          ElMessage.error(err.message)
        }
      }
    }
  }
</script>

<style scoped>
  .template-list-page {
    display: flex;
    flex-direction: column;
  }
</style>
