import { defineConfig } from "vitepress";

const description = "gocron - Lightweight distributed scheduled task management system";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "gocron",
  description: description,
  cleanUrls: false,
  lastUpdated: true,
  ignoreDeadLinks: true,

  // 国际化配置
  locales: {
    root: {
      label: 'English',
      lang: 'en-US',
      link: '/en/',
      themeConfig: {
        nav: [
          { text: 'Home', link: '/en/' },
          { text: 'Guide', link: '/en/guide/introduction' },
        ],
        sidebar: {
          '/en/guide/': [
            {
              text: 'Getting Started',
              items: [
                { text: 'Introduction', link: '/en/guide/introduction' },
                { text: 'Quick Start', link: '/en/guide/quick-start' },
                { text: 'Configuration', link: '/en/guide/configuration' },
              ]
            },
            {
              text: 'Core Features',
              items: [
                { text: 'Scheduled Tasks', link: '/en/guide/scheduled-tasks' },
                { text: 'Agent Auto-Registration', link: '/en/guide/agent-registration' },
                { text: 'Log Management', link: '/en/guide/log-management' },
                { text: 'Notifications', link: '/en/guide/notifications' },
                { text: 'AI Features', link: '/en/guide/ai-features' },
                { text: 'API Documentation', link: '/en/guide/api' },
              ]
            },
            {
              text: 'Deployment',
              items: [
                { text: 'Kubernetes (Helm)', link: '/en/guide/kubernetes' },
                { text: 'High Availability', link: '/en/guide/high-availability' },
              ]
            },
            {
              text: 'Security',
              items: [
                { text: 'Two-Factor Authentication (2FA)', link: '/en/guide/security-2fa' },
                { text: 'TLS Mutual Authentication', link: '/en/guide/security-tls' },
                { text: 'Security hardening', link: '/en/guide/security-hardening' },
                { text: 'Password Reset (CLI)', link: '/en/guide/password-reset' },
              ]
            },
            {
              text: 'About',
              items: [
                { text: 'Contributing Guide', link: '/en/guide/contributing' },
                { text: 'About Project', link: '/en/guide/about' },
              ]
            }
          ]
        },
        outline: {
          level: [2, 4],
          label: 'On this page',
        },
        docFooter: {
          prev: 'Previous',
          next: 'Next',
        },
        darkModeSwitchLabel: 'Theme',
        sidebarMenuLabel: 'Menu',
        returnToTopLabel: 'Return to top',
        lastUpdatedText: 'Last updated',
        editLink: {
          text: 'Edit this page on GitHub',
          pattern: 'https://github.com/gocronx-team/gocron/edit/master/docs/:path',
        },
      }
    },
    zh: {
      label: '简体中文',
      lang: 'zh-CN',
      link: '/zh/',
      themeConfig: {
        nav: [
          { text: '首页', link: '/zh/' },
          { text: '指南', link: '/zh/guide/introduction' },
        ],
        sidebar: {
          '/zh/guide/': [
            {
              text: '开始',
              items: [
                { text: '简介', link: '/zh/guide/introduction' },
                { text: '快速开始', link: '/zh/guide/quick-start' },
                { text: '配置文件', link: '/zh/guide/configuration' },
              ]
            },
            {
              text: '核心功能',
              items: [
                { text: '定时任务', link: '/zh/guide/scheduled-tasks' },
                { text: 'Agent 自动注册', link: '/zh/guide/agent-registration' },
                { text: '日志管理', link: '/zh/guide/log-management' },
                { text: '通知配置', link: '/zh/guide/notifications' },
                { text: 'AI 功能', link: '/zh/guide/ai-features' },
                { text: 'API 文档', link: '/zh/guide/api' },
              ]
            },
            {
              text: '部署',
              items: [
                { text: 'Kubernetes (Helm)', link: '/zh/guide/kubernetes' },
                { text: '高可用部署', link: '/zh/guide/high-availability' },
              ]
            },
            {
              text: '安全',
              items: [
                { text: '双因素认证 (2FA)', link: '/zh/guide/security-2fa' },
                { text: 'TLS 双向认证', link: '/zh/guide/security-tls' },
                { text: '安全加固', link: '/zh/guide/security-hardening' },
                { text: '密码重置 (CLI)', link: '/zh/guide/password-reset' },
              ]
            },
            {
              text: '关于',
              items: [
                { text: '贡献指南', link: '/zh/guide/contributing' },
                { text: '关于项目', link: '/zh/guide/about' },
              ]
            }
          ]
        },
        outline: {
          level: [2, 4],
          label: '本页导航',
        },
        docFooter: {
          prev: '上一页',
          next: '下一页',
        },
        darkModeSwitchLabel: '主题',
        sidebarMenuLabel: '菜单',
        returnToTopLabel: '返回顶部',
        lastUpdatedText: '上次更新时间',
        editLink: {
          text: '在 GitHub 上编辑此页',
          pattern: 'https://github.com/gocronx-team/gocron/edit/master/docs/:path',
        },
      }
    }
  },

  head: [
    ["link", { rel: "icon", type: "image/x-icon", href: "/favicon.ico" }],
    ["meta", { property: "og:type", content: "website" }],
    ["meta", { property: "og:title", content: "gocron | Distributed Scheduled Task System" }],
    ["meta", { property: "og:site_name", content: "gocron" }],
    ["meta", { property: "og:description", content: description }],
    ["meta", { name: "description", content: description }],
    ["meta", { name: "author", content: "gocron" }],
    ["meta", { name: "keywords", content: "gocron,定时任务,scheduled tasks,cron,distributed" }],
  ],

  markdown: {
    lineNumbers: true,
    image: {
      lazyLoading: true,
    },
  },

  sitemap: {
    hostname: "https://gocron.io",
  },

  themeConfig: {
    socialLinks: [
      {
        icon: "github",
        link: "https://github.com/gocronx-team/gocron",
      },
    ],
    search: {
      provider: "local",
      options: {
        locales: {
          zh: {
            translations: {
              button: {
                buttonText: '搜索文档',
                buttonAriaLabel: '搜索文档'
              },
              modal: {
                noResultsText: '无法找到相关结果',
                resetButtonTitle: '清除查询条件',
                footer: {
                  selectText: '选择',
                  navigateText: '切换'
                }
              }
            }
          }
        }
      }
    },
  },
});
