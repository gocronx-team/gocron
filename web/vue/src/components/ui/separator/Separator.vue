<script setup>
import { reactiveOmit } from "@vueuse/core";
import { Separator } from "reka-ui";
import { cn } from "@/lib/utils";

const props = defineProps({
  orientation: { type: String, required: false, default: "horizontal" },
  decorative: { type: Boolean, required: false, default: true },
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
  class: { type: null, required: false },
});

const delegatedProps = reactiveOmit(props, "class");
</script>

<template>
  <Separator
    v-bind="delegatedProps"
    :class="
      cn(
        'tw-shrink-0 tw-bg-border',
        props.orientation === 'horizontal' ? 'tw-h-px tw-w-full' : 'tw-w-px tw-h-full',
        props.class,
      )
    "
  />
</template>
