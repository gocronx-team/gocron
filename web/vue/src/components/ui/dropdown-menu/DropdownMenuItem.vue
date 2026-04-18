<script setup>
import { reactiveOmit } from "@vueuse/core";
import { DropdownMenuItem, useForwardProps } from "reka-ui";
import { cn } from "@/lib/utils";

const props = defineProps({
  disabled: { type: Boolean, required: false },
  textValue: { type: String, required: false },
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
  class: { type: null, required: false },
  inset: { type: Boolean, required: false },
});

const delegatedProps = reactiveOmit(props, "class");

const forwardedProps = useForwardProps(delegatedProps);
</script>

<template>
  <DropdownMenuItem
    v-bind="forwardedProps"
    :class="
      cn(
        'tw-relative tw-flex tw-cursor-default tw-select-none tw-items-center tw-rounded-sm tw-gap-2 tw-px-2 tw-py-1.5 tw-text-sm tw-outline-none tw-transition-colors focus:tw-bg-accent focus:tw-text-accent-foreground data-[disabled]:tw-pointer-events-none data-[disabled]:tw-opacity-50 [&>svg]:tw-size-4 [&>svg]:tw-shrink-0',
        inset && 'tw-pl-8',
        props.class,
      )
    "
  >
    <slot />
  </DropdownMenuItem>
</template>
