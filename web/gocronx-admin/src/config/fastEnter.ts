/**
 * 快速入口配置
 * 包含：应用列表、快速链接等配置
 */
import { WEB_LINKS } from '@/utils/constants'
import type { FastEnterConfig } from '@/types/config'

const fastEnterConfig: FastEnterConfig = {
  // 显示条件（屏幕宽度）
  minWidth: 1200,
  // 应用列表
  applications: [
    {
      name: '工作台',
      description: '系统概览与数据统计',
      icon: 'ri:pie-chart-line',
      iconColor: '#377dff',
      enabled: true,
      order: 1,
      routeName: 'Console'
    },
    {
      name: '官方文档',
      description: '使用指南与开发文档',
      icon: 'ri:bill-line',
      iconColor: '#ffb100',
      enabled: true,
      order: 2,
      link: WEB_LINKS.DOCS
    },
    {
      name: '技术支持',
      description: '技术支持与问题反馈',
      icon: 'ri:user-location-line',
      iconColor: '#ff6b6b',
      enabled: true,
      order: 3,
      link: WEB_LINKS.COMMUNITY
    },
    {
      name: '哔哩哔哩',
      description: '技术分享与交流',
      icon: 'ri:bilibili-line',
      iconColor: '#FB7299',
      enabled: true,
      order: 4,
      link: WEB_LINKS.BILIBILI
    }
  ],
  // 快速链接
  quickLinks: [
    {
      name: '登录',
      enabled: true,
      order: 1,
      routeName: 'Login'
    },
    {
      name: '注册',
      enabled: true,
      order: 2,
      routeName: 'Register'
    },
    {
      name: '忘记密码',
      enabled: true,
      order: 3,
      routeName: 'ForgetPassword'
    },
    {
      name: '个人中心',
      enabled: true,
      order: 4,
      routeName: 'UserCenter'
    }
  ]
}

export default Object.freeze(fastEnterConfig)
