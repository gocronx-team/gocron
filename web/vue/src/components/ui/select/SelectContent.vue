<script setup>
import { reactiveOmit } from "@vueuse/core";
import {
  SelectContent,
  SelectPortal,
  SelectViewport,
  useForwardPropsEmits,
} from "reka-ui";
import { cn } from "@/lib/utils";
import { SelectScrollDownButton, SelectScrollUpButton } from ".";

defineOptions({
  inheritAttrs: false,
});

const props = defineProps({
  forceMount: { type: Boolean, required: false },
  position: { type: String, required: false, default: "popper" },
  bodyLock: { type: Boolean, required: false },
  side: { type: null, required: false },
  sideOffset: { type: Number, required: false },
  sideFlip: { type: Boolean, required: false },
  align: { type: null, required: false },
  alignOffset: { type: Number, required: false },
  alignFlip: { type: Boolean, required: false },
  avoidCollisions: { type: Boolean, required: false },
  collisionBoundary: { type: null, required: false },
  collisionPadding: { type: [Number, Object], required: false },
  arrowPadding: { type: Number, required: false },
  hideShiftedArrow: { type: Boolean, required: false },
  sticky: { type: String, required: false },
  hideWhenDetached: { type: Boolean, required: false },
  positionStrategy: { type: String, required: false },
  updatePositionStrategy: { type: String, required: false },
  disableUpdateOnLayoutShift: { type: Boolean, required: false },
  prioritizePosition: { type: Boolean, required: false },
  reference: { type: null, required: false },
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
  disableOutsidePointerEvents: { type: Boolean, required: false },
  class: { type: null, required: false },
});
const emits = defineEmits([
  "closeAutoFocus",
  "escapeKeyDown",
  "pointerDownOutside",
]);

const delegatedProps = reactiveOmit(props, "class");

const forwarded = useForwardPropsEmits(delegatedProps, emits);
</script>

<template>
  <SelectPortal>
    <SelectContent
      v-bind="{ ...forwarded, ...$attrs }"
      :class="
        cn(
          'tw-relative tw-z-50 tw-max-h-96 tw-min-w-32 tw-overflow-hidden tw-rounded-md tw-border tw-bg-popover tw-text-popover-foreground tw-shadow-md data-[state=open]:tw-animate-in data-[state=closed]:tw-animate-out data-[state=closed]:tw-fade-out-0 data-[state=open]:tw-fade-in-0 data-[state=closed]:tw-zoom-out-95 data-[state=open]:tw-zoom-in-95 data-[side=bottom]:tw-slide-in-from-top-2 data-[side=left]:tw-slide-in-from-right-2 data-[side=right]:tw-slide-in-from-left-2 data-[side=top]:tw-slide-in-from-bottom-2',
          position === 'tw-popper' &&
            'data-[side=bottom]:tw-translate-y-1 data-[side=left]:tw--translate-x-1 data-[side=right]:tw-translate-x-1 data-[side=top]:tw--translate-y-1',
          props.class,
        )
      "
    >
      <SelectScrollUpButton />
      <SelectViewport
        :class="
          cn(
            'tw-p-1',
            position === 'tw-popper' &&
              'tw-h-[--reka-select-trigger-height] tw-w-full tw-min-w-[--reka-select-trigger-width]',
          )
        "
      >
        <slot />
      </SelectViewport>
      <SelectScrollDownButton />
    </SelectContent>
  </SelectPortal>
</template>
