import { AppRouteRecord } from '@/types/router'
import { systemRoutes } from './system'
import { hostRoutes } from './host'
import { taskRoutes } from './task'
import { userRoutes } from './user'

/**
 * 导出所有模块化路由。
 *
 * 顶级组（对应 legacy web/vue 顶部导航）：
 * 1. 任务管理 — Dashboard / 任务列表 / 任务日志 / 模板
 * 2. 节点管理 — Host List
 * 3. 用户管理 — User List（个人中心 / 2FA / edit 等为 hidden 子路由）
 * 4. 系统 — Notification / Login Log / Audit Log / Log Retention
 *
 * taskRoutes 第一，这样 getFirstMenuPath 会把默认首页解析成 /dashboard/console。
 */
export const routeModules: AppRouteRecord[] = [taskRoutes, hostRoutes, userRoutes, systemRoutes]
