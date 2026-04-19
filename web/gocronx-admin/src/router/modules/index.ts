import { AppRouteRecord } from '@/types/router'
import { dashboardRoutes } from './dashboard'
import { systemRoutes } from './system'
import { hostRoutes } from './host'
import { taskRoutes } from './task'
import { templateRoutes } from './template'

/**
 * 导出所有模块化路由
 */
export const routeModules: AppRouteRecord[] = [
  dashboardRoutes,
  hostRoutes,
  taskRoutes,
  templateRoutes,
  systemRoutes
]
