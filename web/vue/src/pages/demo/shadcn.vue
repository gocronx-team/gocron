<script setup>
import { h, ref, computed } from 'vue'
import { Moon, Sun, Trash2, Pencil } from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@/components/ui/dialog'
import { Switch } from '@/components/ui/switch'
import { Separator } from '@/components/ui/separator'
import { DataTable, DataTableColumnHeader } from '@/components/ui/data-table'
import { Toaster } from '@/components/ui/sonner'
import { toast } from 'vue-sonner'

import { useDarkMode } from '@/composables/useDarkMode'
import { useNotify } from '@/composables/useNotify'

// ===== Dark mode demo =====
const { isDark, toggle } = useDarkMode()

// ===== Notify demo =====
const notify = useNotify()
function showToastViaFacade() {
  notify.success('Hello from useNotify facade')
}
function showSonnerDirect() {
  toast('Direct vue-sonner toast', { description: 'This bypasses the facade.' })
}
async function showConfirm() {
  const ok = await notify.confirm('确认执行这个操作吗？', '提示')
  if (ok) notify.info('点了确认')
  else notify.warning('点了取消')
}

// ===== Switch facade backend demo =====
const useSonnerBackend = ref(localStorage.getItem('notify.backend') === 'sonner')
function updateBackend(v) {
  useSonnerBackend.value = v
  localStorage.setItem('notify.backend', v ? 'sonner' : 'element')
  notify.success(`Backend now: ${v ? 'sonner' : 'element'} (reload to re-init)`)
}

// ===== DataTable demo =====
const page = ref(1)
const pageSize = ref(10)
const loading = ref(false)
const selected = ref([])

const allTasks = Array.from({ length: 47 }, (_, i) => ({
  id: i + 1,
  name: `demo-task-${i + 1}`,
  spec: i % 3 === 0 ? '0 */5 * * * *' : '0 0 0 * * *',
  status: i % 4 === 0 ? 'failed' : 'success',
  runs: Math.floor(Math.random() * 1000)
}))

const data = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return allTasks.slice(start, start + pageSize.value)
})
const total = computed(() => allTasks.length)

const columns = [
  {
    accessorKey: 'id',
    header: ({ column }) => h(DataTableColumnHeader, { column, title: 'ID' }),
    size: 80
  },
  {
    accessorKey: 'name',
    header: ({ column }) => h(DataTableColumnHeader, { column, title: '任务名' })
  },
  {
    accessorKey: 'spec',
    header: 'Cron',
    cell: ({ row }) =>
      h(
        'code',
        { class: 'tw-font-mono tw-text-xs tw-text-muted-foreground' },
        row.getValue('spec')
      )
  },
  {
    accessorKey: 'status',
    header: ({ column }) => h(DataTableColumnHeader, { column, title: '状态' }),
    cell: ({ row }) => {
      const v = row.getValue('status')
      return h(
        Badge,
        { variant: v === 'success' ? 'secondary' : 'destructive' },
        () => v
      )
    }
  },
  {
    accessorKey: 'runs',
    header: ({ column }) => h(DataTableColumnHeader, { column, title: '执行次数' })
  },
  {
    id: 'actions',
    header: '操作',
    cell: ({ row }) =>
      h('div', { class: 'tw-flex tw-gap-1' }, [
        h(
          Button,
          {
            variant: 'ghost',
            size: 'icon-sm',
            onClick: () => notify.info(`编辑 #${row.original.id}`)
          },
          () => h(Pencil, { class: 'tw-size-4' })
        ),
        h(
          Button,
          {
            variant: 'ghost',
            size: 'icon-sm',
            onClick: () => notify.error(`删除 #${row.original.id}（仅演示）`)
          },
          () => h(Trash2, { class: 'tw-size-4 tw-text-destructive' })
        )
      ])
  }
]
</script>

