# gocron 前端迁移 shadcn-vue 方案

> 目标：把现有 Element Plus 前端平滑迁移到 shadcn-vue + Tailwind CSS，不影响任何现有功能，支持多 agent 并行开发。

---

## 1. 迁移原则

1. **功能等价** — 每页迁移后行为、API 调用、i18n、跳转逻辑必须和原版一字不差
2. **共存再替换** — Element Plus 和 shadcn-vue 同时存在，直到最后一 phase 才卸载 Element Plus
3. **可独立回滚** — 每个 ticket 一个独立 commit，坏了 `git revert` 精准撤回
4. **并行安全** — ticket 之间文件不重叠（或只追加不修改）
5. **不动后端** — 本次迁移**只改前端**，Go 服务一行不动

---

## 2. 技术栈决策

| 项 | 决策 | 说明 |
|---|------|------|
| CSS 引擎 | Tailwind CSS，**`prefix: "tw-"`，`preflight: false`** | 避免和 Element Plus 全局样式冲突 |
| 组件库 | shadcn-vue（copy 到 `src/components/ui/`） | 官方 Vue 版，copy-paste 哲学 |
| 图标 | `lucide-vue-next`（新）+ `@element-plus/icons-vue`（老，逐页替换） | 过渡期双轨 |
| 表格 | `@tanstack/vue-table` + 自封装 `<DataTable>` | 替代 `el-table` |
| 表单校验 | `vee-validate` + `zod` | 替代 `el-form` rules |
| Toast | `vue-sonner`，通过 `useNotify()` facade 抽象 | 过渡期 backend 可切换 |
| Dialog | shadcn `Dialog` / `AlertDialog` | 替代 `el-dialog` / `ElMessageBox` |
| i18n | `vue-i18n`（**不动**） | 和 UI 库解耦 |
| 状态管理 | `pinia`（**不动**） | 和 UI 库解耦 |
| HTTP | `axios` + `utils/httpClient`（**不动**） | 和 UI 库解耦 |

---

## 3. 分支策略

```
master                                    ← 稳定线，迁移过程不动
  │
  └─ feat/shadcn-migration                ← 集成分支，本次迁移的 base
       │
       ├─ feat/shadcn-migration/phase-0-foundation
       ├─ feat/shadcn-migration/phase-1-infra
       ├─ feat/shadcn-migration/login     ← 各 agent 的 ticket 分支
       ├─ feat/shadcn-migration/install
       ├─ feat/shadcn-migration/user-list
       └─ ...
```

每个 ticket 一个子分支，完成后 PR 合入 `feat/shadcn-migration`。全部完成后 `feat/shadcn-migration` 合入 master。

---

## 4. Phase 依赖图

```
Phase 0 (foundation)
    │
    ▼
Phase 1 (infra: notify facade, DataTable, ui primitives)
    │
    ├─ Phase 2 (CronPreview/HeatmapSvg 重写)  ← 可选先做
    │
    └─ Phase 3-7 (逐页迁移，可并行)
         ├─ Phase 3: login, install, editMyPassword, editPassword, twoFactor
         ├─ Phase 4: loginLog, auditLog, statistics
         ├─ Phase 5: user list/edit, host list/edit, notification, logRetention
         ├─ Phase 6: task list/edit/log, template list/edit, task/sidebar
         └─ Phase 7: App, header, sidebar, navMenu, footer, notFound
                │
                ▼
         Phase 8 (cleanup: 卸载 Element Plus)
```

**Phase 0 → 1 必须串行。Phase 2-7 内部各 ticket 大多可并行。Phase 8 必须最后。**

---

## 5. 文件所有权矩阵

**共享文件 = 所有 ticket 只能追加，不能修改现有内容：**

| 文件 | 规则 |
|------|------|
| `web/vue/src/locales/zh-CN.js` | **只追加**新 key，按 namespace 分组 |
| `web/vue/src/locales/en-US.js` | 同上 |
| `web/vue/src/router/index.js` | **只追加**新路由，不改现有 |
| `web/vue/src/App.vue` | Phase 7 才动 |
| `web/vue/src/main.js` | 只有 Phase 0、8 动 |
| `web/vue/package.json` | 只有 Phase 0、1、8 动 |
| `web/vue/vite.config.js` | Phase 0 动一次 |
| `web/vue/tailwind.config.js` | Phase 0 创建，后续原则上不动 |

**独占文件 = 每个 ticket 独占一或多个，不会和别的 ticket 冲突：**

| Ticket | 独占文件 |
|--------|---------|
| TICKET-LOGIN | `pages/user/login.vue` |
| TICKET-INSTALL | `pages/install/index.vue` |
| TICKET-USER-LIST | `pages/user/list.vue` |
| ... | ... |

所有权列表见 [AGENT_TASKS.md](./AGENT_TASKS.md)。

---

## 6. 共用组件清单（由 Phase 1 提供）

Ticket 实施时**优先用这些**，不要自己造轮子：

| 组件 / 工具 | 位置 | 用途 |
|------------|------|------|
| `Button`, `Input`, `Label`, `Card`, `Badge`, `Dialog`, `AlertDialog`, `Select`, `Popover`, `Tooltip`, `Tabs`, `Switch`, `Checkbox`, `RadioGroup`, `Form`, `Toaster`, `Separator`, `DropdownMenu` | `src/components/ui/*` | shadcn 基础组件 |
| `DataTable.vue` | `src/components/ui/data-table/` | TanStack Table + shadcn 封装，统一 table API |
| `useNotify()` | `src/composables/useNotify.js` | Toast + confirm 统一入口 |
| `useDarkMode()` | `src/composables/useDarkMode.js` | 暗色模式切换 |
| `cn()` | `src/lib/utils.js` | classname 合并工具（shadcn 标配） |
| `lucide-vue-next` | npm | 图标 |

