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
    }
  ]
}
