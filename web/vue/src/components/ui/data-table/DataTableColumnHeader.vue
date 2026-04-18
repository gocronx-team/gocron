<script setup>
import { ArrowDown, ArrowUp, ChevronsUpDown } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

/**
 * 可排序列头，用在 column 的 header 里：
 *   header: ({ column }) => h(DataTableColumnHeader, { column, title: '任务名' })
 */
const props = defineProps({
  column: { type: Object, required: true },
  title: { type: String, required: true },
  class: { type: String, default: '' }
})

function toggleSort() {
  const current = props.column.getIsSorted()
  if (current === false) props.column.toggleSorting(false)
  else if (current === 'asc') props.column.toggleSorting(true)
  else props.column.clearSorting()
}
</script>

<template>
  <div :class="cn('tw-flex tw-items-center', props.class)">
    <Button
      v-if="column.getCanSort()"
      variant="ghost"
      size="sm"
      class="tw--ml-3 tw-h-8 tw-data-[state=open]:tw-bg-accent"
      @click="toggleSort"
    >
      <span>{{ title }}</span>
      <ArrowDown v-if="column.getIsSorted() === 'desc'" class="tw-ml-2 tw-size-4" />
      <ArrowUp v-else-if="column.getIsSorted() === 'asc'" class="tw-ml-2 tw-size-4" />
      <ChevronsUpDown v-else class="tw-ml-2 tw-size-4 tw-opacity-50" />
    </Button>
    <div v-else>{{ title }}</div>
  </div>
</template>
