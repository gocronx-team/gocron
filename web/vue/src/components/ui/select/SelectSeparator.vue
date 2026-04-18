<script setup>
import { reactiveOmit } from "@vueuse/core";
import { SelectSeparator } from "reka-ui";
import { cn } from "@/lib/utils";

const props = defineProps({
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
  class: { type: null, required: false },
});

const delegatedProps = reactiveOmit(props, "class");
</script>

<template>
  <SelectSeparator
    v-bind="delegatedProps"
    :class="cn('tw--mx-1 tw-my-1 tw-h-px tw-bg-muted', props.class)"
  />
</template>
