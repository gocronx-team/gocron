<template>
  <el-aside width="150px">
    <el-menu
      :default-active="currentRoute"
      mode="vertical"
      background-color="#545c64"
      text-color="#fff"
      active-text-color="#ffd04b"
      router>
      <el-menu-item index="/system">{{ t('system.notification') }}</el-menu-item>
      <el-menu-item index="/system/login-log">{{ t('system.loginLog') }}</el-menu-item>
      <el-menu-item index="/system/log-retention">{{ t('system.logCleanup') }}</el-menu-item>
    </el-menu>
  </el-aside>
</template>

<script>
import { useI18n } from 'vue-i18n'
export default {
  name: 'system-sidebar',
  setup() {
    const { t } = useI18n()
    return { t }
  },
  data () {
    return {}
  },
  computed: {
    currentRoute () {
      if (this.$route.path === '/system/login-log') {
        return '/system/login-log'
      }
      if (this.$route.path === '/system/log-retention') {
        return '/system/log-retention'
      }
      return '/system'
    }
  }
}
</script>
