<script setup>
import { cn } from "@/lib/utils";

const props = defineProps({
  class: { type: null, required: false },
});
</script>

<template>
  <div :class="cn('tw-animate-pulse tw-rounded-md tw-bg-muted', props.class)" />
</template>
