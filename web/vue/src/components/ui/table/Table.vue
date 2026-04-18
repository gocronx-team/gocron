<script setup>
import { cn } from "@/lib/utils";

const props = defineProps({
  class: { type: null, required: false },
});
</script>

<template>
  <div class="tw-relative tw-w-full tw-overflow-auto">
    <table :class="cn('tw-w-full tw-caption-bottom tw-text-sm', props.class)">
      <slot />
    </table>
  </div>
</template>
