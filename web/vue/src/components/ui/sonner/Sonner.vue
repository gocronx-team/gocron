<script setup>
import { reactiveOmit } from '@vueuse/core'
import {
  CircleCheckIcon,
  InfoIcon,
  Loader2Icon,
  OctagonXIcon,
  TriangleAlertIcon,
  XIcon
} from 'lucide-vue-next'
import { Toaster as Sonner } from 'vue-sonner'

const props = defineProps({
  id: { type: String, required: false },
  invert: { type: Boolean, required: false },
  theme: { type: String, required: false },
  position: { type: String, required: false },
  closeButtonPosition: { type: String, required: false },
  hotkey: { type: Array, required: false },
  richColors: { type: Boolean, required: false },
  expand: { type: Boolean, required: false },
  duration: { type: Number, required: false },
  gap: { type: Number, required: false },
  visibleToasts: { type: Number, required: false },
  closeButton: { type: Boolean, required: false },
  toastOptions: { type: Object, required: false },
  class: { type: String, required: false },
  style: { type: null, required: false },
  offset: { type: [Object, String, Number], required: false },
  mobileOffset: { type: [Object, String, Number], required: false },
  dir: { type: String, required: false },
  swipeDirections: { type: Array, required: false },
  icons: { type: Object, required: false },
  containerAriaLabel: { type: String, required: false }
})
const delegatedProps = reactiveOmit(props, 'toastOptions')
</script>

<template>
  <Sonner
    class="tw-toaster tw-group"
    :toast-options="{
      classes: {
        toast:
          'group toast group-[.toaster]:tw-bg-background group-[.toaster]:tw-text-foreground group-[.toaster]:tw-border-border group-[.toaster]:tw-shadow-lg',
        description: 'group-[.toast]:tw-text-muted-foreground',
        actionButton:
          'group-[.toast]:tw-bg-primary group-[.toast]:tw-text-primary-foreground',
        cancelButton:
          'group-[.toast]:tw-bg-muted group-[.toast]:tw-text-muted-foreground'
      }
    }"
    v-bind="delegatedProps"
  >
    <template #success-icon>
      <CircleCheckIcon class="tw-size-4" />
    </template>
    <template #info-icon>
      <InfoIcon class="tw-size-4" />
    </template>
    <template #warning-icon>
      <TriangleAlertIcon class="tw-size-4" />
    </template>
    <template #error-icon>
      <OctagonXIcon class="tw-size-4" />
    </template>
    <template #loading-icon>
      <div>
        <Loader2Icon class="tw-size-4 tw-animate-spin" />
      </div>
    </template>
    <template #close-icon>
      <XIcon class="tw-size-4" />
    </template>
  </Sonner>
</template>
