<template>
  <el-aside width="200px" class="global-sidebar">
    <div class="sidebar-header">
      <h2 class="app-title">gocron</h2>
    </div>
    
    <el-menu
      :default-active="currentRoute"
      mode="vertical"
      background-color="#304156"
      text-color="#bfcbd9"
      active-text-color="#409EFF"
      :unique-opened="true"
      router>
      
      <!-- 任务管理 -->
      <el-sub-menu index="task">
        <template #title>
          <el-icon><Calendar /></el-icon>
          <span>{{ t('nav.taskManage') }}</span>
        </template>
        <el-menu-item index="/task">
          <el-icon><List /></el-icon>
          <span>{{ t('task.list') }}</span>
        </el-menu-item>
        <el-menu-item index="/task/log">
          <el-icon><Document /></el-icon>
          <span>{{ t('task.log') }}</span>
        </el-menu-item>
        <el-menu-item index="/template">
          <el-icon><Files /></el-icon>
          <span>{{ t('template.list') }}</span>
        </el-menu-item>
        <el-menu-item index="/statistics">
          <el-icon><TrendCharts /></el-icon>
          <span>{{ t('nav.statistics') }}</span>
        </el-menu-item>
      </el-sub-menu>
      
      <!-- 任务节点 -->
      <el-menu-item index="/host">
        <el-icon><Monitor /></el-icon>
        <span>{{ t('nav.taskNode') }}</span>
      </el-menu-item>
      
      <!-- 用户管理 -->
      <el-menu-item v-if="userStore.isAdmin" index="/user">
        <el-icon><User /></el-icon>
        <span>{{ t('nav.userManage') }}</span>
      </el-menu-item>
      
      <!-- 系统管理 -->
      <el-sub-menu v-if="userStore.isAdmin" index="system">
        <template #title>
          <el-icon><Setting /></el-icon>
          <span>{{ t('nav.systemManage') }}</span>
        </template>
        <el-menu-item index="/system">
          <el-icon><Bell /></el-icon>
          <span>{{ t('system.notification') }}</span>
        </el-menu-item>
        <el-menu-item index="/system/login-log">
          <el-icon><Document /></el-icon>
          <span>{{ t('system.loginLog') }}</span>
        </el-menu-item>
        <el-menu-item index="/system/audit-log">
          <el-icon><Notebook /></el-icon>
          <span>{{ t('audit.log') }}</span>
        </el-menu-item>
        <el-menu-item index="/system/log-retention">
          <el-icon><Delete /></el-icon>
          <span>{{ t('system.logCleanup') }}</span>
        </el-menu-item>
      </el-sub-menu>
    </el-menu>
    
    <!-- 底部语言切换 -->
    <div class="sidebar-footer">
      <LanguageSwitcher />
    </div>
  </el-aside>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '../../stores/user'
import LanguageSwitcher from './LanguageSwitcher.vue'
import {
  Calendar,
  List,
  Document,
  Files,
  TrendCharts,
  Monitor,
  User,
  Setting,
  Bell,
  Delete,
  Notebook
} from '@element-plus/icons-vue'

const { t } = useI18n()
const route = useRoute()
const userStore = useUserStore()

const currentRoute = computed(() => {
  const path = route.path
  // 精确匹配路由
  if (path === '/task/log') return '/task/log'
  if (path === '/statistics') return '/statistics'
  if (path.startsWith('/template')) return '/template'
  if (path.startsWith('/task')) return '/task'
  if (path.startsWith('/host')) return '/host'
  if (path.startsWith('/user')) return '/user'
  if (path.startsWith('/system')) {
    if (path === '/system/login-log') return '/system/login-log'
    if (path === '/system/audit-log') return '/system/audit-log'
    if (path === '/system/log-retention') return '/system/log-retention'
    return '/system'
  }
  return '/task'
})
</script>

<style scoped>
.global-sidebar {
  background-color: #304156;
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.sidebar-header {
  padding: 20px 20px 20px 32px;
  text-align: left;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.app-title {
  margin: 0;
  color: #409EFF;
  font-size: 24px;
  font-weight: bold;
  letter-spacing: 1px;
}

.el-menu {
  flex: 1;
  border: none;
  overflow-y: auto;
}

.el-menu::-webkit-scrollbar {
  width: 6px;
}

.el-menu::-webkit-scrollbar-thumb {
  background-color: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}

.sidebar-footer {
  padding: 15px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  background-color: #263445;
}

.sidebar-footer :deep(.language-switcher) {
  color: #bfcbd9;
  font-weight: 500;
  justify-content: center;
  padding: 8px 12px;
  border-radius: 4px;
  transition: all 0.3s;
}

.sidebar-footer :deep(.language-switcher:hover) {
  background-color: rgba(64, 158, 255, 0.1);
  color: #409EFF;
}

/* 子菜单样式 */
:deep(.el-sub-menu__title) {
  color: #bfcbd9 !important;
  display: flex;
  align-items: center;
  padding-left: 20px !important;
  padding-right: 40px !important;
}

:deep(.el-sub-menu__title:hover) {
  background-color: rgba(0, 0, 0, 0.2) !important;
}

:deep(.el-menu-item) {
  display: flex;
  align-items: center;
  padding-left: 20px !important;
  padding-right: 20px !important;
}

:deep(.el-menu-item:hover) {
  background-color: rgba(0, 0, 0, 0.2) !important;
}

:deep(.el-menu-item.is-active) {
  background-color: rgba(64, 158, 255, 0.2) !important;
}

/* 确保图标和文字垂直居中 */
:deep(.el-icon) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  vertical-align: middle;
  margin-right: 8px;
  flex-shrink: 0;
}

:deep(.el-sub-menu__title span),
:deep(.el-menu-item span) {
  vertical-align: middle;
  line-height: normal;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 子菜单的子项增加缩进 */
:deep(.el-menu--inline .el-menu-item) {
  padding-left: 48px !important;
}

/* 确保展开箭头不被遮挡 */
:deep(.el-sub-menu__icon-arrow) {
  margin-left: auto !important;
  flex-shrink: 0;
}
</style>
