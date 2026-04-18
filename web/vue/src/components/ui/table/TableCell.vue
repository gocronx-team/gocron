<script setup>
import { cn } from "@/lib/utils";

const props = defineProps({
  class: { type: null, required: false },
});
</script>

<template>
  <td
    :class="cn('tw-p-4 tw-align-middle [&:has([role=checkbox])]:tw-pr-0', props.class)"
  >
    <slot />
  </td>
</template>
