<script setup>
import { cn } from "@/lib/utils";

const props = defineProps({
  class: { type: null, required: false },
});
</script>

<template>
  <caption :class="cn('tw-mt-4 tw-text-sm tw-text-muted-foreground', props.class)">
    <slot />
  </caption>
</template>
