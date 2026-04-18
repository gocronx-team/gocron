# GoCronX Admin

<p align="center">
  <img src="https://img.shields.io/badge/Vue-3.5-brightgreen.svg" alt="Vue">
  <img src="https://img.shields.io/badge/TypeScript-5.6-blue.svg" alt="TypeScript">
  <img src="https://img.shields.io/badge/Element_Plus-2.11-409EFF.svg" alt="Element Plus">
  <img src="https://img.shields.io/badge/license-MIT-green.svg" alt="License">
</p>

<p align="center">
  <a href="./README.zh-CN.md">简体中文</a> | English
</p>

## Introduction

A modern, production-ready Vue3 admin template for GoCronX projects. Built with the latest frontend technologies, perfect for building enterprise-level admin systems.

## Features

✨ **Modern Stack**: Vue3 + TypeScript + Vite + Element Plus + Tailwind CSS

🎨 **Beautiful UI**: Modern design with smooth animations and transitions

🌓 **Dark Mode**: Built-in light/dark theme switching

📱 **Responsive**: Mobile-friendly responsive design

🔐 **Permission**: Complete permission system (route & button level)

🛠️ **Developer Friendly**: Rich components and hooks for rapid development

📦 **Clean Start**: One command to remove demo data and start fresh

## Tech Stack

- **Framework**: Vue 3.5 (Composition API)
- **Language**: TypeScript 5.6
- **Build Tool**: Vite 7
- **UI Library**: Element Plus 2.11
- **CSS Framework**: Tailwind CSS 4
- **State Management**: Pinia 3
- **Router**: Vue Router 4
- **HTTP Client**: Axios
- **Charts**: ECharts 6
- **Code Quality**: ESLint + Prettier + Stylelint + Husky

## Quick Start

```bash
# Clone repository
git clone https://github.com/gocronx/gocronx-admin.git
cd gocronx-admin

# Install dependencies
pnpm install

# Start development server
pnpm dev

# Build for production
pnpm build
```

## Clean Demo Data

Remove all demo pages and data to start with a clean base:

```bash
pnpm clean:dev
```

## Project Structure

```
gocronx-admin/
├── public/                 # Static assets
├── src/
│   ├── api/               # API requests
│   ├── assets/            # Images, styles, etc.
│   ├── components/        # Reusable components
│   │   ├── core/         # Core UI components
│   │   └── business/     # Business components
│   ├── config/           # App configuration
│   ├── directives/       # Custom directives
│   ├── hooks/            # Composable hooks
│   ├── locales/          # i18n translations
│   ├── router/           # Route configuration
│   ├── store/            # Pinia stores
│   ├── types/            # TypeScript types
│   ├── utils/            # Utility functions
│   ├── views/            # Page components
│   ├── App.vue           # Root component
│   └── main.ts           # Entry point
├── .env                   # Environment variables
├── .env.development       # Development env
├── .env.production        # Production env
├── vite.config.ts        # Vite configuration
└── package.json          # Dependencies
```

## Usage

### Create New Page

1. Create page component in `src/views/`
2. Add route in `src/router/modules/`
3. Add API in `src/api/`

### Use Built-in Components

```vue
<template>
  <!-- Table with pagination -->
  <art-table 
    :columns="columns" 
    :data="tableData"
    :loading="loading"
  />
  
  <!-- Form with validation -->
  <art-form 
    :config="formConfig"
    v-model="formData"
  />
  
  <!-- Charts -->
  <art-line-chart :data="chartData" />
</template>
```

### Use Hooks

```typescript
import { useTable } from '@/hooks'

// Auto-handle table loading, pagination, etc.
const { tableData, loading, loadData } = useTable(apiFunction)
```

## Browser Support

Modern browsers (Chrome, Firefox, Safari, Edge)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Credits

This template is based on [art-design-pro](https://github.com/Daymychen/art-design-pro). Thanks to the original author for the excellent work!