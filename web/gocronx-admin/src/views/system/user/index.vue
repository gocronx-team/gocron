<!-- 用户管理 -->
<template>
  <div class="user-page art-full-height">
    <!-- Table card -->
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader :loading="loading" v-model:columns="columnChecks" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <span class="text-base font-medium">{{ t('menus.system.user') }}</span>
            <ElButton type="primary" @click="toEdit(null)" v-ripple>
              {{ t('user.addUser') }}
            </ElButton>
          </ElSpace>
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
  import { h } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useRouter } from 'vue-router'
  import { ElTag, ElButton, ElMessageBox } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchUserList,
    fetchUserRemove,
    fetchUserEnable,
    fetchUserDisable,
    type UserListItem
  } from '@/api/user'
  import { formatDateTime } from '@/utils/date'

  defineOptions({ name: 'UserManage' })

  const { t } = useI18n()
  const router = useRouter()

  // ── useTable ──────────────────────────────────────────────────────────────
  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    refreshData,
    handleSizeChange,
    handleCurrentChange
  } = useTable({
    core: {
      apiFn: fetchUserList,
      apiParams: {
        page: 1,
        page_size: 20
      },
      paginationKey: {
        current: 'page',
        size: 'page_size'
      },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#', align: 'center' },
        {
          prop: 'name',
          label: t('user.name'),
          align: 'center'
        },
        {
          prop: 'email',
          label: t('user.email'),
          align: 'center'
        },
        {
          prop: 'is_admin',
          label: t('user.role'),
          width: 110,
          align: 'center',
          formatter: (row: UserListItem) =>
            h(
              ElTag,
              { type: row.is_admin === 1 ? 'danger' : 'info', size: 'small' },
              () => (row.is_admin === 1 ? t('user.admin') : t('user.normalUser'))
            )
        },
        {
          prop: 'status',
          label: t('user.status'),
          width: 110,
          align: 'center',
          formatter: (row: UserListItem) =>
            h(
              ElTag,
              { type: row.status === 1 ? 'success' : 'warning', size: 'small' },
              () => (row.status === 1 ? t('user.enabled') : t('user.disabled'))
            )
        },
        {
          prop: 'created',
          label: t('user.createdAt'),
          width: 180,
          align: 'center',
          formatter: (row: UserListItem) => formatDateTime(row.created)
        },
        {
          prop: 'operation',
          label: t('user.actions'),
          width: 240,
          fixed: 'right',
          align: 'center',
          formatter: (row: UserListItem) =>
            h('div', { class: 'flex justify-center gap-1 flex-wrap' }, [
              // Edit
              h(
                ElButton,
                { type: 'primary', size: 'small', onClick: () => toEdit(row) },
                () => t('user.edit')
              ),
              // Change password
              h(
                ElButton,
                { type: 'success', size: 'small', onClick: () => toEditPassword(row) },
                () => t('user.changePassword')
              ),
              // Enable / Disable toggle
              row.status === 1
                ? h(
                    ElButton,
                    { type: 'warning', size: 'small', onClick: () => toggleStatus(row) },
                    () => t('user.disable')
                  )
                : h(
                    ElButton,
                    { type: 'info', size: 'small', onClick: () => toggleStatus(row) },
                    () => t('user.enable')
                  ),
              // Delete
              h(
                ElButton,
                { type: 'danger', size: 'small', onClick: () => removeUser(row) },
                () => t('user.delete')
              )
            ])
        }
      ]
    }
  })

  // ── Navigation ────────────────────────────────────────────────────────────

  function toEdit(row: UserListItem | null) {
    if (row === null) {
      router.push('/system/user/edit/0')
    } else {
      router.push(`/system/user/edit/${row.id}`)
    }
  }

  function toEditPassword(row: UserListItem) {
    router.push(`/system/user/edit-password/${row.id}`)
  }

  // ── Enable / Disable ──────────────────────────────────────────────────────

  async function toggleStatus(row: UserListItem) {
    try {
      if (row.status === 1) {
        await fetchUserDisable(row.id)
        ElMessage.success(t('user.disableSuccess'))
      } else {
        await fetchUserEnable(row.id)
        ElMessage.success(t('user.enableSuccess'))
      }
      refreshData()
    } catch {
      // error toast handled by http interceptor
    }
  }

  // ── Delete ────────────────────────────────────────────────────────────────

  function removeUser(row: UserListItem) {
    ElMessageBox.confirm(t('user.confirmDelete', { name: row.name }), t('common.tips'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
      .then(async () => {
        try {
          await fetchUserRemove(row.id)
          ElMessage.success(t('user.deleteSuccess'))
          refreshData()
        } catch {
          // error toast handled by http interceptor
        }
      })
      .catch(() => {})
  }
</script>

<style scoped>
  .user-page {
    display: flex;
    flex-direction: column;
  }
</style>
