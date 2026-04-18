<script setup>
import { reactiveOmit } from "@vueuse/core";
import { SelectGroup } from "reka-ui";
import { cn } from "@/lib/utils";

const props = defineProps({
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
  class: { type: null, required: false },
});

const delegatedProps = reactiveOmit(props, "class");
</script>

<template>
  <SelectGroup :class="cn('tw-p-1 tw-w-full', props.class)" v-bind="delegatedProps">
    <slot />
  </SelectGroup>
</template>
