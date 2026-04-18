<script setup>
import { cn } from "@/lib/utils";

const props = defineProps({
  class: { type: null, required: false },
});
</script>

<template>
  <h3
    :class="
      cn('tw-text-2xl tw-font-semibold tw-leading-none tw-tracking-tight', props.class)
    "
  >
    <slot />
  </h3>
</template>
