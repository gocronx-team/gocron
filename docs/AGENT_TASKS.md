# gocron shadcn 迁移 — Agent 任务清单

> 每个 ticket 可独立派给一个 agent。认领时在 `Assignee` 填 agent 标识，完成后标 ✅。
>
> 必读先修：[SHADCN_MIGRATION.md](./SHADCN_MIGRATION.md)

---

## 🚦 Ticket 状态
- `⚪ TODO`：未认领
- `🟡 IN PROGRESS`：某 agent 正在做
- `✅ DONE`：已合并到 `feat/shadcn-migration`
- `🔴 BLOCKED`：依赖未完成

---

## Phase 0 — Foundation（必须最先做，串行）

### TICKET-PHASE-0 — Tailwind + shadcn-vue 基础设施

**Status**: ⚪ TODO
**Assignee**: —
**Branch**: `feat/shadcn-migration/phase-0-foundation`
**Depends**: —
**Blocks**: 所有后续 ticket

**目标**：在 `feat/shadcn-migration` 基础上装好 Tailwind + shadcn-vue，和 Element Plus 共存，现有所有页面**视觉零变化**。

**操作**：
```bash
cd web/vue
pnpm add -D tailwindcss postcss autoprefixer
pnpm add class-variance-authority clsx tailwind-merge
pnpm add lucide-vue-next
pnpm add vue-sonner
pnpm add @tanstack/vue-table
pnpm add vee-validate @vee-validate/zod zod
pnpm add -D @types/node  # 如还没装
```

**配置文件**：

`web/vue/tailwind.config.js`（新建）:
```js
/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  prefix: 'tw-',                          // ← 关键
  corePlugins: { preflight: false },      // ← 关键
  theme: {
    extend: {
      colors: {
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))',
        },
        // ... 完整 shadcn color system
      },
      borderRadius: {
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)',
      },
    },
  },
  plugins: [],
}
```

`web/vue/postcss.config.js`（新建）:
```js
export default {
  plugins: { tailwindcss: {}, autoprefixer: {} },
}
```

`web/vue/src/assets/tailwind.css`（新建）:
```css
@tailwind utilities;
/* 不 import base 和 components，避免 preflight 副作用 */

@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --primary: 222.2 47.4% 11.2%;
    --primary-foreground: 210 40% 98%;
    --border: 214.3 31.8% 91.4%;
    --input: 214.3 31.8% 91.4%;
    --ring: 222.2 84% 4.9%;
    --radius: 0.5rem;
    /* ... 完整 shadcn CSS variables */
  }
  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    /* ... */
  }
}
```

`web/vue/src/main.js`：在 `import 'element-plus/dist/index.css'` **之后**加：
```js
import './assets/tailwind.css'
```

**shadcn-vue 初始化**：
```bash
pnpm dlx shadcn-vue@latest init
```
选择：Default / Slate / CSS variables / aliases `@/components` `@/composables` / tsconfig path

**新建工具文件**：

`web/vue/src/lib/utils.js`:
```js
import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs) {
  return twMerge(clsx(inputs))
}
```

**验证清单**：
- [ ] `pnpm dev` 启动，所有现有页面视觉和迁移前完全一致（截图对比）
- [ ] `pnpm build` 成功，dist 产物正常
- [ ] 访问 `/` 登录页、任务列表、模板列表等核心页面功能正常
- [ ] 控制台无新报错
- [ ] 新增依赖合计 bundle 增长 < 50KB gzip

**交付**：
```
chore(shadcn): add tailwind + shadcn-vue infrastructure alongside element plus
```

---

## Phase 1 — 公共基础设施（串行，Phase 0 完成后）

### TICKET-PHASE-1-UI-PRIMITIVES — 批量添加 shadcn 基础组件

**Status**: ⚪ TODO
**Assignee**: —
**Branch**: `feat/shadcn-migration/phase-1-ui-primitives`
**Depends**: TICKET-PHASE-0
**Blocks**: 所有页面 ticket

**目标**：一次性把所有页面 ticket 会用到的 shadcn 基础组件都 add 进项目。

**操作**：
```bash
cd web/vue
pnpm dlx shadcn-vue@latest add button input label card badge \
  dialog alert-dialog select popover tooltip tabs switch \
  checkbox radio-group separator dropdown-menu sonner \
  form toaster skeleton avatar
```

