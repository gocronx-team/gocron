<script setup>
import { cn } from "@/lib/utils";

const props = defineProps({
  class: { type: null, required: false },
});
</script>

<template>
  <div
    :class="
      cn(
        'tw-rounded-lg tw-border tw-bg-card tw-text-card-foreground tw-shadow-sm',
        props.class,
      )
    "
  >
    <slot />
  </div>
</template>
