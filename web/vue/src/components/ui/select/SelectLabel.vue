<script setup>
import { SelectLabel } from "reka-ui";
import { cn } from "@/lib/utils";

const props = defineProps({
  for: { type: String, required: false },
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
  class: { type: null, required: false },
});
</script>

<template>
  <SelectLabel
    :class="cn('tw-py-1.5 tw-pl-8 tw-pr-2 tw-text-sm tw-font-semibold', props.class)"
  >
    <slot />
  </SelectLabel>
</template>