产物：`src/components/ui/**` 一批 `.vue` 文件。

**验证**：
- [ ] 所有组件 import 路径 `@/components/ui/<name>` 能解析
- [ ] 每个组件在 `/demo-shadcn`（见下 ticket）渲染正常

---

### TICKET-PHASE-1-NOTIFY-FACADE — 通知统一入口

**Status**: ⚪ TODO
**Assignee**: —
**Branch**: `feat/shadcn-migration/phase-1-notify-facade`
**Depends**: TICKET-PHASE-0
**Blocks**: 使用 `ElMessage` / `ElMessageBox` 的所有页面 ticket

**目标**：建一个 `useNotify()` 包装，现在代理 Element Plus，未来切 sonner。**一次性把全站 `ElMessage.*` / `ElMessageBox.*` 调用替换成 `useNotify().xxx`**。

**新建**：
`web/vue/src/composables/useNotify.js`:
```js
import { ElMessage, ElMessageBox } from 'element-plus'
import { toast as sonnerToast } from 'vue-sonner'

// 通过 localStorage 切换 backend，方便测试阶段对比
const BACKEND = () => localStorage.getItem('notify.backend') || 'element'

export function useNotify() {
  const backend = BACKEND()
  const useSonner = backend === 'sonner'

  return {
    success(msg) { useSonner ? sonnerToast.success(msg) : ElMessage.success(msg) },
    error(msg)   { useSonner ? sonnerToast.error(msg)   : ElMessage.error(msg) },
    info(msg)    { useSonner ? sonnerToast(msg)         : ElMessage.info(msg) },
    warning(msg) { useSonner ? sonnerToast.warning(msg) : ElMessage.warning(msg) },
    async confirm(message, title = '提示') {
      if (useSonner) {
        // 触发全局 AlertDialog（见下）
        return window.__shadcnConfirm?.(message, title) ?? true
      }
      try { await ElMessageBox.confirm(message, title); return true }
      catch { return false }
    },
  }
}
```

**全局替换**（grep 定位所有调用点）：
```bash
grep -rln "ElMessage\|ElMessageBox" src/
# 每个文件替换 import 为 useNotify
# 调用 ElMessage.success(x) → notify.success(x)
# 调用 ElMessageBox.confirm(x) → await notify.confirm(x)
```

**估计替换点**：~40 处。

**验证**：
- [ ] 全站所有 toast 功能正常（触发错误、成功提示等）
- [ ] `localStorage.setItem('notify.backend', 'sonner')` 切到 shadcn Toast，所有提示变样式但功能一致
- [ ] 切回 `element`，恢复原样

**交付**：
```
refactor(shadcn): introduce useNotify facade and replace ElMessage/ElMessageBox callsites
```

---

### TICKET-PHASE-1-DATA-TABLE — 通用 DataTable 组件

**Status**: ⚪ TODO
**Assignee**: —
**Branch**: `feat/shadcn-migration/phase-1-data-table`
**Depends**: TICKET-PHASE-0, TICKET-PHASE-1-UI-PRIMITIVES
**Blocks**: 所有含 table 的页面 ticket（Phase 4-6 多数）

**目标**：基于 `@tanstack/vue-table` 封装 `<DataTable>` 组件，API 对齐原 `el-table` 常用能力：
- 列定义（column defs）
- 分页
- loading 状态
- 行选择（单选 / 多选）
- 排序
- 空态

**新建**：
```
src/components/ui/data-table/
  DataTable.vue          // 主组件
  DataTablePagination.vue
  DataTableColumnHeader.vue
  DataTableToolbar.vue
  types.js               // ColumnDef 类型定义
```

**API 设计（参考）**：
```vue
<DataTable
  :columns="columns"
  :data="tasks"
  :loading="loading"
  :total="total"
  :page="page"
  :page-size="pageSize"
  selectable
  @update:page="onPageChange"
  @update:selected="onSelect"
/>
```

`columns` 用 TanStack 格式：
```js
const columns = [
  { accessorKey: 'name', header: '任务名' },
  { accessorKey: 'spec', header: 'cron' },
  {
    id: 'status',
    header: '状态',
    cell: ({ row }) => h(StatusBadge, { value: row.original.status })
  },
  {
    id: 'actions',
    cell: ({ row }) => h(TaskActionsMenu, { task: row.original })
  },
]
```

