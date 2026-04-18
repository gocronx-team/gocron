<script setup>
import { computed } from 'vue'
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select'

const props = defineProps({
  page: { type: Number, required: true },
  pageSize: { type: Number, required: true },
  total: { type: Number, required: true },
  pageSizes: { type: Array, default: () => [10, 20, 50, 100] }
})
const emit = defineEmits(['update:page', 'update:pageSize'])

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const canPrev = computed(() => props.page > 1)
const canNext = computed(() => props.page < totalPages.value)

function goto(p) {
  const next = Math.max(1, Math.min(totalPages.value, p))
  if (next !== props.page) emit('update:page', next)
}
function changeSize(v) {
  emit('update:pageSize', Number(v))
}
</script>

<template>
  <div
    class="tw-flex tw-items-center tw-justify-between tw-gap-4 tw-flex-wrap"
  >
    <div class="tw-text-sm tw-text-muted-foreground">
      Total {{ total }} · Page {{ page }} / {{ totalPages }}
    </div>
    <div class="tw-flex tw-items-center tw-gap-4">
      <div class="tw-flex tw-items-center tw-gap-2">
        <span class="tw-text-sm tw-text-muted-foreground">Rows per page</span>
        <Select :model-value="String(pageSize)" @update:model-value="changeSize">
          <SelectTrigger class="tw-h-8 tw-w-[70px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem v-for="s in pageSizes" :key="s" :value="String(s)">
              {{ s }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div class="tw-flex tw-items-center tw-gap-1">
        <Button
          variant="outline"
          size="icon-sm"
          :disabled="!canPrev"
          @click="goto(1)"
        >
          <ChevronsLeft class="tw-size-4" />
        </Button>
        <Button
          variant="outline"
          size="icon-sm"
          :disabled="!canPrev"
          @click="goto(page - 1)"
        >
          <ChevronLeft class="tw-size-4" />
        </Button>
        <Button
          variant="outline"
          size="icon-sm"
          :disabled="!canNext"
          @click="goto(page + 1)"
        >
          <ChevronRight class="tw-size-4" />
        </Button>
        <Button
          variant="outline"
          size="icon-sm"
          :disabled="!canNext"
          @click="goto(totalPages)"
        >
          <ChevronsRight class="tw-size-4" />
        </Button>
      </div>
    </div>
  </div>
</template>
