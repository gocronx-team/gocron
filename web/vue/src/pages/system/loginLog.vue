<script setup>
import { h, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import { RefreshCw } from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle
} from '@/components/ui/card'
import { DataTable } from '@/components/ui/data-table'
import { useNotify } from '@/composables/useNotify'

import systemService from '@/api/system'

const { t } = useI18n()
const notify = useNotify()

// ── State ────────────────────────────────────────────────────────────────────
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

// ── Columns ──────────────────────────────────────────────────────────────────
const columns = [
  {
    accessorKey: 'id',
    header: 'ID',
    size: 80
  },
  {
    accessorKey: 'username',
    header: t('user.username')
  },
  {
    accessorKey: 'ip',
    header: t('system.loginIp')
  },
  {
    accessorKey: 'created',
    header: t('system.loginTime'),
    cell: ({ row }) => {
      const val = row.getValue('created')
      if (!val) return h('span', { class: 'tw-text-muted-foreground' }, '-')
      return dayjs(val).format('YYYY-MM-DD HH:mm:ss')
    }
  }
]

// ── Data fetching ─────────────────────────────────────────────────────────────
function loadData() {
  loading.value = true
  const params = {
    page: page.value,
    page_size: pageSize.value
  }
  systemService.loginLogList(params, (data) => {
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

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="tw-p-6 tw-space-y-4">
    <Card>
      <CardHeader class="tw-pb-3">
        <div class="tw-flex tw-items-center tw-justify-between">
          <CardTitle>{{ t('system.loginLog') }}</CardTitle>
          <Button
            variant="outline"
            size="sm"
            :disabled="loading"
            @click="loadData"
          >
            <RefreshCw class="tw-size-4 tw-mr-2" :class="{ 'tw-animate-spin': loading }" />
            {{ t('common.refresh') }}
          </Button>
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
