<script setup>
import { cn } from "@/lib/utils";

const props = defineProps({
  class: { type: null, required: false },
});
</script>

<template>
  <tbody :class="cn('[&_tr:last-child]:tw-border-0', props.class)">
    <slot />
  </tbody>
</template>
