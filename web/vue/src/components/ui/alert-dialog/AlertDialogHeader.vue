<script setup>
import { cn } from "@/lib/utils";

const props = defineProps({
  class: { type: null, required: false },
});
</script>

<template>
  <div
    :class="cn('tw-flex tw-flex-col tw-gap-y-2 tw-text-center sm:tw-text-left', props.class)"
  >
    <slot />
  </div>
</template>
