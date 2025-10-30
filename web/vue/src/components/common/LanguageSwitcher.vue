<template>
  <el-dropdown @command="handleCommand">
    <span class="language-switcher">
      🌐 {{ currentLanguage }}
    </span>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item command="zh-CN" :disabled="locale === 'zh-CN'">
          简体中文
        </el-dropdown-item>
        <el-dropdown-item command="en-US" :disabled="locale === 'en-US'">
          English
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { locale } = useI18n()

const currentLanguage = computed(() => {
  return locale.value === 'zh-CN' ? '简体中文' : 'English'
})

const handleCommand = (command) => {
  locale.value = command
  localStorage.setItem('locale', command)
}
</script>

<style scoped>
.language-switcher {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 0 12px;
  font-size: 14px;
  white-space: nowrap;
}

.language-switcher:hover {
  color: #409eff;
}
</style>
