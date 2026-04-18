<script setup>
import { reactiveOmit } from "@vueuse/core";
import { Primitive, useForwardProps } from "reka-ui";
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
  <Primitive
    v-bind="forwardedProps"
    :class="cn('tw-flex tw-items-center', props.class)"
  >
    <slot />
  </Primitive>
</template>
