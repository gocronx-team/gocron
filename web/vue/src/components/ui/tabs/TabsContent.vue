<script setup>
import { reactiveOmit } from "@vueuse/core";
import { TabsContent } from "reka-ui";
import { cn } from "@/lib/utils";

const props = defineProps({
  value: { type: [String, Number], required: true },
  forceMount: { type: Boolean, required: false },
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
  class: { type: null, required: false },
});

const delegatedProps = reactiveOmit(props, "class");
</script>

<template>
  <TabsContent
    :class="
      cn(
        'tw-mt-2 tw-ring-offset-background focus-visible:tw-outline-none focus-visible:tw-ring-2 focus-visible:tw-ring-ring focus-visible:tw-ring-offset-2',
        props.class,
      )
    "
    v-bind="delegatedProps"
  >
    <slot />
  </TabsContent>
</template>
