# gocron 前端迁移到 gocronx-admin 模板

> 把 `web/vue/` 的业务逻辑逐页搬到 `web/gocronx-admin/`，利用 art-design-pro 风格模板的成品视觉。
>
> 后端零改动、老前端 `web/vue/` 不动，新前端独立端口运行。

---

## 🏗 目录结构总览

```
gocron/
├── web/
│   ├── vue/                   ← 老前端（保持不动、用户随时回退）
│   └── gocronx-admin/         ← 新前端，本次迁移目标
│       ├── src/
│       │   ├── api/           ← gocron API 调用，.ts 带类型
│       │   ├── views/         ← 页面，按业务 namespace 分组
│       │   │   ├── auth/     （含 login，已在模板里）
│       │   │   ├── dashboard/（模板自带，我们会改造成 gocron 的 statistics）
│       │   │   ├── system/   （模板自带用户/角色/菜单；gocron 版在这里加 task/host/template 等）
│       │   │   └── ...
│       │   ├── router/        ← vue-router 动态加载
│       │   ├── store/         ← pinia
│       │   ├── locales/       ← vue-i18n
│       │   ├── utils/http.ts  ← axios 封装（**所有 API 走这个**）
│       │   └── hooks/         ← useTable / useForm 等模板亮点
│       ├── .env               ← 端口 3006
│       ├── .env.development   ← 代理到 gocron 后端 http://localhost:5920
│       └── vite.config.ts
├── cmd/gocron/                ← 后端，**不动**
├── internal/                   ← 后端，**不动**
└── ...
```

---

## 🚀 快速启动

```bash
# 后端（air 或 go run）：默认 5920 端口
air

# 新前端：独立 3006 端口，代理 /api 到 5920
cd web/gocronx-admin
pnpm install      # 首次
pnpm dev          # 启动 http://localhost:3006

# 老前端继续在 8080（如果要对比）
cd ../vue
pnpm dev
```

---

## 📋 迁移模板（每个页面 agent 按这个走）

### Step 1 — 读**老页面**
- 读 `web/vue/src/pages/<page>.vue` 完整代码
- 读 `web/vue/src/api/<module>.js` 找到该页用的 API
- 列出等价行为清单（字段 / 校验 / API 调用 / i18n key / 跳转逻辑）

### Step 2 — 在新前端复刻 API 层
- `web/gocronx-admin/src/api/<module>.ts` 新建或补充
- 用模板自带的 `request` 工具：
  ```ts
  import request from '@/utils/http'
  export function fetchTaskList(params: { page: number; page_size: number }) {
    return request.get<{ data: Task[]; total: number }>({
      url: '/api/task',
      params
    })
  }
  ```
- 类型可以先写 `any`，**不要为了完美类型耽误进度**
- 端点 / 请求参数 / 响应结构**复用**老版本（后端没动）

### Step 3 — 建页面
- 在 `src/views/<namespace>/<page>/index.vue` 新建
- **优先借模板现有 hook**：
  - **useTable**（表格 + 分页 + loading 一把梭）
  - **useForm**（表单封装）
- **照抄模板页面结构**（比如要建 task list，去看 `views/system/user/index.vue` 怎么写的，改字段/API 即可）

### Step 4 — 注册路由
- `src/router/modules/` 下新增路由定义或扩展已有
- 路由元数据（title / icon）配好，Sidebar 自动出菜单
- **不走动态菜单**（env `VITE_ACCESS_MODE=frontend`），静态路由 + `roles` 控制可见性

### Step 5 — i18n
- `src/locales/langs/zh-CN.ts` / `en-US.ts` 追加 key
- 现有老前端 `web/vue/src/locales/*.js` 的 key **可以直接搬过来**（格式要转成 TS 对象）

### Step 6 — 本地验证
```bash
pnpm build     # TS + vite build 必须过
# 访问 http://localhost:3006 手动点一遍该页面所有功能
```

### Step 7 — Commit
```
feat(gocronx-admin): migrate <page> from web/vue

- add api/<module>.ts with typed endpoints
- add views/<ns>/<page>/index.vue using useTable + ArtForm
- register route + sidebar entry
- add i18n keys (reused from web/vue)
```

