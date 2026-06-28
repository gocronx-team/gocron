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
    {
      path: 'mcp-token',
      name: 'McpToken',
      component: '/system/mcp-token/index',
      meta: {
        title: 'menus.system.mcpToken',
        icon: 'ri:key-2-line',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN']
      }
    },
    {
      path: 'secret',
      name: 'Secret',
      component: '/system/secret/index',
      meta: {
        title: 'menus.system.secret',
        icon: 'ri:lock-password-line',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN']
      }
    },
    {
      path: 'ai-config',
      name: 'AiConfig',
      component: '/system/ai-config/index',
      meta: {
        title: 'menus.system.aiConfig',
        icon: 'ri:robot-2-line',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN']
      }
    },
    // ── hidden admin user-management routes ───────────────────────────────────
    // User list is a standalone sidebar entry (router/modules/user.ts), but its
    // admin support pages — edit user, reset password — live here as hidden
    // children of the (admin-only) System layout. URLs are unchanged; they just
    // don't render in the sidebar.
    //
    // Self-service pages (user center, 2FA) are NOT here — they live in
    // router/modules/account.ts so normal users can reach them too.
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
