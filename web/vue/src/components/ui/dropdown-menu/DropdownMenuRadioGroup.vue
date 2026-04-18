<script setup>
import { DropdownMenuRadioGroup, useForwardPropsEmits } from "reka-ui";

const props = defineProps({
  modelValue: { type: null, required: false },
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
});
const emits = defineEmits(["update:modelValue"]);

const forwarded = useForwardPropsEmits(props, emits);
</script>

<template>
  <DropdownMenuRadioGroup v-bind="forwarded">
    <slot />
  </DropdownMenuRadioGroup>
</template>
