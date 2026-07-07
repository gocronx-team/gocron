<!-- AI 运维助手聊天面板 -->
<template>
  <ArtIconButton icon="ri:robot-2-line" :title="t('aiChat.title')" @click="openDrawer" />

  <!-- 可拖拽改宽手柄：骑在抽屉左边界上，Teleport 到 body 避免被父级 stacking context 限制 -->
  <Teleport to="body">
    <div
      v-if="visible"
      class="ai-chat-resizer"
      :style="{ right: drawerWidth + 'px' }"
      @mousedown.prevent.stop="startResize"
    ></div>
  </Teleport>

  <ElDrawer
    v-model="visible"
    :title="t('aiChat.title')"
    direction="rtl"
    :size="drawerWidth"
    class="ai-chat-drawer"
    @close="cancelStream"
  >
    <template #header>
      <div class="ai-chat-header">
        <span class="ai-chat-title">{{ t('aiChat.title') }}</span>
        <ElButton link :disabled="messages.length === 0 || loading" @click="clearConversation">
          {{ t('aiChat.clear') }}
        </ElButton>
      </div>
    </template>

    <div class="ai-chat-body">
      <div ref="listRef" class="ai-chat-list">
        <p v-if="messages.length === 0" class="ai-chat-empty">{{ t('aiChat.empty') }}</p>

        <div
          v-for="(msg, index) in messages"
          :key="index"
          class="ai-chat-row"
          :class="msg.role === 'user' ? 'is-user' : 'is-assistant'"
        >
          <div class="ai-chat-bubble">
            <!-- 思考过程（思考型模型）：可折叠，默认展开，让用户在长时间推理时看到模型在动 -->
            <details
              v-if="msg.role === 'assistant' && reasoningByIndex[index]"
              class="ai-chat-reasoning"
              open
            >
              <summary>{{ t('aiChat.reasoning') }}</summary>
              <div class="ai-chat-reasoning-body">{{ reasoningByIndex[index] }}</div>
            </details>
            <!-- 助手消息渲染 Markdown（renderMarkdown 已做 HTML 转义 + 白名单，XSS 安全）；用户消息保持纯文本 -->
            <div
              v-if="msg.content && msg.role === 'assistant'"
              class="ai-chat-md"
              v-html="renderMarkdown(msg.content)"
            ></div>
            <template v-else-if="msg.content">{{ msg.content }}</template>
            <span
              v-else-if="msg.role === 'assistant' && loading && !reasoningByIndex[index]"
              class="ai-chat-thinking"
            >
              {{ t('aiChat.thinking') }}
            </span>
          </div>
          <ElButton
            v-if="msg.content"
            link
            size="small"
            class="ai-chat-copy"
            @click="copyMessage(msg.content)"
          >
            {{ t('aiChat.copy') }}
          </ElButton>
          <div v-if="msg.role === 'assistant' && toolsByIndex[index]?.length" class="ai-chat-tools">
            <ElTag
              v-for="tool in toolsByIndex[index]"
              :key="tool.id"
              size="small"
              :type="
                tool.status === 'failed' ? 'danger' : tool.status === 'done' ? 'success' : 'info'
              "
            >
              {{ toolLabel(tool) }}
            </ElTag>
          </div>

          <!-- run_task 确认:模型不会自动执行任务,必须用户在此点确认 -->
          <div
            v-if="msg.role === 'assistant' && pendingRunByIndex[index]"
            class="ai-chat-run-confirm"
          >
            <span class="ai-chat-run-label">
              {{ t('aiChat.runConfirm', { name: pendingRunByIndex[index].taskName }) }}
            </span>
            <template v-if="pendingRunByIndex[index].status === 'pending'">
              <ElButton type="primary" size="small" @click="confirmRun(index)">
                {{ t('aiChat.run') }}
              </ElButton>
              <ElButton size="small" @click="cancelRun(index)">{{ t('aiChat.cancel') }}</ElButton>
            </template>
            <ElTag
              v-else
              size="small"
              :type="pendingRunByIndex[index].status === 'done' ? 'success' : 'info'"
            >
              {{ runStatusLabel(pendingRunByIndex[index].status) }}
            </ElTag>
          </div>
        </div>
      </div>

      <div class="ai-chat-input">
        <ElInput
          v-model="draft"
          type="textarea"
          :rows="3"
          resize="none"
          :placeholder="t('aiChat.placeholder')"
          @keydown.enter="onEnter"
        />
        <ElButton v-if="loading" type="danger" @click="cancelStream">
          {{ t('aiChat.stop') }}
        </ElButton>
        <ElButton v-else type="primary" :disabled="!draft.trim()" @click="send">
          {{ t('aiChat.send') }}
        </ElButton>
      </div>
    </div>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { ElButton, ElDrawer, ElInput, ElMessage, ElTag } from 'element-plus'
  import { streamAiChat, confirmRunTask, type AiChatMessage } from '@/api/ai'
  import { copyToClipboard } from '@/utils/clipboard'
  import { renderMarkdown } from '@/utils/markdown'

  defineOptions({ name: 'ArtAiChat' })

  const { t } = useI18n()

  // 单条消息内展示的工具调用 chip：running → done/failed。
  type ToolChip = { id: string; name: string; status: 'running' | 'done' | 'failed' }

  const visible = ref(false)
  const loading = ref(false)
  const draft = ref('')
  // 抽屉宽度（px），可由左边界手柄拖拽调整。
  const drawerWidth = ref(420)
  const MIN_WIDTH = 360
  const messages = ref<AiChatMessage[]>([])
  // 工具列表按 messages 下标存放，仅用于展示，不回传给后端。
  const toolsByIndex = ref<Record<number, ToolChip[]>>({})
  // 思考过程按 messages 下标存放，仅用于展示，不回传给后端。
  const reasoningByIndex = ref<Record<number, string>>({})
  // run_task 确认请求：模型想执行任务时，需用户在此确认后才真正执行。
  type PendingRun = {
    taskId: number
    taskName: string
    status: 'pending' | 'done' | 'failed' | 'cancelled'
  }
  const pendingRunByIndex = ref<Record<number, PendingRun>>({})
  const listRef = ref<HTMLElement>()
  // 进行中流的取消句柄，清空/关闭时用来中断，避免泄漏。
  let controller: AbortController | null = null

  const openDrawer = (): void => {
    visible.value = true
  }

  // 拖拽改宽：rtl 抽屉锚定右侧，宽度 = 视口宽 - 鼠标 X，限制在 [MIN_WIDTH, 92vw 且 ≤1000]。
  const doResize = (e: MouseEvent): void => {
    const max = Math.min(Math.round(window.innerWidth * 0.92), 1000)
    drawerWidth.value = Math.max(MIN_WIDTH, Math.min(window.innerWidth - e.clientX, max))
  }
  const stopResize = (): void => {
    document.removeEventListener('mousemove', doResize)
    document.removeEventListener('mouseup', stopResize)
    document.body.style.userSelect = ''
  }
  const startResize = (): void => {
    document.addEventListener('mousemove', doResize)
    document.addEventListener('mouseup', stopResize)
    document.body.style.userSelect = 'none'
  }

  const toolLabel = (tool: ToolChip): string => {
    if (tool.status === 'failed') return t('aiChat.toolFailed', { name: tool.name })
    if (tool.status === 'done') return t('aiChat.toolDone', { name: tool.name })
    return t('aiChat.calling', { name: tool.name })
  }

  const scrollToBottom = (): void => {
    nextTick(() => {
      if (listRef.value) listRef.value.scrollTop = listRef.value.scrollHeight
    })
  }

  /**
   * 发送消息：推入用户消息 + 一条空的助手消息（流式写入目标），带全量历史
   * 调用流式接口。各事件回调直接改写最后一条助手消息的 content / 工具 chip，
   * 触发响应式更新实现逐字渲染。
   */
  const send = (): void => {
    const content = draft.value.trim()
    if (!content || loading.value) return

    messages.value = [
      ...messages.value,
      { role: 'user', content },
      { role: 'assistant', content: '' }
    ]
    const assistantIndex = messages.value.length - 1
    // 只把真正的对话历史（不含刚插入的空助手占位）传给后端。
    const history = messages.value.slice(0, assistantIndex) as AiChatMessage[]
    draft.value = ''
    loading.value = true
    scrollToBottom()

    controller = new AbortController()
    streamAiChat(
      history,
      {
        onMessage: (delta) => {
          const target = messages.value[assistantIndex]
          if (target) target.content += delta
          scrollToBottom()
        },
        onReasoning: (delta) => {
          // 思考型模型的推理过程，单独累积、单独展示，不混入答案。
          reasoningByIndex.value = {
            ...reasoningByIndex.value,
            [assistantIndex]: (reasoningByIndex.value[assistantIndex] ?? '') + delta
          }
          scrollToBottom()
        },
        onToolCall: (tcall) => {
          const list = toolsByIndex.value[assistantIndex] ?? []
          toolsByIndex.value = {
            ...toolsByIndex.value,
            [assistantIndex]: [...list, { id: tcall.id, name: tcall.name, status: 'running' }]
          }
          scrollToBottom()
        },
        onToolResult: (tres) => {
          const list = toolsByIndex.value[assistantIndex] ?? []
          toolsByIndex.value = {
            ...toolsByIndex.value,
            [assistantIndex]: list.map((chip) =>
              chip.id === tres.id ? { ...chip, status: tres.ok ? 'done' : 'failed' } : chip
            )
          }
        },
        onConfirmRequired: (req) => {
          // 模型想执行任务：不自动执行，挂一个待用户确认的请求。
          pendingRunByIndex.value = {
            ...pendingRunByIndex.value,
            [assistantIndex]: { taskId: req.taskId, taskName: req.taskName, status: 'pending' }
          }
          scrollToBottom()
        },
        onError: (msg) => {
          ElMessage.error(msg)
          const target = messages.value[assistantIndex]
          if (target && !target.content) target.content = t('aiChat.failed')
          loading.value = false
        },
        onDone: () => {
          loading.value = false
          controller = null
          scrollToBottom()
        }
      },
      controller.signal
    )
  }

  const cancelStream = (): void => {
    if (controller) {
      controller.abort()
      controller = null
    }
    loading.value = false
  }

  /**
   * Enter 发送，Shift+Enter 换行。
   */
  const onEnter = (e: Event | KeyboardEvent): void => {
    if ((e as KeyboardEvent).shiftKey) return
    e.preventDefault()
    send()
  }

  const copyMessage = async (content: string): Promise<void> => {
    if (await copyToClipboard(content)) {
      ElMessage.success(t('aiChat.copied'))
    } else {
      ElMessage.error(t('aiChat.copyFailed'))
    }
  }

  const setRunStatus = (index: number, status: PendingRun['status']): void => {
    const cur = pendingRunByIndex.value[index]
    if (cur) pendingRunByIndex.value = { ...pendingRunByIndex.value, [index]: { ...cur, status } }
  }

  const confirmRun = async (index: number): Promise<void> => {
    const run = pendingRunByIndex.value[index]
    if (!run || run.status !== 'pending') return
    try {
      await confirmRunTask(run.taskId)
      setRunStatus(index, 'done')
      ElMessage.success(t('aiChat.runStarted', { name: run.taskName }))
    } catch {
      // 错误已由 http 拦截器提示
      setRunStatus(index, 'failed')
    }
  }

  const cancelRun = (index: number): void => {
    setRunStatus(index, 'cancelled')
  }

  const runStatusLabel = (s: PendingRun['status']): string => {
    if (s === 'done') return t('aiChat.runDone')
    if (s === 'failed') return t('aiChat.runFailed')
    return t('aiChat.runCancelled')
  }

  const clearConversation = (): void => {
    cancelStream()
    messages.value = []
    toolsByIndex.value = {}
    reasoningByIndex.value = {}
    pendingRunByIndex.value = {}
  }

  onBeforeUnmount(() => {
    cancelStream()
    stopResize()
  })
