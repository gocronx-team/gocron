import { createRouter, createWebHashHistory } from 'vue-router'
import { useUserStore } from '../stores/user'

const routes = [
  {
    path: '/',
    redirect: '/task'
  },
  {
    path: '/install',
    name: 'install',
    component: () => import('../pages/install/index.vue'),
    meta: { noLogin: true, noNeedAdmin: true }
  },
  {
    path: '/task',
    name: 'task-list',
    component: () => import('../pages/task/list.vue'),
    meta: { noNeedAdmin: true }
  },
  {
    path: '/task/create',
    name: 'task-create',
    component: () => import('../pages/task/edit.vue')
  },
  {
    path: '/task/edit/:id',
    name: 'task-edit',
    component: () => import('../pages/task/edit.vue')
  },
  {
    path: '/task/log',
    name: 'task-log',
    component: () => import('../pages/taskLog/list.vue'),
    meta: { noNeedAdmin: true }
  },
  {
    path: '/template',
    name: 'template-list',
    component: () => import('../pages/template/list.vue'),
    meta: { noNeedAdmin: true }
  },
  {
    path: '/template/create',
    name: 'template-create',
    component: () => import('../pages/template/edit.vue')
  },
  {
    path: '/template/edit/:id',
    name: 'template-edit',
    component: () => import('../pages/template/edit.vue')
  },
  {
    path: '/host',
    name: 'host-list',
    component: () => import('../pages/host/list.vue'),
    meta: { noNeedAdmin: true }
  },
  {
    path: '/host/create',
    name: 'host-create',
    component: () => import('../pages/host/edit.vue')
  },
  {
    path: '/host/edit/:id',
    name: 'host-edit',
    component: () => import('../pages/host/edit.vue')
  },
  {
    path: '/user',
    name: 'user-list',
    component: () => import('../pages/user/list.vue')
  },
  {
    path: '/user/create',
    name: 'user-create',
    component: () => import('../pages/user/edit.vue')
  },
  {
    path: '/user/edit/:id',
    name: 'user-edit',
    component: () => import('../pages/user/edit.vue')
  },
  {
    path: '/user/login',
    name: 'user-login',
    component: () => import('../pages/user/login.vue'),
    meta: { noLogin: true }
  },
  {
    path: '/user/edit-password/:id',
    name: 'user-edit-password',
    component: () => import('../pages/user/editPassword.vue')
  },
  {
    path: '/user/edit-my-password',
    name: 'user-edit-my-password',
    component: () => import('../pages/user/editMyPassword.vue'),
    meta: { noNeedAdmin: true }
  },
  {
    path: '/user/two-factor',
    name: 'user-two-factor',
    component: () => import('../pages/user/twoFactor.vue'),
    meta: { noNeedAdmin: true }
  },
  {
    path: '/system',
    redirect: '/system/notification/email'
  },
  {
    path: '/system/notification/email',
    name: 'system-notification-email',
    component: () => import('../pages/system/notification/email.vue')
  },
  {
    path: '/system/notification/slack',
    name: 'system-notification-slack',
    component: () => import('../pages/system/notification/slack.vue')
  },
  {
    path: '/system/notification/webhook',
    name: 'system-notification-webhook',
    component: () => import('../pages/system/notification/webhook.vue')
  },
  {
    path: '/system/login-log',
    name: 'login-log',
    component: () => import('../pages/system/loginLog.vue')
  },
  {
    path: '/system/log-retention',
    name: 'log-retention',
    component: () => import('../pages/system/logRetention.vue')
  },
  {
    path: '/system/audit-log',
    name: 'system-audit-log',
    component: () => import('../pages/system/auditLog.vue')
  },
  {
    path: '/statistics',
    name: 'statistics',
    component: () => import('../pages/statistics/index.vue'),
    meta: { noNeedAdmin: true }
  },
  {
    path: '/:pathMatch(.*)*',
    component: () => import('../components/common/notFound.vue'),
    meta: { noLogin: true, noNeedAdmin: true }
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  // 防止登录后访问登录页
  if (to.path === '/user/login') {
    const userStore = useUserStore()
    if (userStore.token) {
      next({ path: '/' })
      return
    }
  }

  if (to.meta.noLogin) {
    next()
    return
  }

  const userStore = useUserStore()

  if (userStore.token) {
    if (userStore.isAdmin || to.meta.noNeedAdmin) {
      next()
      return
    }
    next({ path: '/404.html' })
    return
  }

  next({
    path: '/user/login',
    query: { redirect: to.fullPath }
  })
})

export default router