**参考实现**：
- shadcn 官方 tasks example: https://ui.shadcn.com/examples/tasks
- shadcn-vue 官方 data-table doc: https://www.shadcn-vue.com/docs/components/data-table

**验证**：
- [ ] `/demo-shadcn/data-table` 路由展示 DataTable 跑通 mock 数据
- [ ] 分页、排序、选择都工作
- [ ] 移动端窄屏时表格可横向滚动

**交付**：
```
feat(shadcn): add reusable DataTable component based on tanstack/vue-table
```

---

### TICKET-PHASE-1-DARK-MODE — 暗色模式开关

**Status**: ⚪ TODO
**Assignee**: —
**Branch**: `feat/shadcn-migration/phase-1-dark-mode`
**Depends**: TICKET-PHASE-0
**Blocks**: —

**目标**：加一个全局 dark mode toggle（仅影响 shadcn 组件，Element Plus 组件**暂不暗色**，Phase 7 才全面生效）。

**新建**：
`src/composables/useDarkMode.js`:
```js
import { ref, watch } from 'vue'

const isDark = ref(localStorage.getItem('theme') === 'dark')

export function useDarkMode() {
  watch(isDark, v => {
    document.documentElement.classList.toggle('dark', v)
    localStorage.setItem('theme', v ? 'dark' : 'light')
  }, { immediate: true })

  return { isDark, toggle: () => { isDark.value = !isDark.value } }
}
```

暂不集成到 header（Phase 7 再做），只做 composable。

---

### TICKET-PHASE-1-DEMO-PAGE — 迁移者参考页

**Status**: ⚪ TODO
**Assignee**: —
**Branch**: `feat/shadcn-migration/phase-1-demo`
**Depends**: TICKET-PHASE-1-UI-PRIMITIVES, TICKET-PHASE-1-DATA-TABLE
**Blocks**: —

**目标**：建 `/demo-shadcn` 路由，展示所有 shadcn 组件的用法，作为后续 ticket 的**参考样板**。

**新建**：
- `src/pages/demo/shadcn.vue`（展示 Button / Input / Form / Dialog / DataTable / Toast / Dark mode 等）
- `src/router/index.js` 追加路由（**追加，不改现有**）

路由只在 dev 模式或有 query `?demo=1` 时显示，避免暴露给最终用户。

---

## Phase 2 — 已有组件改造（可与 Phase 3 并行）

### TICKET-CRON-PREVIEW — CronPreview / HeatmapSvg 改用 shadcn

**Status**: ⚪ TODO
**Assignee**: —
**Branch**: `feat/shadcn-migration/cron-preview`
**Depends**: Phase 1 完成
**Blocks**: —

**独占文件**：
- `src/components/common/CronPreview.vue`
- `src/components/common/HeatmapSvg.vue`

**改动点**：
- `<el-icon><WarningFilled /></el-icon>` → `<WarningCircle class="tw-size-4" />`（lucide）
- `<el-tag>` → shadcn `<Badge>`
- `.section` 等 scoped CSS → Tailwind utility
- 移除 Element Plus icon import，换 lucide

**验证**：
- [ ] 在任务编辑页和模板编辑页 CronPreview 视觉 / 行为不变
- [ ] 错误态、loading、heatmap 都正常

---

## Phase 3 — 独立页面（可并行，认领顺序不限）

所有 Phase 3 ticket 互相**无依赖**，可 3 个 agent 同时干。

---

### TICKET-LOGIN — `pages/user/login.vue`

**Status**: ⚪ TODO
**Assignee**: —
**Branch**: `feat/shadcn-migration/login`
**Depends**: Phase 1 完成
**Parallel-safe with**: TICKET-INSTALL, TICKET-TWOFACTOR, TICKET-EDIT-MY-PASSWORD, TICKET-EDIT-PASSWORD

**独占文件**：`web/vue/src/pages/user/login.vue`

**老代码行为清单（必须保留）**：
- 渲染：用户名 / 密码 / 2FA code（可选）/ 登录按钮
- 用户名为空 → 红字"请输入用户名"
- 密码为空 → 红字"请输入密码"
- 2FA code 为空且后端要求时 → 红字
- 登录成功 → 跳转 `/task`（或 `redirect` query 参数）
- 登录失败 → toast 错误
- 支持中英文切换

