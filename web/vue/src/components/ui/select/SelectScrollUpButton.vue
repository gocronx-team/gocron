<script setup>
import { reactiveOmit } from "@vueuse/core";
import { ChevronUp } from "lucide-vue-next";
import { SelectScrollUpButton, useForwardProps } from "reka-ui";
import { cn } from "@/lib/utils";

const props = defineProps({
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
  class: { type: null, required: false },
});

const delegatedProps = reactiveOmit(props, "class");

const forwardedProps = useForwardProps(delegatedProps);
</script>

<template>
  <SelectScrollUpButton
    v-bind="forwardedProps"
    :class="
      cn('tw-flex tw-cursor-default tw-items-center tw-justify-center tw-py-1', props.class)
    "
  >
    <slot>
      <ChevronUp class=""tw-h-4 tw-w-4"" />
    </slot>
  </SelectScrollUpButton>
</template>
