<script setup>
import { Dot } from "lucide-vue-next";
import { Primitive, useForwardProps } from "reka-ui";

const props = defineProps({
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
});
const forwardedProps = useForwardProps(props);
</script>

<template>
  <Primitive v-bind="forwardedProps">
    <slot>
      <Dot />
    </slot>
  </Primitive>
</template>
