<script setup>
import { cn } from "@/lib/utils";

const props = defineProps({
  class: { type: null, required: false },
});
</script>

<template>
  <span :class="cn('tw-ml-auto tw-text-xs tw-tracking-widest tw-opacity-60', props.class)">
    <slot />
  </span>
</template>
