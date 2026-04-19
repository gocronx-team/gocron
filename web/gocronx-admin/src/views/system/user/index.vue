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
  import { ElTag, ElButton, ElSwitch, ElMessageBox } from 'element-plus'
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
          width: 100,
          align: 'center',
          formatter: (row: UserListItem) =>
            h(ElSwitch, {
              modelValue: row.status === 1,
              activeValue: true,
              inactiveValue: false,
              'onUpdate:modelValue': (val: string | number | boolean) =>
                handleStatusToggle(row, Boolean(val))
            })
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
          width: 180,
          fixed: 'right',
          align: 'center',
          formatter: (row: UserListItem) =>
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
                  () => t('user.edit')
                ),
                h(
                  ElButton,
                  {
                    type: 'success',
                    size: 'small',
                    style: 'margin:0',
                    onClick: () => toEditPassword(row)
                  },
                  () => t('user.changePassword')
                ),
                h(
                  ElButton,
                  {
                    type: 'danger',
                    size: 'small',
                    style: 'margin:0;grid-column:1 / -1',
                    onClick: () => removeUser(row)
                  },
                  () => t('user.delete')
                )
              ]
            )
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

  async function handleStatusToggle(row: UserListItem, enabled: boolean) {
    try {
      if (enabled) {
        await fetchUserEnable(row.id)
        ElMessage.success(t('user.enableSuccess'))
      } else {
        await fetchUserDisable(row.id)
        ElMessage.success(t('user.disableSuccess'))
      }
      refreshData()
    } catch {
      // revert optimistic switch state on failure
      refreshData()
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
