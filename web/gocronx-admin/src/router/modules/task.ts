import { AppRouteRecord } from '@/types/router'

/**
 * 任务管理：将 Dashboard、任务列表、任务日志、模板合并到同一个侧边栏父级下。
 * 这和 legacy web/vue 的导航结构一致（任务管理 → 子菜单）。
 *
 * 子项用绝对路径（以 / 开头），Vue Router 会按根路径解析 URL，但仍然挂在本
 * 父级的 Layout 组件下。这样既保留了原有的 /dashboard/console、/task/list、
 * /template/list 等 URL（书签不失效），又能在侧边栏里归到一个组里。
 */
export const taskRoutes: AppRouteRecord = {
  path: '/task',
  name: 'TaskMgmt',
  component: '/index/index',
  meta: {
    title: 'menus.task.title',
    icon: 'ri:calendar-check-line',
    roles: ['R_SUPER', 'R_ADMIN']
  },
  children: [
    {
      path: '/dashboard/console',
      name: 'Console',
      component: '/dashboard/console',
      meta: {
        title: 'menus.dashboard.console',
        icon: 'ri:pie-chart-line',
        keepAlive: false,
        fixedTab: true
      }
    },
    {
      path: '/task/list',
      name: 'TaskList',
      component: '/task/index',
      meta: {
        title: 'menus.task.list',
        icon: 'ri:list-check-2',
        keepAlive: true
      }
    },
    {
      path: '/task/log',
      name: 'TaskLog',
      component: '/task/log',
      meta: {
        title: 'menus.task.log',
        icon: 'ri:file-list-3-line',
        keepAlive: false
      }
    },
    {
      path: '/template/list',
      name: 'TemplateList',
      component: '/template/index',
      meta: {
        title: 'menus.template.list',
        icon: 'ri:file-copy-line',
        keepAlive: true
      }
    },
    // ── hidden child routes for create/edit flows ────────────────────────────
    {
      path: '/task/create',
      name: 'TaskCreate',
      component: '/task/edit',
      meta: {
        title: 'menus.task.create',
        isHide: true,
        keepAlive: false
      }
    },
    {
      path: '/task/edit/:id',
      name: 'TaskEdit',
      component: '/task/edit',
      meta: {
        title: 'menus.task.edit',
        isHide: true,
        keepAlive: false
      }
    },
    {
      path: '/template/create',
      name: 'TemplateCreate',
      component: '/template/edit',
      meta: {
        title: 'menus.template.create',
        isHide: true,
        keepAlive: false
      }
    },
    {
      path: '/template/edit/:id',
      name: 'TemplateEdit',
      component: '/template/edit',
      meta: {
        title: 'menus.template.edit',
        isHide: true,
        keepAlive: false
      }
    }
  ]
}
