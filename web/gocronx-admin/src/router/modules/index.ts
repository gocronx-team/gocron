import { AppRouteRecord } from '@/types/router'
import { systemRoutes } from './system'
import { hostRoutes } from './host'
import { taskRoutes } from './task'

/**
 * 导出所有模块化路由。
 *
 * taskRoutes 是合并后的「任务管理」父级，里面包含 Dashboard / 任务列表 /
 * 任务日志 / 模板 四个子菜单。以前分散在 dashboardRoutes / templateRoutes
 * 的内容都挪进去了，URL 不变。
 */
// taskRoutes first so the default home resolves to /dashboard/console
// (getFirstMenuPath walks top-to-bottom, child-first).
export const routeModules: AppRouteRecord[] = [taskRoutes, hostRoutes, systemRoutes]
