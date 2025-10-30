<template>
  <div v-cloak class="nav-container">
    <el-menu
      :default-active="currentRoute"
      mode="horizontal"
      background-color="#545c64"
      text-color="#fff"
      active-text-color="#ffd04b"
      router>
      <el-menu-item index="/task">{{ t('nav.taskManage') }}</el-menu-item>
      <el-menu-item index="/host">{{ t('nav.taskNode') }}</el-menu-item>
      <el-menu-item v-if="userStore.isAdmin" index="/user">{{ t('nav.userManage') }}</el-menu-item>
      <el-menu-item v-if="userStore.isAdmin" index="/system">{{ t('nav.systemManage') }}</el-menu-item>
    </el-menu>
    <div v-if="userStore.isLogin" class="user-menu">
      <el-dropdown trigger="click">
        <span class="user-info">
          <el-icon><User /></el-icon>
          <span>{{ userStore.username }}</span>
          <el-icon><ArrowDown /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="$router.push('/user/edit-my-password')">{{ t('nav.changePassword') }}</el-dropdown-item>
            <el-dropdown-item @click="$router.push('/user/two-factor')">{{ t('nav.twoFactor') }}</el-dropdown-item>
            <el-dropdown-item divided @click="logout">{{ t('nav.logout') }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '../../stores/user'
import { ArrowDown, User } from '@element-plus/icons-vue'

const { t } = useI18n()

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const currentRoute = computed(() => {
  if (route.path === '/') return '/task'
  const segments = route.path.split('/')
  return `/${segments[1]}`
})

const logout = () => {
  userStore.logout()
  router.push('/user/login').then(() => {
    window.location.reload()
  })
}
</script>

<style scoped>
.nav-container {
  display: flex;
  align-items: center;
  background-color: #545c64;
}
.el-menu {
  flex: 1;
  border: none;
}
.user-menu {
  padding: 0 20px;
}
.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #fff;
  cursor: pointer;
  padding: 10px;
  border-radius: 4px;
  transition: background-color 0.3s;
}
.user-info:hover {
  background-color: rgba(255, 255, 255, 0.1);
}
</style>