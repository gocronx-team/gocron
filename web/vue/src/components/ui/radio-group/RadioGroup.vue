<script setup>
import { reactiveOmit } from "@vueuse/core";
import { RadioGroupRoot, useForwardPropsEmits } from "reka-ui";
import { cn } from "@/lib/utils";

const props = defineProps({
  modelValue: { type: null, required: false },
  defaultValue: { type: null, required: false },
  disabled: { type: Boolean, required: false },
  orientation: { type: String, required: false },
  dir: { type: String, required: false },
  loop: { type: Boolean, required: false },
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
  name: { type: String, required: false },
  required: { type: Boolean, required: false },
  class: { type: null, required: false },
});
const emits = defineEmits(["update:modelValue"]);

const delegatedProps = reactiveOmit(props, "class");

const forwarded = useForwardPropsEmits(delegatedProps, emits);
</script>

<template>
  <RadioGroupRoot :class="cn('tw-grid tw-gap-2', props.class)" v-bind="forwarded">
    <slot />
  </RadioGroupRoot>
</template>