</script>

<style scoped>
  @reference '@styles/core/tailwind.css';

  .ai-chat-resizer {
    position: fixed;
    top: 0;
    bottom: 0;
    z-index: 2050; /* 高于 el-drawer 面板，便于拖拽 */
    width: 6px;
    margin-right: -3px; /* 居中骑在抽屉左边界上 */
    cursor: ew-resize;
    transition: background-color 0.15s;
  }

  .ai-chat-resizer:hover {
    background-color: var(--main-color);
  }

  .ai-chat-header {
    @apply flex items-center justify-between w-full pr-6;
  }

  .ai-chat-title {
    @apply text-base font-medium;
  }

  .ai-chat-body {
    @apply flex flex-col h-full;
  }

  .ai-chat-list {
    @apply flex-1 overflow-y-auto pr-1;
  }

  .ai-chat-empty {
    @apply mt-10 text-sm text-center text-g-500 select-none;
  }

  .ai-chat-row {
    @apply flex flex-col mb-3;
  }

  .ai-chat-row.is-user {
    @apply items-end;
  }

  .ai-chat-row.is-assistant {
    @apply items-start;
  }

  .ai-chat-bubble {
    @apply max-w-[85%] px-3 py-2 text-sm rounded-lg;

    word-break: break-word;
    white-space: pre-wrap;
  }

  .is-user .ai-chat-bubble {
    color: #fff;
    background-color: var(--main-color);
  }

  .is-assistant .ai-chat-bubble {
    background-color: var(--art-gray-200);
  }

  .ai-chat-thinking {
    @apply text-g-500;
  }

  .ai-chat-reasoning {
    @apply mb-2 text-xs text-g-500;
  }

  .ai-chat-reasoning summary {
    @apply cursor-pointer select-none;
  }

  .ai-chat-reasoning-body {
    @apply mt-1 pl-2 max-h-40 overflow-y-auto border-l-2 border-g-300;

    word-break: break-word;
    white-space: pre-wrap;
  }

  .ai-chat-copy {
    @apply h-auto p-0 mt-0.5 text-xs text-g-500;
  }

  .ai-chat-tools {
    @apply flex flex-wrap items-center gap-1 mt-1;
  }

  .ai-chat-run-confirm {
    @apply flex flex-wrap items-center gap-2 mt-2 p-2 rounded-md;

    background-color: var(--art-gray-100);
    border: 1px solid var(--art-gray-300);
  }

  .ai-chat-run-label {
    @apply text-xs text-g-700;
  }

  .ai-chat-tools-label {
    @apply text-xs text-g-500;
  }

  .ai-chat-input {
    @apply flex flex-col gap-2 pt-3 mt-2 border-t border-g-300/80;
  }
