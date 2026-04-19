import { AppRouteRecord } from '@/types/router'

export const taskRoutes: AppRouteRecord = {
  path: '/task',
  name: 'Task',
  component: '/index/index',
  meta: {
    title: 'menus.task.title',
    icon: 'ri:calendar-check-line',
    roles: ['R_SUPER', 'R_ADMIN']
  },
  children: [
    {
      path: 'list',
      name: 'TaskList',
      component: '/task/index',
      meta: {
        title: 'menus.task.list',
        keepAlive: true
      }
    },
    {
      path: 'create',
      name: 'TaskCreate',
      component: '/task/edit',
      meta: {
        title: 'menus.task.create',
        isHide: true,
        keepAlive: false
      }
    },
    {
      path: 'edit/:id',
      name: 'TaskEdit',
      component: '/task/edit',
      meta: {
        title: 'menus.task.edit',
        isHide: true,
        keepAlive: false
      }
    },
    {
      path: 'log',
      name: 'TaskLog',
      component: '/task/log',
      meta: {
        title: 'menus.task.log',
        isHide: true,
        keepAlive: false
      }
    }
  ]
}