**API**：`POST /api/user/login`（`api/user.js`）

**参考 Block**：https://www.shadcn-vue.com/blocks 选一个 auth 相关 block

**验证清单**：见 [SHADCN_MIGRATION.md §8](./SHADCN_MIGRATION.md#8-测试策略)

---

### TICKET-INSTALL — `pages/install/index.vue`

**Status**: ⚪ TODO
**Assignee**: —
**Branch**: `feat/shadcn-migration/install`
**Depends**: Phase 1 完成

**独占文件**：`web/vue/src/pages/install/index.vue`

**老代码行为清单**：
- 渲染：多步表单（DB 配置 / 管理员配置 / 配置预览）
- 所有字段校验
- 提交 → 后端初始化，成功后跳 `/user/login`

**API**：`POST /api/install/store`

**注意**：**只在首次部署会见到的页面**，但一旦坏了新部署无法初始化。必须仔细测试。

---

### TICKET-EDIT-MY-PASSWORD — `pages/user/editMyPassword.vue`

**Status**: ⚪ TODO
**Assignee**: —
**Branch**: `feat/shadcn-migration/edit-my-password`
**Depends**: Phase 1

**独占文件**：`web/vue/src/pages/user/editMyPassword.vue`

**行为**：旧密码 / 新密码 / 确认新密码 → 提交。

---

### TICKET-EDIT-PASSWORD — `pages/user/editPassword.vue`

**Status**: ⚪ TODO
**Assignee**: —
**Branch**: `feat/shadcn-migration/edit-password`
**Depends**: Phase 1

**独占文件**：`web/vue/src/pages/user/editPassword.vue`

**行为**：管理员改某用户密码（无需旧密码）。

---

### TICKET-TWOFACTOR — `pages/user/twoFactor.vue`

**Status**: ⚪ TODO
**Assignee**: —
**Branch**: `feat/shadcn-migration/twofactor`
**Depends**: Phase 1

**独占文件**：`web/vue/src/pages/user/twoFactor.vue`

**行为**：显示 TOTP 二维码 + 验证码输入 + 启用 / 关闭按钮。

---

## Phase 4 — 只读监控页（可并行）

### TICKET-LOGIN-LOG — `pages/system/loginLog.vue`

**Status**: ⚪ TODO
**Independent file**: `web/vue/src/pages/system/loginLog.vue`
**Depends**: TICKET-PHASE-1-DATA-TABLE
**Parallel-safe with**: TICKET-AUDIT-LOG, TICKET-STATISTICS

**行为**：登录日志列表 + 分页 + 用户/IP 筛选。

---

### TICKET-AUDIT-LOG — `pages/system/auditLog.vue`

**Status**: ⚪ TODO
**Independent file**: `web/vue/src/pages/system/auditLog.vue`
**Depends**: TICKET-PHASE-1-DATA-TABLE

**行为**：审计日志列表 + 分页 + Module/Action 筛选 + 详情 Dialog。

---

### TICKET-STATISTICS — `pages/statistics/index.vue`

**Status**: ⚪ TODO
**Independent file**: `web/vue/src/pages/statistics/index.vue`
**Depends**: Phase 1

**行为**：Overview 数据面板（Card 展示）。

---

## Phase 5 — 用户 / 系统管理（可并行）

### TICKET-USER-LIST — `pages/user/list.vue`
### TICKET-USER-EDIT — `pages/user/edit.vue`
### TICKET-HOST-LIST — `pages/host/list.vue`
### TICKET-HOST-EDIT — `pages/host/edit.vue`
### TICKET-NOTIFICATION-EMAIL — `pages/system/notification/email.vue`
### TICKET-NOTIFICATION-SLACK — `pages/system/notification/slack.vue`
### TICKET-NOTIFICATION-WEBHOOK — `pages/system/notification/webhook.vue`
### TICKET-NOTIFICATION-TAB — `pages/system/notification/tab.vue`
### TICKET-LOG-RETENTION — `pages/system/logRetention.vue`

各 ticket 格式同 Phase 3 / 4。独占各自文件，互不冲突。**至多 3 个 agent 并行**（避免 shared file 冲突率过高）。

---

## Phase 6 — 任务主战场（半并行）

### TICKET-TASK-LIST — `pages/task/list.vue`
**Status**: ⚪ TODO
**Depends**: TICKET-PHASE-1-DATA-TABLE，**建议这个 phase 里第一个做**
**Blocks**: TICKET-TASK-LOG（沿用相似 pattern）

---

### TICKET-TASK-EDIT — `pages/task/edit.vue`
**Status**: ⚪ TODO
**Complexity**: **最高**，~1100 行超大表单
**Depends**: Phase 1
**Blocks**: TICKET-TEMPLATE-EDIT（相似 pattern）

**建议**：用 v0.dev 描述 "cron task edit form with sections: basic info, schedule, execution, notification, retry" 生成骨架，然后接入 vee-validate + zod 校验逻辑。

---

### TICKET-TASK-LOG — `pages/tasklog/list.vue`
### TICKET-TEMPLATE-LIST — `pages/template/list.vue`
### TICKET-TEMPLATE-EDIT — `pages/template/edit.vue`
### TICKET-TASK-SIDEBAR — `pages/task/sidebar.vue`

格式同上。

---

## Phase 7 — 外壳（由 1 个 agent 完成，避免竞争）

### TICKET-SHELL — App / header / sidebar / navMenu / footer / notFound / LanguageSwitcher

**Status**: ⚪ TODO
**Complexity**: 中
**Branch**: `feat/shadcn-migration/shell`
**Depends**: Phase 3-6 完成

**独占文件**：
- `src/App.vue`
- `src/components/common/header.vue`
- `src/components/common/sidebar.vue`
- `src/components/common/navMenu.vue`
- `src/components/common/footer.vue`
- `src/components/common/notFound.vue`
- `src/components/common/LanguageSwitcher.vue`

**此 phase 顺便做**：
- 全局启用 dark mode toggle（header 加按钮）
- 布局改用 Tailwind grid / flex
- 主色 / 字体 / 间距统一刷

---

## Phase 8 — 卸载 Element Plus（必最后）

### TICKET-REMOVE-ELEMENT-PLUS

**Status**: ⚪ TODO
**Depends**: Phase 0-7 全部 ✅
**Branch**: `feat/shadcn-migration/phase-8-cleanup`

**操作**：
1. `grep -r "from 'element-plus'" src/` 确认为空
2. `grep -r "@element-plus/icons-vue" src/` 确认为空
3. `package.json` 移除：
   - `element-plus`
   - `@element-plus/icons-vue`
4. `main.js` 删掉：
   ```js
   // import ElementPlus from 'element-plus'
   // import 'element-plus/dist/index.css'
   // app.use(ElementPlus)
   ```
5. `tailwind.config.js`：
   ```diff
   - prefix: 'tw-',
   - corePlugins: { preflight: false },
   ```
6. 批量去 `tw-` 前缀：**谨慎**，用 VS Code 正则 `\btw-([a-z])` → `$1`，review 后 commit
7. 启用 preflight 后手工走全站，检查样式偏移
8. `useNotify.js` 里移除 ElMessage / ElMessageBox 分支，只留 sonner

**验证**：
- [ ] 全站所有页面样式正常
- [ ] bundle 减少 ~280KB gzip（Element Plus 卸载）
- [ ] `go test ./...` 后端零影响
- [ ] `pnpm build` + `pnpm dev` 都正常

**交付**：
```
chore(shadcn): remove element-plus and finalize migration
```

---

## 🎯 最终合入 master

所有 ticket ✅ 后：
```bash
git checkout master
git merge --no-ff feat/shadcn-migration
git push origin master
```

保留 feat/shadcn-migration 分支 1 个月以备回滚，确认生产无异常后再删除。

---

## 📊 Ticket 看板（更新这里）

```
Phase 0: ⚪
Phase 1: ⚪ × 5
Phase 2: ⚪ × 1
Phase 3: ⚪ × 5
Phase 4: ⚪ × 3
Phase 5: ⚪ × 9
Phase 6: ⚪ × 6
Phase 7: ⚪ × 1
Phase 8: ⚪ × 1

Total: 0/31 DONE
```

每次 ticket 完成，PR 合入后更新这里为 ✅ 并在状态下加一行：
```
✅ 2026-04-18 by @agent-name
```
