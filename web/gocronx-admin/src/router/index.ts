import type { App } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import { staticRoutes } from './routes/staticRoutes'
import { configureNProgress } from '@/utils/router'
import { setupBeforeEachGuard } from './guards/beforeEach'
import { setupAfterEachGuard } from './guards/afterEach'

// 创建路由实例
export const router = createRouter({
  history: createWebHashHistory(),
  routes: staticRoutes // 静态路由
})

// 初始化路由
export function initRouter(app: App<Element>): void {
  configureNProgress() // 顶部进度条
  setupBeforeEachGuard(router) // 路由前置守卫
  setupAfterEachGuard(router) // 路由后置守卫
  app.use(router)
}

// 主页路径 —— 登录后默认跳转到此。任务列表是 gocron 用户最高频的页面，
// 选它当首页省去一次额外点击。留空则 fall back 到 getFirstMenuPath（菜单第一项）。
export const HOME_PAGE_PATH = '/task/list'
