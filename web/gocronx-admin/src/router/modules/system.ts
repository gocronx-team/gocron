import { AppRouteRecord } from '@/types/router'

export const systemRoutes: AppRouteRecord = {
  path: '/system',
  name: 'System',
  component: '/index/index',
  meta: {
    title: 'menus.system.title',
    icon: 'ri:settings-3-line',
    roles: ['R_SUPER', 'R_ADMIN']
  },
  children: [
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
    // ── hidden user-related routes ────────────────────────────────────────────
    // User list is a standalone sidebar entry (router/modules/user.ts), but its
    // support pages — user center, 2FA, edit user, reset password — live here
    // as hidden children of the System layout. URLs are unchanged; they just
    // don't render in the sidebar.
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
    }
  ]
}
