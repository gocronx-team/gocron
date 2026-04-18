<script setup>
import { reactiveOmit } from "@vueuse/core";
import { PinInputInput, useForwardProps } from "reka-ui";
import { cn } from "@/lib/utils";

const props = defineProps({
  index: { type: Number, required: true },
  disabled: { type: Boolean, required: false },
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
  class: { type: null, required: false },
});

const delegatedProps = reactiveOmit(props, "class");

const forwardedProps = useForwardProps(delegatedProps);
</script>

<template>
  <PinInputInput
    v-bind="forwardedProps"
    :class="
      cn(
        'tw-relative tw-text-center focus:tw-outline-none focus:tw-ring-2 focus:tw-ring-ring focus:tw-relative focus:tw-z-10 tw-flex tw-h-10 tw-w-10 tw-items-center tw-justify-center tw-border-y tw-border-r tw-border-input tw-text-sm tw-transition-all first:tw-rounded-l-md first:tw-border-l last:tw-rounded-r-md',
        props.class,
      )
    "
  />
</template>
