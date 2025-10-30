<template>
  <el-aside width="150px" class="sidebar-container">
    <el-menu
      :default-active="currentRoute"
      mode="vertical"
      background-color="#545c64"
      text-color="#fff"
      active-text-color="#ffd04b"
      router>
      <el-menu-item index="/task">{{ t('task.list') }}</el-menu-item>
      <el-menu-item index="/task/log">{{ t('task.log') }}</el-menu-item>
    </el-menu>
    <div class="sidebar-language-switcher">
      <LanguageSwitcher />
    </div>
  </el-aside>
</template>

<script>
import { useI18n } from 'vue-i18n'
import LanguageSwitcher from '../../components/common/LanguageSwitcher.vue'

export default {
  name: 'task-sidebar',
  components: { LanguageSwitcher },
  setup() {
    const { t } = useI18n()
    return { t }
  },
  data () {
    return {}
  },
  computed: {
    currentRoute () {
      if (this.$route.path === '/task/log') {
        return '/task/log'
      }
      return '/task'
    }
  }
}
</script>

<style scoped>
.sidebar-container {
  position: relative;
  display: flex;
  flex-direction: column;
}

.sidebar-language-switcher {
  position: absolute;
  bottom: 20px;
  left: 0;
  right: 0;
  padding: 10px;
  background-color: #545c64;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.sidebar-language-switcher :deep(.language-switcher) {
  color: #67c23a;
  font-weight: 500;
  justify-content: center;
  padding: 8px 12px;
  border-radius: 4px;
  transition: all 0.3s;
}

.sidebar-language-switcher :deep(.language-switcher:hover) {
  background-color: rgba(103, 194, 58, 0.1);
  color: #85ce61;
}
</style>
