<script setup>
import { h, ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import {
  RefreshCw,
  Plus,
  MoreHorizontal,
  ShieldCheck,
  User
} from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { DataTable } from '@/components/ui/data-table'
import { useNotify } from '@/composables/useNotify'

import userService from '@/api/user'
import { useUserStore } from '@/stores/user'

const { t } = useI18n()
const router = useRouter()
const notify = useNotify()
const userStore = useUserStore()

// ── State ─────────────────────────────────────────────────────────────────────
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const isAdmin = computed(() => userStore.isAdmin)

// ── Data fetching ─────────────────────────────────────────────────────────────
function loadData() {
  loading.value = true
  userService.list({ page: page.value, page_size: pageSize.value }, (data) => {
    list.value = data.data ?? []
    total.value = data.total ?? 0
    loading.value = false
  })
}

function onPageChange(newPage) {
  page.value = newPage
  loadData()
}

function onPageSizeChange(newSize) {
  pageSize.value = newSize
  page.value = 1
  loadData()
}

function refresh() {
  loadData()
  notify.success(t('message.refreshSuccess'))
}

// ── Navigation ────────────────────────────────────────────────────────────────
function goCreate() {
  router.push('/user/create')
}

function goEdit(id) {
  router.push(`/user/edit/${id}`)
}

function goEditPassword(id) {
  router.push(`/user/edit-password/${id}`)
}

// ── Actions ───────────────────────────────────────────────────────────────────
function enable(id) {
  userService.enable(id, () => {
    loadData()
  })
}

function disable(id) {
  userService.disable(id, () => {
    loadData()
  })
}

async function remove(id) {
  const ok = await notify.confirm(t('message.confirmDeleteUser'))
  if (!ok) return
  userService.remove(id, () => {
    loadData()
  })
}

// ── Columns ───────────────────────────────────────────────────────────────────
const columns = computed(() => {
  const cols = [
    {
      accessorKey: 'id',
      header: 'ID',
      size: 80
    },
    {
      accessorKey: 'name',
      header: t('user.username')
    },
    {
      accessorKey: 'email',
      header: t('user.email')
    },
    {
      accessorKey: 'is_admin',
      header: t('user.role'),
      size: 130,
      cell: ({ row }) => {
        const admin = row.getValue('is_admin')
        if (admin === 1) {
          return h(
            Badge,
            { variant: 'default', class: 'tw-gap-1 tw-whitespace-nowrap' },
            () => [
              h(ShieldCheck, { class: 'tw-size-3' }),
              t('user.admin')
            ]
          )
        }
        return h(
          Badge,
          { variant: 'secondary', class: 'tw-gap-1 tw-whitespace-nowrap' },
          () => [
            h(User, { class: 'tw-size-3' }),
            t('user.normalUser')
          ]
        )
      }
    },
    {
      accessorKey: 'status',
      header: t('common.status'),
      size: 110,
      cell: ({ row }) => {
        const status = row.getValue('status')
        return h(
          Badge,
          { variant: status === 1 ? 'secondary' : 'outline' },
          () => status === 1 ? t('common.enabled') : t('common.disabled')
        )
      }
    }
  ]

  if (isAdmin.value) {
    cols.push({
      id: 'actions',
      header: t('common.operation'),
      size: 80,
      cell: ({ row }) => {
        const user = row.original
        return h(
          DropdownMenu,
          {},
          {
            default: () => [
              h(
                DropdownMenuTrigger,
                { asChild: true },
                () => h(
                  Button,
                  { variant: 'ghost', size: 'icon' },
                  () => h(MoreHorizontal, { class: 'tw-size-4' })
                )
              ),
              h(
                DropdownMenuContent,
                { align: 'end' },
                () => [
                  h(
                    DropdownMenuItem,
                    { onClick: () => goEdit(user.id) },
                    () => t('common.edit')
                  ),
                  h(
                    DropdownMenuItem,
                    { onClick: () => goEditPassword(user.id) },
                    () => t('user.changePassword')
                  ),
                  user.status === 1
                    ? h(
                        DropdownMenuItem,
                        { onClick: () => disable(user.id) },
                        () => t('common.disabled')
                      )
                    : h(
                        DropdownMenuItem,
                        { onClick: () => enable(user.id) },
                        () => t('common.enabled')
                      ),
                  h(DropdownMenuSeparator),
                  h(
                    DropdownMenuItem,
                    {
                      class: 'tw-text-destructive focus:tw-text-destructive',
                      onClick: () => remove(user.id)
                    },
                    () => t('common.delete')
                  )
                ]
              )
            ]
          }
        )
      }
    })
  }

  return cols
})

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="tw-p-6 tw-space-y-4">
    <Card>
      <CardHeader class="tw-pb-3">
        <div class="tw-flex tw-items-center tw-justify-between">
          <CardTitle>{{ t('user.list') }}</CardTitle>
          <div class="tw-flex tw-gap-2">
            <Button
              variant="outline"
              size="sm"
              :disabled="loading"
              @click="refresh"
            >
              <RefreshCw
                class="tw-size-4 tw-mr-2"
                :class="{ 'tw-animate-spin': loading }"
              />
              {{ t('common.refresh') }}
            </Button>
            <Button
              v-if="isAdmin"
              size="sm"
              @click="goCreate"
            >
              <Plus class="tw-size-4 tw-mr-1" />
              {{ t('user.createNew') }}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent class="tw-pt-0">
        <DataTable
          :columns="columns"
          :data="list"
          :loading="loading"
          :total="total"
          :page="page"
          :page-size="pageSize"
          @update:page="onPageChange"
          @update:page-size="onPageSizeChange"
        />
      </CardContent>
    </Card>
  </div>
</template>
