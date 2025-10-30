<template>
  <div ref="footerContent" :class="{ 'empty-footer': isEmpty }">
    <slot></slot>
  </div>
</template>

<script>
import { ref, computed, defineExpose } from 'vue'

export default {
  name: 'app-footer',
  setup() {
    const footerContent = ref(null)
    
    const isEmpty = computed(() => {
      if (!footerContent.value) return true
      const content = footerContent.value.textContent?.trim() || ''
      const hasSlotContent = footerContent.value.children.length > 0
      return !content && !hasSlotContent
    })
    
    defineExpose({ isEmpty })
    
    return {
      footerContent,
      isEmpty
    }
  }
}
</script>

<style scoped>
.empty-footer {
  height: 0;
  overflow: hidden;
}
</style>