</style>

<!--
  Markdown 渲染内容通过 v-html 注入，拿不到 scoped 的 data-v- 属性，因此这些样式
  放在非 scoped 块里，用 .ai-chat-md 类名做命名空间（避免全局污染）。这样无需 :deep()，
  也不会触发 Tailwind(lightningcss) 对 :deep() 伪类的解析告警。
-->
<style>
  .ai-chat-md {
    white-space: normal;
  }

  .ai-chat-md p {
    margin: 0 0 8px;
  }

  .ai-chat-md p:last-child {
    margin-bottom: 0;
  }

  .ai-chat-md :is(h1, h2, h3, h4, h5, h6) {
    margin: 10px 0 6px;
    font-size: 14px;
    font-weight: 600;
  }

  .ai-chat-md :is(ul, ol) {
    padding-left: 20px;
    margin: 4px 0;
  }

  .ai-chat-md li {
    margin: 2px 0;
  }

  .ai-chat-md code {
    padding: 1px 4px;
    font-family: monospace;
    font-size: 12px;
    background: var(--art-gray-300);
    border-radius: 3px;
  }

  .ai-chat-md pre {
    padding: 8px 10px;
    margin: 6px 0;
    overflow-x: auto;
    background: var(--art-gray-300);
    border-radius: 6px;
  }

  .ai-chat-md pre code {
    padding: 0;
    background: none;
  }

  .ai-chat-md table {
    margin: 6px 0;
    font-size: 12px;
    border-collapse: collapse;
  }

  .ai-chat-md :is(th, td) {
    padding: 4px 8px;
    border: 1px solid var(--art-gray-400);
  }

  .ai-chat-md a {
    color: var(--main-color);
    text-decoration: underline;
  }

  .ai-chat-md blockquote {
    padding-left: 10px;
    margin: 6px 0;
    color: var(--art-gray-600);
    border-left: 3px solid var(--art-gray-400);
  }

  .ai-chat-md hr {
    margin: 8px 0;
    border: none;
    border-top: 1px solid var(--art-gray-400);
  }
</style>