---

## 🧩 关键复用点（**不要重新造轮子**）

| 我们要干的事 | 模板里已有的工具 | 位置 |
|------------|----------------|------|
| 列表 + 分页 + 筛选 | `useTable` hook | `src/hooks/useTable.ts` |
| 表单校验 | `useForm` + Element Plus `el-form` | 组件里用 Element Plus 原生 API 即可 |
| 通知 toast | `ElMessage`（注意：**此项目可以直接用 Element Plus，不用 facade**） | element-plus 全局装 |
| 确认对话框 | `ElMessageBox.confirm` | 同上 |
| 图标 | `@iconify/vue` + `@element-plus/icons-vue` | 优先 iconify（线性图标库多） |
| 图表 | ECharts 6 已装 | 统计页用得上 |
| 国际化 | vue-i18n | 已配 |

---

## ⚠️ 约定和纪律

### DO
- ✅ 用 TypeScript。不追求完美类型，但文件 `.ts` / `.vue + <script setup lang="ts">`
- ✅ API endpoint 跟 gocron 后端现有路径**完全一致**（后端没动）
- ✅ 复用模板已有的 layout / sidebar / header
- ✅ 新功能（useTable / useForm）学会用
- ✅ 每个 ticket 独立 commit，分支名 `gocronx-<page-name>`

### DON'T
- ❌ 不要改后端任何 Go 文件
- ❌ 不要动 `web/vue/` 下任何文件（老前端要保留）
- ❌ 不要把业务写到模板的 `examples/` 里（那是 demo）
- ❌ 不要大改 `App.vue` / `router/core/` / `router/guards/` 等核心
- ❌ 不要装新 npm 包（模板已有的应有尽有）

---

## 📦 迁移优先级（按 user flow 排序）

**Phase 1: 登录认证链路**（最先，不跑通其他页面打不开）
1. **LOGIN** — `views/auth/login` 已有，接 gocron `/api/user/login` + JWT
2. **USER INFO** — 登录后拿 user info，set Pinia store
3. **ROUTES GUARD** — 未登录跳 /login，已登录不跳 /login

**Phase 2: 只读页面**（验证 DataTable 和路由）
4. **DASHBOARD/STATISTICS** — 改造模板 `views/dashboard/analysis` 成 gocron 统计
5. **AUDIT LOG** — 新建 `views/system/audit-log`
6. **LOGIN LOG** — 新建 `views/system/login-log`

**Phase 3: 基础 CRUD**
7. **USER LIST + EDIT** — 复用模板 `views/system/user` 的壳 + gocron API
8. **HOST LIST + EDIT** — 新建 `views/host`
9. **EDIT MY PASSWORD / 2FA**（登录相关）

**Phase 4: 任务核心**
10. **TASK LIST** — 新建 `views/task/list`（最复杂 table）
11. **TASK EDIT** — 新建 `views/task/edit`（最大 form）
12. **TASK LOG** — `views/task/log`
13. **TEMPLATE LIST/EDIT** — `views/template/*`

**Phase 5: 系统设置**
14. **NOTIFICATION (email/slack/webhook)** — `views/system/notification/*`
15. **LOG RETENTION** — `views/system/log-retention`

**Phase 6: Install + 收尾**
16. **INSTALL** — 首次部署页
17. 删除模板里的 demo 模块（`article` / `widgets` / `examples` 等）
18. 把后端 embed 指向新 dist

---

## 🧵 多 agent 并行原则

### 可并行
- 同 Phase 内、不同独占文件的 ticket
- 示例：AUDIT-LOG + LOGIN-LOG + STATISTICS 可三个 agent 同时干

### 必须串行
- Phase 1（登录链路）必须最先做、单人串行
- 路由 config 变更涉及多个文件时要排队避免冲突

### Agent 交付必带
- 分支名 `gocronx-<page>`（扁平，避免子路径冲突）
- Commit message 按上面格式
- 等价行为清单
- `pnpm build` 跑过证据
- 手动测试清单
