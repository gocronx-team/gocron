# GoCronX Admin

<p align="center">
  <img src="https://img.shields.io/badge/Vue-3.5-brightgreen.svg" alt="Vue">
  <img src="https://img.shields.io/badge/TypeScript-5.6-blue.svg" alt="TypeScript">
  <img src="https://img.shields.io/badge/Element_Plus-2.11-409EFF.svg" alt="Element Plus">
  <img src="https://img.shields.io/badge/license-MIT-green.svg" alt="License">
</p>

<p align="center">
  简体中文 | <a href="./README.md">English</a>
</p>

## 简介

一个现代化的、生产就绪的 Vue3 后台管理系统模板，专为 GoCronX 项目打造。采用最新的前端技术栈构建，适合快速开发企业级后台管理系统。

## 特性

✨ **现代技术栈**: Vue3 + TypeScript + Vite + Element Plus + Tailwind CSS

🎨 **精美界面**: 现代化设计，流畅的动画和过渡效果

🌓 **暗黑模式**: 内置亮色/暗色主题切换

📱 **响应式**: 移动端友好的响应式设计

🔐 **权限系统**: 完整的权限控制（路由级 & 按钮级）

🛠️ **开发友好**: 丰富的组件和 Hooks，快速开发

📦 **快速开始**: 一键清理演示数据，立即开始开发

## 技术栈

- **框架**: Vue 3.5 (Composition API)
- **语言**: TypeScript 5.6
- **构建工具**: Vite 7
- **UI 库**: Element Plus 2.11
- **CSS 框架**: Tailwind CSS 4
- **状态管理**: Pinia 3
- **路由**: Vue Router 4
- **HTTP 客户端**: Axios
- **图表**: ECharts 6
- **代码质量**: ESLint + Prettier + Stylelint + Husky

## 快速开始

```bash
# 克隆仓库
git clone https://github.com/gocronx/gocronx-admin.git
cd gocronx-admin

# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev

# 生产环境打包
pnpm build
```

## 清理演示数据

移除所有演示页面和数据，获得一个干净的基础项目：

```bash
pnpm clean:dev
```

## 项目结构

```
gocronx-admin/
├── public/                 # 静态资源
├── src/
│   ├── api/               # API 请求
│   ├── assets/            # 图片、样式等
│   ├── components/        # 可复用组件
│   │   ├── core/         # 核心 UI 组件
│   │   └── business/     # 业务组件
│   ├── config/           # 应用配置
│   ├── directives/       # 自定义指令
│   ├── hooks/            # 组合式函数
│   ├── locales/          # 国际化翻译
│   ├── router/           # 路由配置
│   ├── store/            # Pinia 状态管理
│   ├── types/            # TypeScript 类型
│   ├── utils/            # 工具函数
│   ├── views/            # 页面组件
│   ├── App.vue           # 根组件
│   └── main.ts           # 入口文件
├── .env                   # 环境变量
├── .env.development       # 开发环境变量
├── .env.production        # 生产环境变量
├── vite.config.ts        # Vite 配置
└── package.json          # 依赖配置
```

## 使用指南

### 创建新页面

1. 在 `src/views/` 创建页面组件
2. 在 `src/router/modules/` 添加路由
3. 在 `src/api/` 添加 API 接口

### 使用内置组件

```vue
<template>
  <!-- 带分页的表格 -->
  <art-table 
    :columns="columns" 
    :data="tableData"
    :loading="loading"
  />
  
  <!-- 带验证的表单 -->
  <art-form 
    :config="formConfig"
    v-model="formData"
  />
  
  <!-- 图表 -->
  <art-line-chart :data="chartData" />
</template>
```

### 使用 Hooks

```typescript
import { useTable } from '@/hooks'

// 自动处理表格加载、分页等
const { tableData, loading, loadData } = useTable(apiFunction)
```

## 浏览器支持

现代浏览器（Chrome、Firefox、Safari、Edge）

## 贡献

欢迎贡献！请随时提交 Pull Request。

## 致谢

本模板基于 [art-design-pro](https://github.com/Daymychen/art-design-pro) 开发。感谢原作者的优秀工作！