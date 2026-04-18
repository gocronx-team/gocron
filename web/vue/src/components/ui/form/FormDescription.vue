<script setup>
import { cn } from "@/lib/utils";
import { useFormField } from "./useFormField";

const props = defineProps({
  class: { type: null, required: false },
});

const { formDescriptionId } = useFormField();
</script>

<template>
  <p
    :id="formDescriptionId"
    :class="cn('tw-text-sm tw-text-muted-foreground', props.class)"
  >
    <slot />
  </p>
</template>