<template>
  <div class="tw-container tw-mx-auto tw-p-6 tw-space-y-8">
    <div class="tw-flex tw-items-center tw-justify-between">
      <div>
        <h1 class="tw-text-3xl tw-font-bold tw-tracking-tight">shadcn-vue 组件演示</h1>
        <p class="tw-text-muted-foreground tw-mt-1">
          Phase 1 基础设施验证页，供后续 ticket 参考。
        </p>
      </div>
      <Button variant="outline" size="icon" @click="toggle">
        <Sun v-if="isDark" class="tw-size-4" />
        <Moon v-else class="tw-size-4" />
      </Button>
    </div>

    <Separator />

    <!-- Button Variants -->
    <Card>
      <CardHeader>
        <CardTitle>Button variants</CardTitle>
        <CardDescription>Default / Secondary / Destructive / Outline / Ghost / Link</CardDescription>
      </CardHeader>
      <CardContent class="tw-flex tw-flex-wrap tw-gap-2">
        <Button>Default</Button>
        <Button variant="secondary">Secondary</Button>
        <Button variant="destructive">Destructive</Button>
        <Button variant="outline">Outline</Button>
        <Button variant="ghost">Ghost</Button>
        <Button variant="link">Link</Button>
        <Button disabled>Disabled</Button>
      </CardContent>
    </Card>

    <!-- Inputs -->
    <Card>
      <CardHeader>
        <CardTitle>Input + Label</CardTitle>
      </CardHeader>
      <CardContent class="tw-space-y-3 tw-max-w-md">
        <div class="tw-space-y-1">
          <Label for="demo-input">Name</Label>
          <Input id="demo-input" placeholder="type something" />
        </div>
      </CardContent>
    </Card>

    <!-- Badge -->
    <Card>
      <CardHeader>
        <CardTitle>Badge</CardTitle>
      </CardHeader>
      <CardContent class="tw-flex tw-gap-2">
        <Badge>Default</Badge>
        <Badge variant="secondary">Secondary</Badge>
        <Badge variant="destructive">Destructive</Badge>
        <Badge variant="outline">Outline</Badge>
      </CardContent>
    </Card>

    <!-- Dialog -->
    <Card>
      <CardHeader>
        <CardTitle>Dialog</CardTitle>
      </CardHeader>
      <CardContent>
        <Dialog>
          <DialogTrigger as-child>
            <Button variant="outline">Open Dialog</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Are you sure?</DialogTitle>
              <DialogDescription>This is a shadcn-vue Dialog primitive.</DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button variant="outline">Cancel</Button>
              <Button>Confirm</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </CardContent>
    </Card>

    <!-- Notify facade -->
    <Card>
      <CardHeader>
        <CardTitle>Notify facade (useNotify)</CardTitle>
        <CardDescription>
          切换底层后刷新页面生效。默认 element（Element Plus Message），切到 sonner 使用 shadcn Toast。
        </CardDescription>
      </CardHeader>
      <CardContent class="tw-space-y-3">
        <div class="tw-flex tw-items-center tw-gap-3">
          <Switch :model-value="useSonnerBackend" @update:model-value="updateBackend" />
          <Label>use sonner backend</Label>
        </div>
        <div class="tw-flex tw-flex-wrap tw-gap-2">
          <Button size="sm" variant="outline" @click="showToastViaFacade">notify.success</Button>
          <Button size="sm" variant="outline" @click="showSonnerDirect">sonner direct</Button>
          <Button size="sm" variant="outline" @click="showConfirm">notify.confirm</Button>
        </div>
      </CardContent>
    </Card>

    <!-- DataTable -->
    <Card>
      <CardHeader>
        <CardTitle>DataTable</CardTitle>
        <CardDescription>
          TanStack Table + shadcn primitives. 支持排序、选择、分页。选中：{{ selected.length }}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <DataTable
          :columns="columns"
          :data="data"
          :loading="loading"
          :total="total"
          :page="page"
          :page-size="pageSize"
          @update:page="(v) => (page = v)"
          @update:page-size="(v) => (pageSize = v)"
          @update:selected="(v) => (selected = v)"
        />
      </CardContent>
    </Card>

    <!-- Toaster mount -->
    <Toaster position="top-right" />
  </div>
</template>
