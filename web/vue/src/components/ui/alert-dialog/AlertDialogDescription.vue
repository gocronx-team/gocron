<script setup>
import { reactiveOmit } from "@vueuse/core";
import { AlertDialogDescription } from "reka-ui";
import { cn } from "@/lib/utils";

const props = defineProps({
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
  class: { type: null, required: false },
});

const delegatedProps = reactiveOmit(props, "class");
</script>

<template>
  <AlertDialogDescription
    v-bind="delegatedProps"
    :class="cn('tw-text-sm tw-text-muted-foreground', props.class)"
  >
    <slot />
  </AlertDialogDescription>
</template>