---

## 7. 迁移模板（每个 ticket 都遵循）

### 步骤
1. **读老版本** — 完整 `Read` 老文件，列出：
   - data / methods / computed / watch / lifecycle
   - 所有调用的 API
   - 所有使用的 i18n key
   - 所有 form rules 和条件渲染
2. **写等价行为清单** — 用户流程级，不是代码级
3. **引用 block / example** — shadcn-vue Blocks 或官方 Examples 找相近结构
4. **实现新版** — 用 Phase 1 提供的共用组件
5. **本地测试** — 完整走完行为清单
6. **Commit** — 一个 ticket 一个 commit，消息格式：`feat(shadcn): migrate <page-name> to shadcn-vue`

### 文件结构约定（新页面）
```vue
<template>
  <!-- 使用 tw-* 前缀的 Tailwind class + shadcn 组件 -->
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { z } from 'zod'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'

// shadcn 组件
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

// 业务
import userApi from '@/api/user'
import { useNotify } from '@/composables/useNotify'

// Schema（替代 el-form rules）
const schema = toTypedSchema(z.object({
  username: z.string().min(1, '...'),
  password: z.string().min(1, '...'),
}))

const { handleSubmit, errors } = useForm({ validationSchema: schema })

// ... 业务逻辑
</script>
```

**注意**：优先用 Composition API `<script setup>`。老代码的 Options API 逻辑（data/methods）改写为 ref/function。

---

## 8. 测试策略

### 每个 ticket 的必测 matrix

| 维度 | 内容 |
|------|------|
| **用户流程** | 老版本的所有核心流程走一遍 |
| **响应式** | 375px (手机) / 768px (平板) / 1280px (桌面) |
| **语言** | zh-CN / en-US 切换无错位 |
| **暗色模式** | light / dark 切换正常 |
| **单元测试** | 如老代码有单测（`__tests__/`）保留并更新 |
| **Build** | `pnpm build` 无 error，bundle 增量 < 10KB/ticket |

### 可选项：Playwright e2e
Phase 4 之后，关键 table 页面建议加 Playwright 脚本防回归：
```bash
tests/e2e/shadcn/
  ├─ login.spec.ts
  ├─ task-list.spec.ts
  └─ audit-log.spec.ts
```

### 集成测试（由 QA Agent 定期跑）
每 3-5 个 ticket 合入 `feat/shadcn-migration` 后，跑一次全站回归：
- 所有页面能打开
- 核心 CRUD 流程（登录 / 建任务 / 改密 / 查日志）
- 后端 `go test ./...` 确认零影响

---

## 9. 回滚策略

### 单 ticket 回滚
每个 ticket 一个 commit，问题定位明确：
```bash
git revert <ticket-commit>
```

### Phase 级回滚
若整个 phase 有问题（很少见），整段 revert 到 phase 开始点。

### 全面回滚（最坏情况）
`feat/shadcn-migration` 分支整个丢弃，master 未动，零影响。

---

## 10. 并行协作流程

### Agent 协作步骤
```
1. Agent 从 AGENT_TASKS.md 挑一张未被认领的 ticket
2. 从 feat/shadcn-migration 切子分支：
   git checkout -b feat/shadcn-migration/<ticket-id> feat/shadcn-migration
3. 读老代码 + 写新版
4. 本地测试通过
5. Commit 并 push 子分支
6. 提 PR 到 feat/shadcn-migration
7. Reviewer（另一个 agent 或人）走完接收测试 matrix
8. 合入，标记 ticket done
```

### 冲突处理
- 独占文件不会冲突
- 共享文件（locales / router）按"先到先得"规则，后到的 agent 手动 rebase 解决追加冲突（通常 1 分钟内解决）

### 推荐并发数
- **2-3 个 dev agent 同时干活** + **1 个 review/QA agent**
- 不建议 4+，容易出现 phase 1 infra 没稳定前就乱套

---

## 11. 时间线估算

| Phase | 工作量 | 并行度 | 墙钟时间 |
|-------|-------|--------|---------|
| 0 | 1-2 天 | 1 | 1-2 天 |
| 1 | 3-5 天 | 1 | 3-5 天 |
| 2 | 0.5 天 | 1 | 0.5 天 |
| 3 | 5 tickets × 0.5-1 天 | 2-3 | 2-3 天 |
| 4 | 3 tickets × 1-2 天 | 2-3 | 2-3 天 |
| 5 | 9 tickets × 0.5-1 天 | 3 | 3-5 天 |
| 6 | 6 tickets × 1-3 天 | 2 | 5-8 天 |
| 7 | 6 tickets × 0.3 天 | 2 | 2 天 |
| 8 | 3-5 天 | 1 | 3-5 天 |
| **合计** | | | **~4 周** |

相比单 agent 串行（10-12 周），并行 3 个 agent 压到 **4 周左右**。

---

## 12. 禁止事项

- ❌ 不要同时改两个独占文件（违反 ticket 粒度）
- ❌ 不要跨 phase 做改动（phase 3 ticket 顺带改 phase 5 文件）
- ❌ 不要修改 i18n 文件里**已有的** key（只能追加）
- ❌ 不要动后端代码
- ❌ 不要引入新的 UI 库（除了本文档列出的）
- ❌ 不要手写 CSS（用 Tailwind utility class + shadcn 组件 variant）
- ❌ 不要在这个 phase 做"顺手的其他重构"（保持迁移纯粹）
- ❌ 不要修改 ElMessage / ElMessageBox 的直接调用（必须走 `useNotify()` facade）

---

## 13. 开工后第一件事：Phase 0

参见 [AGENT_TASKS.md](./AGENT_TASKS.md) TICKET-PHASE-0。
