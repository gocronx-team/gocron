import { AppRouteRecord } from '@/types/router'

export const templateRoutes: AppRouteRecord = {
  path: '/template',
  name: 'Template',
  component: '/index/index',
  meta: {
    title: 'menus.template.title',
    icon: 'ri:file-copy-line',
    roles: ['R_SUPER', 'R_ADMIN']
  },
  children: [
    {
      path: 'list',
      name: 'TemplateList',
      component: '/template/index',
      meta: {
        title: 'menus.template.list',
        keepAlive: true
      }
    },
    {
      path: 'create',
      name: 'TemplateCreate',
      component: '/template/edit',
      meta: {
        title: 'menus.template.create',
        isHide: true,
        keepAlive: false
      }
    },
    {
      path: 'edit/:id',
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
