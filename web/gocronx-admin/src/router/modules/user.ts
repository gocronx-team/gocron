import { AppRouteRecord } from '@/types/router'

/**
 * 用户管理独立侧边栏组。和 legacy web/vue 一致（legacy 顶栏 /user 是独立的一
 * 项，不在 /system 下）。
 *
 * 子路由用绝对路径保留原有 URL（/system/user/...、/system/user-center、
 * /system/two-factor），不让已有书签失效；只在视觉上把它们从系统组里分出来。
 */
export const userRoutes: AppRouteRecord = {
  path: '/user-mgmt',
  name: 'UserMgmt',
  component: '/index/index',
  meta: {
    title: 'menus.userMgmt.title',
    icon: 'ri:user-3-line',
    roles: ['R_SUPER', 'R_ADMIN']
  },
  children: [
    {
      path: '/system/user',
      name: 'User',
      component: '/system/user',
      meta: {
        title: 'menus.system.user',
        icon: 'ri:team-line',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN']
      }
    },
    // ── hidden children (reachable via nav or user-menu dropdown) ─────────────
    {
      path: '/system/user-center',
      name: 'UserCenter',
      component: '/system/user-center',
      meta: {
        title: 'menus.system.userCenter',
        isHide: true,
        keepAlive: true,
        isHideTab: true
      }
    },
    {
      path: '/system/two-factor',
      name: 'TwoFactor',
      component: '/system/user-center/two-factor',
      meta: {
        title: 'menus.system.twoFactor',
        isHide: true,
        keepAlive: false
      }
    },
    {
      path: '/system/user/edit/:id',
      name: 'UserEdit',
      component: '/system/user/edit',
      meta: {
        title: 'menus.system.userEdit',
        isHide: true,
        keepAlive: false
      }
    },
    {
      path: '/system/user/edit-password/:id',
      name: 'UserEditPassword',
      component: '/system/user/edit-password',
      meta: {
        title: 'menus.system.userEditPassword',
        isHide: true,
        keepAlive: false
      }
    }
  ]
}
