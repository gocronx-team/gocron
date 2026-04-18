<script setup>
import { cn } from "@/lib/utils";

const props = defineProps({
  class: { type: null, required: false },
});
</script>

<template>
  <div :class="cn('tw-flex tw-flex-col tw-gap-y-1.5 tw-p-6', props.class)">
    <slot />
  </div>
</template>
