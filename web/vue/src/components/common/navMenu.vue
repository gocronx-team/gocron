<template>
  <div v-cloak>
    <el-menu
      :default-active="currentRoute"
      mode="horizontal"
      background-color="#545c64"
      text-color="#fff"
      active-text-color="#ffd04b"
      router>
      <el-menu-item index="/task">任务管理</el-menu-item>
      <el-menu-item index="/host">任务节点</el-menu-item>
      <el-menu-item v-if="userStore.isAdmin" index="/user">用户管理</el-menu-item>
      <el-menu-item v-if="userStore.isAdmin" index="/system">系统管理</el-menu-item>
      <div style="flex: 1;"></div>
      <el-dropdown v-if="userStore.token" trigger="click" style="margin-left: auto;">
        <span style="color: #fff; cursor: pointer; padding: 0 20px; line-height: 60px;">
          {{ userStore.username || '用户' }}
          <el-icon style="margin-left: 5px;"><arrow-down /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="$router.push('/user/edit-my-password')">修改密码</el-dropdown-item>
            <el-dropdown-item @click="$router.push('/user/two-factor')">双因素认证</el-dropdown-item>
            <el-dropdown-item divided @click="logout">退出</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </el-menu>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../../stores/user'
import { ArrowDown } from '@element-plus/icons-vue'

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
.el-menu {
  display: flex;
  align-items: center;
}
</style>