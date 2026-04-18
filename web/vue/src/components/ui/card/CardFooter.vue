<script setup>
import { cn } from "@/lib/utils";

const props = defineProps({
  class: { type: null, required: false },
});
</script>

<template>
  <div :class="cn('tw-flex tw-items-center tw-p-6 tw-pt-0', props.class)">
    <slot />
  </div>
</template>
