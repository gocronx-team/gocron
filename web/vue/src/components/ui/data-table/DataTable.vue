<script setup>
import { computed, ref, watch } from 'vue'
import {
  getCoreRowModel,
  getSortedRowModel,
  useVueTable
} from '@tanstack/vue-table'
import { FlexRender } from '@tanstack/vue-table'

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import DataTablePagination from './DataTablePagination.vue'

/**
 * 通用 DataTable，API 参考老 el-table 常用能力。
 *
 * props:
 *   columns      — TanStack ColumnDef[]（具体写法见 tanstack/vue-table 文档）
 *   data         — row[]
 *   loading      — 加载中
 *   total        — 总行数（外部分页）
 *   page         — 当前页（1-based），不传则不分页
 *   pageSize     — 每页行数
 *   pageSizes    — 分页 size 可选值，默认 [10, 20, 50, 100]
 *   selectable   — 开启多选（首列加 checkbox）
 *   emptyText    — 无数据提示
 *   rowKey       — 行唯一键字段名，用于 :key 和选择追踪，默认 'id'
 *
 * emits:
 *   update:page       — 切页
 *   update:pageSize   — 改每页数量
 *   update:selected   — 选择变化，payload: row[]
 */
const props = defineProps({
  columns: { type: Array, required: true },
  data: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
  total: { type: Number, default: 0 },
  page: { type: Number, default: 0 },
  pageSize: { type: Number, default: 20 },
  pageSizes: { type: Array, default: () => [10, 20, 50, 100] },
  selectable: { type: Boolean, default: false },
  emptyText: { type: String, default: '' },
  rowKey: { type: String, default: 'id' },
  class: { type: String, default: '' }
})

const emit = defineEmits(['update:page', 'update:pageSize', 'update:selected'])

const sorting = ref([])
const rowSelection = ref({})

// 每次 data 变化时，如果选择里的 row 已经不存在，清理掉
watch(
  () => props.data,
  () => {
    const keys = new Set(props.data.map((row, idx) => row[props.rowKey] ?? idx))
    for (const k of Object.keys(rowSelection.value)) {
      if (!keys.has(k)) delete rowSelection.value[k]
    }
  }
)

watch(
  rowSelection,
  () => {
    const selected = props.data.filter(
      (row, idx) => rowSelection.value[row[props.rowKey] ?? idx]
    )
    emit('update:selected', selected)
  },
  { deep: true }
)

const table = useVueTable({
  get data() {
    return props.data
  },
  get columns() {
    return props.columns
  },
  state: {
    get sorting() {
      return sorting.value
    },
    get rowSelection() {
      return rowSelection.value
    }
  },
  getRowId: (row, idx) => String(row[props.rowKey] ?? idx),
  getCoreRowModel: getCoreRowModel(),
  getSortedRowModel: getSortedRowModel(),
  onSortingChange: updater => {
    sorting.value = typeof updater === 'function' ? updater(sorting.value) : updater
  },
  onRowSelectionChange: updater => {
    rowSelection.value =
      typeof updater === 'function' ? updater(rowSelection.value) : updater
  },
  enableRowSelection: () => props.selectable
})

const showPagination = computed(() => props.page > 0 && props.total > 0)
const columnCount = computed(() => props.columns.length)

function handlePageChange(newPage) {
  emit('update:page', newPage)
}

function handlePageSizeChange(newSize) {
  emit('update:pageSize', newSize)
}

defineExpose({ table })
</script>

<template>
  <div :class="cn('tw-space-y-3', props.class)">
    <div class="tw-rounded-md tw-border tw-border-border tw-bg-background">
      <Table>
        <TableHeader>
          <TableRow v-for="headerGroup in table.getHeaderGroups()" :key="headerGroup.id">
            <TableHead
              v-for="header in headerGroup.headers"
              :key="header.id"
              :style="{ width: header.getSize() !== 150 ? `${header.getSize()}px` : undefined }"
            >
              <FlexRender
                v-if="!header.isPlaceholder"
                :render="header.column.columnDef.header"
                :props="header.getContext()"
              />
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <!-- Loading -->
          <template v-if="loading">
            <TableRow v-for="n in 5" :key="`skeleton-${n}`">
              <TableCell v-for="col in columnCount" :key="col">
                <Skeleton class="tw-h-4 tw-w-full" />
              </TableCell>
            </TableRow>
          </template>
          <!-- Data -->
          <template v-else-if="table.getRowModel().rows.length">
            <TableRow
              v-for="row in table.getRowModel().rows"
              :key="row.id"
              :data-state="row.getIsSelected() ? 'selected' : undefined"
            >
              <TableCell v-for="cell in row.getVisibleCells()" :key="cell.id">
                <FlexRender
                  :render="cell.column.columnDef.cell"
                  :props="cell.getContext()"
                />
              </TableCell>
            </TableRow>
          </template>
          <!-- Empty -->
          <TableRow v-else>
            <TableCell
              :colspan="columnCount"
              class="tw-h-24 tw-text-center tw-text-muted-foreground"
            >
              {{ emptyText || 'No data' }}
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
    <DataTablePagination
      v-if="showPagination"
      :page="page"
      :page-size="pageSize"
      :total="total"
      :page-sizes="pageSizes"
      @update:page="handlePageChange"
      @update:page-size="handlePageSizeChange"
    />
  </div>
</template>
