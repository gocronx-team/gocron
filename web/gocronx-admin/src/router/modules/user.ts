import { AppRouteRecord } from '@/types/router'

/**
 * 用户列表独立在侧边栏一级条目，不作可展开父级（下面没有子菜单）。
 *
 * 没有 children → RouteTransformer 会把它当 first-level route，自动包一层
 * Layout，URL 保持 /system/user。个人中心、两步验证、编辑用户、重置密码这
 * 些页面的路由挂在 systemRoutes 下（以 isHide 的子路由形式），共用同一个
 * Layout。
 */
export const userRoutes: AppRouteRecord = {
  path: '/system/user',
  name: 'User',
  component: '/system/user',
  meta: {
    title: 'menus.system.user',
    icon: 'ri:user-3-line',
    keepAlive: true,
    roles: ['R_SUPER', 'R_ADMIN']
  }
}
