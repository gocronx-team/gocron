<script setup>
import { reactiveOmit } from "@vueuse/core";
import { cn } from "@/lib/utils";
import TableCell from "./TableCell.vue";
import TableRow from "./TableRow.vue";

const props = defineProps({
  class: { type: null, required: false },
  colspan: { type: Number, required: false, default: 1 },
});

const delegatedProps = reactiveOmit(props, "class");
</script>

<template>
  <TableRow>
    <TableCell
      :class="
        cn(
          'tw-p-4 tw-whitespace-nowrap tw-align-middle tw-text-sm tw-text-foreground',
          props.class,
        )
      "
      v-bind="delegatedProps"
    >
      <div class="tw-flex tw-items-center tw-justify-center tw-py-10">
        <slot />
      </div>
    </TableCell>
  </TableRow>
</template>
