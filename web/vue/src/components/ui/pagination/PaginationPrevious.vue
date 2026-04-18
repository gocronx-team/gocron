<script setup>
import { reactiveOmit } from "@vueuse/core";
import { ChevronLeftIcon } from "lucide-vue-next";
import { PaginationPrev, useForwardProps } from "reka-ui";
import { cn } from "@/lib/utils";
import { buttonVariants } from '@/components/ui/button';

const props = defineProps({
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
  size: { type: null, required: false, default: "default" },
  class: { type: null, required: false },
});

const delegatedProps = reactiveOmit(props, "class", "size");
const forwarded = useForwardProps(delegatedProps);
</script>

<template>
  <PaginationPrev
    data-slot="pagination-previous"
    :class="
      cn(
        buttonVariants({ variant: 'ghost', size }),
        'tw-gap-1 tw-px-2.5 sm:tw-pr-2.5',
        props.class,
      )
    "
    v-bind="forwarded"
  >
    <slot>
      <ChevronLeftIcon />
      <span class="tw-hidden sm:tw-block">Previous</span>
    </slot>
  </PaginationPrev>
</template>
