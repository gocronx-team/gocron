<script setup>
import { h, ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import { RefreshCw, Eye } from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { DataTable } from '@/components/ui/data-table'
import { useNotify } from '@/composables/useNotify'

import auditService from '@/api/audit'

const { t } = useI18n()
const notify = useNotify()

// ── State ─────────────────────────────────────────────────────────────────────
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const filterModule = ref('')
const filterAction = ref('')
const filterUsername = ref('')
const filterStartDate = ref('')
const filterEndDate = ref('')

// ── Detail dialog ─────────────────────────────────────────────────────────────
const detailOpen = ref(false)
const detailRows = ref([])

// ── Module / Action options ───────────────────────────────────────────────────
const moduleOptions = computed(() => [
  { value: '', label: t('message.all') },
  { value: 'task', label: t('audit.module_task') },
  { value: 'host', label: t('audit.module_host') },
  { value: 'user', label: t('audit.module_user') },
  { value: 'system', label: t('audit.module_system') }
])

const actionOptions = computed(() => [
  { value: '', label: t('message.all') },
  { value: 'create', label: t('audit.action_create') },
  { value: 'update', label: t('audit.action_update') },
  { value: 'delete', label: t('audit.action_delete') },
  { value: 'enable', label: t('audit.action_enable') },
  { value: 'disable', label: t('audit.action_disable') },
  { value: 'run', label: t('audit.action_run') },
  { value: 'batch-enable', label: t('audit.action_batch_enable') },
  { value: 'batch-disable', label: t('audit.action_batch_disable') },
  { value: 'batch-remove', label: t('audit.action_batch_remove') },
  { value: 'change-password', label: t('audit.action_change_password') },
  { value: 'reset-password', label: t('audit.action_reset_password') }
])

// ── Badge helpers ─────────────────────────────────────────────────────────────
function moduleVariant(module) {
  const map = {
    task: 'secondary',
    host: 'default',
    user: 'outline',
    system: 'destructive'
  }
  return map[module] || 'secondary'
}

function moduleLabel(module) {
  const found = moduleOptions.value.find(o => o.value === module)
  return found ? found.label : module
}

function actionVariant(action) {
  const map = {
    create: 'default',
    enable: 'default',
    'batch-enable': 'default',
    update: 'secondary',
    'change-password': 'secondary',
    'reset-password': 'secondary',
    delete: 'destructive',
    'batch-remove': 'destructive',
    disable: 'outline',
    'batch-disable': 'outline'
  }
  return map[action] || 'outline'
}

function actionLabel(action) {
  const found = actionOptions.value.find(o => o.value === action)
  return found ? found.label : action
}

// ── Columns ───────────────────────────────────────────────────────────────────
const columns = computed(() => [
  {
    accessorKey: 'created',
    header: t('system.loginTime'),
    size: 180,
    cell: ({ row }) => {
      const val = row.getValue('created')
      if (!val) return h('span', { class: 'tw-text-muted-foreground' }, '-')
      return h('span', dayjs(val).format('YYYY-MM-DD HH:mm:ss'))
    }
  },
  {
    accessorKey: 'username',
    header: t('user.username')
  },
  {
    accessorKey: 'module',
    header: t('audit.module'),
    size: 110,
    cell: ({ row }) => {
      const mod = row.getValue('module')
      return h(
        Badge,
        { variant: moduleVariant(mod), class: 'tw-whitespace-nowrap' },
        () => moduleLabel(mod)
      )
    }
  },
  {
    accessorKey: 'action',
    header: t('audit.action'),
    size: 130,
    cell: ({ row }) => {
      const act = row.getValue('action')
      return h(
        Badge,
        { variant: actionVariant(act), class: 'tw-whitespace-nowrap' },
        () => actionLabel(act)
      )
    }
  },
  {
    id: 'target',
    header: t('audit.target'),
    cell: ({ row }) => {
      const r = row.original
      return h('span', r.target_name || r.target_id || '-')
    }
  },
  {
    accessorKey: 'ip',
    header: t('system.loginIp'),
    size: 140
  },
  {
    id: 'detail',
    header: t('audit.detail'),
    size: 110,
    cell: ({ row }) => {
      if (!row.original.detail) return null
      return h(
        Button,
        {
          variant: 'ghost',
          size: 'sm',
          class: 'tw-h-7 tw-px-2',
          onClick: () => openDetail(row.original)
        },
        () => [
          h(Eye, { class: 'tw-size-3.5 tw-mr-1' }),
          t('taskLog.viewOutput')
        ]
      )
    }
  }
])

// ── Detail dialog logic ───────────────────────────────────────────────────────
function openDetail(row) {
  detailRows.value = (row.detail || '')
    .split('\n')
    .filter(Boolean)
    .map(line => {
      const parts = line.split(' \u2192 ')
      const fieldAndOld = (parts[0] || '').split(': ')
      return {
        field: fieldAndOld[0] || '',
        old: fieldAndOld.slice(1).join(': ') || '',
        newVal: parts.slice(1).join(' \u2192 ') || ''
      }
    })
  detailOpen.value = true
}

// ── Data fetching ─────────────────────────────────────────────────────────────
function loadData() {
  loading.value = true
  const params = {
    page: page.value,
    page_size: pageSize.value,
    module: filterModule.value,
    action: filterAction.value,
    username: filterUsername.value,
    start_date: filterStartDate.value,
    end_date: filterEndDate.value
  }
  auditService.list(params, (data) => {
    list.value = data.data ?? []
    total.value = data.total ?? 0
    loading.value = false
  })
}

function search() {
  page.value = 1
  loadData()
}

function reset() {
  filterModule.value = ''
  filterAction.value = ''
  filterUsername.value = ''
  filterStartDate.value = ''
  filterEndDate.value = ''
  page.value = 1
  loadData()
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

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="tw-p-6 tw-space-y-4">
    <Card>
      <CardHeader class="tw-pb-3">
        <div class="tw-flex tw-items-center tw-justify-between">
          <CardTitle>{{ t('audit.log') }}</CardTitle>
          <Button
            variant="outline"
            size="sm"
            :disabled="loading"
            @click="loadData"
          >
            <RefreshCw
              class="tw-size-4 tw-mr-2"
              :class="{ 'tw-animate-spin': loading }"
            />
            {{ t('common.refresh') }}
          </Button>
        </div>
      </CardHeader>
      <CardContent class="tw-space-y-4">
        <!-- Filters -->
        <div class="tw-flex tw-flex-wrap tw-gap-2 tw-items-center">
          <!-- Module filter -->
          <Select v-model="filterModule">
            <SelectTrigger class="tw-w-36">
              <SelectValue :placeholder="t('audit.module')" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem
                v-for="opt in moduleOptions"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.label }}
              </SelectItem>
            </SelectContent>
          </Select>

          <!-- Action filter -->
          <Select v-model="filterAction">
            <SelectTrigger class="tw-w-44">
              <SelectValue :placeholder="t('audit.action')" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem
                v-for="opt in actionOptions"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.label }}
              </SelectItem>
            </SelectContent>
          </Select>

          <!-- Username filter -->
          <Input
            v-model="filterUsername"
            :placeholder="t('user.username')"
            class="tw-w-36"
            @keyup.enter="search"
          />

          <!-- Date range: start -->
          <Input
            v-model="filterStartDate"
            type="date"
            class="tw-w-36"
            :title="t('common.date')"
          />
          <span class="tw-text-muted-foreground tw-text-sm">-</span>
          <!-- Date range: end -->
          <Input
            v-model="filterEndDate"
            type="date"
            class="tw-w-36"
            :title="t('common.date')"
          />

          <Button @click="search">
            {{ t('common.search') }}
          </Button>
          <Button variant="outline" @click="reset">
            {{ t('common.reset') }}
          </Button>
        </div>

        <!-- Data table -->
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

    <!-- Detail Dialog -->
    <Dialog v-model:open="detailOpen">
      <DialogContent class="tw-max-w-2xl tw-overflow-y-auto tw-max-h-[80vh]">
        <DialogHeader>
          <DialogTitle>{{ t('audit.detail') }}</DialogTitle>
        </DialogHeader>

        <div class="tw-overflow-x-auto">
          <table class="tw-w-full tw-text-sm tw-border-collapse">
            <thead>
              <tr class="tw-border-b tw-border-border tw-bg-muted/50">
                <th class="tw-text-left tw-px-3 tw-py-2 tw-font-medium tw-w-36">
                  Field
                </th>
                <th class="tw-text-left tw-px-3 tw-py-2 tw-font-medium">
                  Before
                </th>
                <th class="tw-w-8 tw-text-center tw-px-1 tw-py-2 tw-font-medium" />
                <th class="tw-text-left tw-px-3 tw-py-2 tw-font-medium">
                  After
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, idx) in detailRows"
                :key="idx"
                class="tw-border-b tw-border-border last:tw-border-0 hover:tw-bg-muted/30"
              >
                <td class="tw-px-3 tw-py-2 tw-font-medium tw-text-muted-foreground tw-break-all">
                  {{ row.field }}
                </td>
                <td class="tw-px-3 tw-py-2 tw-break-all">
                  {{ row.old }}
                </td>
                <td class="tw-text-center tw-px-1 tw-py-2 tw-text-muted-foreground">
                  &rarr;
                </td>
                <td class="tw-px-3 tw-py-2 tw-break-all">
                  {{ row.newVal }}
                </td>
              </tr>
              <tr v-if="detailRows.length === 0">
                <td
                  colspan="4"
                  class="tw-px-3 tw-py-4 tw-text-center tw-text-muted-foreground"
                >
                  No detail available
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
