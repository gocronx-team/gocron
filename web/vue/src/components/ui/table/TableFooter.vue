<script setup>
import { cn } from "@/lib/utils";

const props = defineProps({
  class: { type: null, required: false },
});
</script>

<template>
  <tfoot
    :class="
      cn('tw-border-t tw-bg-muted/50 tw-font-medium [&>tr]:last:tw-border-b-0', props.class)
    "
  >
    <slot />
  </tfoot>
</template>
