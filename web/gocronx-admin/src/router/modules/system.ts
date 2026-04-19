import { AppRouteRecord } from '@/types/router'

export const systemRoutes: AppRouteRecord = {
  path: '/system',
  name: 'System',
  component: '/index/index',
  meta: {
    title: 'menus.system.title',
    icon: 'ri:user-3-line',
    roles: ['R_SUPER', 'R_ADMIN']
  },
  children: [
    {
      path: 'user',
      name: 'User',
      component: '/system/user',
      meta: {
        title: 'menus.system.user',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN']
      }
    },
    {
      path: 'user-center',
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
      path: 'login-log',
      name: 'LoginLog',
      component: '/system/login-log',
      meta: {
        title: 'menus.system.loginLog',
        icon: 'ri:login-box-line',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN']
      }
    },
    {
      path: 'audit-log',
      name: 'AuditLog',
      component: '/system/audit-log',
      meta: {
        title: 'menus.system.auditLog',
        icon: 'ri:file-shield-line',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN']
      }
    },
    {
      path: 'two-factor',
      name: 'TwoFactor',
      component: '/system/user-center/two-factor',
      meta: {
        title: 'menus.system.twoFactor',
        isHide: true,
        keepAlive: false
      }
    },
    {
      path: 'user/edit/:id',
      name: 'UserEdit',
      component: '/system/user/edit',
      meta: {
        title: 'menus.system.userEdit',
        isHide: true,
        keepAlive: false
      }
    },
    {
      path: 'user/edit-password/:id',
      name: 'UserEditPassword',
      component: '/system/user/edit-password',
      meta: {
        title: 'menus.system.userEditPassword',
        isHide: true,
        keepAlive: false
      }
    },
    {
      path: 'log-retention',
      name: 'LogRetention',
      component: '/system/log-retention/index',
      meta: {
        title: 'menus.system.logRetention',
        icon: 'ri:time-line',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN']
      }
    },
    {
      path: 'notification',
      name: 'Notification',
      component: '/system/notification/index',
      meta: {
        title: 'menus.system.notification',
        icon: 'ri:notification-2-line',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN']
      }
    }
  ]
}
