<script setup>
import { h, ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter, useRoute } from 'vue-router'
import {
  Plus,
  RefreshCw,
  Download,
  Pencil,
  Wifi,
  Trash2,
  ClipboardCopy,
  Monitor,
  Loader2
} from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger
} from '@/components/ui/tabs'
import { DataTable } from '@/components/ui/data-table'
import { useNotify } from '@/composables/useNotify'

import hostService from '@/api/host'
import agentService from '@/api/agent'
import { useUserStore } from '@/stores/user'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const notify = useNotify()
const userStore = useUserStore()

// ── Auth ──────────────────────────────────────────────────────────────────────
const isAdmin = computed(() => userStore.isAdmin)

// ── List state ────────────────────────────────────────────────────────────────
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

// ── Search filters ────────────────────────────────────────────────────────────
const filterId = ref('')
const filterName = ref('')

// ── Auto-register / Agent install dialog ──────────────────────────────────────
const agentDialogOpen = ref(false)
const agentLoading = ref(false)
const installCommand = ref('')
const expiresAt = ref('')
const activeTab = ref('linux')

// Module-level token cache (persists while page component is alive)
let cachedToken = null
let cachedTokenExpires = null

// ── Table columns ─────────────────────────────────────────────────────────────
const columns = computed(() => {
  const cols = [
    {
      accessorKey: 'id',
      header: 'ID',
      size: 70
    },
    {
      accessorKey: 'alias',
      header: t('host.alias')
    },
    {
      accessorKey: 'name',
      header: t('host.name')
    },
    {
      accessorKey: 'port',
      header: t('host.port'),
      size: 90
    },
    {
      id: 'tasks',
      header: t('task.viewLog'),
      size: 120,
      cell: ({ row }) =>
        h(
          Button,
          {
            variant: 'outline',
            size: 'sm',
            class: 'tw-h-7 tw-px-3',
            onClick: () => toTasks(row.original)
          },
          () => t('task.list')
        )
    },
    {
      accessorKey: 'remark',
      header: t('host.remark')
    }
  ]

  if (isAdmin.value) {
    cols.push({
      id: 'actions',
      header: t('common.operation'),
      size: 210,
      cell: ({ row }) => {
        const item = row.original
        return h('div', { class: 'tw-flex tw-items-center tw-gap-1' }, [
          h(
            Button,
            {
              variant: 'outline',
              size: 'sm',
              class: 'tw-h-7 tw-px-2',
              onClick: () => toEdit(item)
            },
            () => [h(Pencil, { class: 'tw-size-3.5 tw-mr-1' }), t('common.edit')]
          ),
          h(
            Button,
            {
              variant: 'outline',
              size: 'sm',
              class: 'tw-h-7 tw-px-2',
              onClick: () => pingHost(item)
            },
            () => [h(Wifi, { class: 'tw-size-3.5 tw-mr-1' }), t('system.testSend')]
          ),
          h(
            Button,
            {
              variant: 'ghost',
              size: 'sm',
              class: 'tw-h-7 tw-px-2 tw-text-destructive hover:tw-text-destructive',
              onClick: () => removeHost(item)
            },
            () => [h(Trash2, { class: 'tw-size-3.5 tw-mr-1' }), t('common.delete')]
          )
        ])
      }
    })
  }

  return cols
})

// ── Data fetching ─────────────────────────────────────────────────────────────
function loadData() {
  loading.value = true
  const params = {
    page: page.value,
    page_size: pageSize.value,
    id: filterId.value,
    name: filterName.value
  }
  hostService.list(params, data => {
    list.value = data.data ?? []
    total.value = data.total ?? 0
    loading.value = false
  })
}

function onPageChange(newPage) {
  page.value = newPage
  loadData()
}

function onPageSizeChange(newSize) {
  pageSize.value = newSize
  page.value = 1
  loadData()
}

function search() {
  page.value = 1
  loadData()
}

function refresh() {
  loadData()
  notify.success(t('message.refreshSuccess'))
}

// ── Row actions ───────────────────────────────────────────────────────────────
function toEdit(item) {
  const path = item === null ? '/host/create' : `/host/edit/${item.id}`
  router.push(path)
}

function toTasks(item) {
  router.push({ path: '/task', query: { host_id: item.id } })
}

function pingHost(item) {
  if (!item.id || item.id <= 0) {
    notify.error(t('message.dataNotFound'))
    return
  }
  hostService.ping(item.id, () => {
    notify.success(t('message.connectionSuccess'))
  })
}

async function removeHost(item) {
  const ok = await notify.confirm(t('message.confirmDeleteNode'), t('common.tip'))
  if (!ok) return
  hostService.remove(item.id, () => {
    loadData()
  })
}

// ── Auto-register: show dialog, fetch / reuse token ──────────────────────────
function showAgentInstall() {
  agentDialogOpen.value = true

  const now = new Date()
  if (cachedToken && cachedTokenExpires && now < cachedTokenExpires) {
    installCommand.value = cachedToken.install_cmd
    expiresAt.value = cachedTokenExpires.toLocaleString()
    return
  }

  installCommand.value = ''
  expiresAt.value = ''
  agentLoading.value = true

  agentService.generateToken(data => {
    installCommand.value = data.install_cmd
    const expiresDate = new Date(data.expires_at)
    expiresAt.value = expiresDate.toLocaleString()
    cachedToken = data
    cachedTokenExpires = expiresDate
    agentLoading.value = false
  })
}

// ── Copy command with execCommand fallback ────────────────────────────────────
function copyCommand() {
  const cmd = installCommand.value
  if (!cmd) return

  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard
      .writeText(cmd)
      .then(() => notify.success(t('message.copySuccess')))
      .catch(() => fallbackCopy(cmd))
  } else {
    fallbackCopy(cmd)
  }
}

function fallbackCopy(text) {
  const ta = document.createElement('textarea')
  ta.value = text
  ta.style.position = 'fixed'
  ta.style.opacity = '0'
  document.body.appendChild(ta)
  ta.focus()
  ta.select()
  try {
    const ok = document.execCommand('copy')
    ok ? notify.success(t('message.copySuccess')) : notify.error(t('message.copyFailed'))
  } catch {
    notify.error(t('message.copyFailed'))
  }
  document.body.removeChild(ta)
}

// ── Route watcher: reload list after returning from create/edit ───────────────
watch(
  () => route.path,
  (to, from) => {
    if (
      to === '/host' &&
      from &&
      (from === '/host/create' || from.startsWith('/host/edit/'))
    ) {
      loadData()
    }
  }
)

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="tw-p-6 tw-space-y-4">
    <Card>
      <CardHeader class="tw-pb-3">
        <div class="tw-flex tw-items-center tw-justify-between tw-flex-wrap tw-gap-3">
          <CardTitle>{{ t('host.list') }}</CardTitle>

          <div class="tw-flex tw-items-center tw-gap-2 tw-flex-wrap">
            <!-- Search filters -->
            <Input
              v-model.trim="filterId"
              class="tw-h-8 tw-w-20"
              placeholder="ID"
              @keyup.enter="search"
            />
            <Input
              v-model.trim="filterName"
              class="tw-h-8 tw-w-36"
              :placeholder="t('host.name')"
              @keyup.enter="search"
            />
            <Button
              size="sm"
              class="tw-h-8"
              @click="search"
            >
              {{ t('common.search') }}
            </Button>

            <div class="tw-h-4 tw-w-px tw-bg-border" />

            <!-- Auto register (admin only) -->
            <Button
              v-if="isAdmin"
              variant="outline"
              size="sm"
              class="tw-h-8"
              @click="showAgentInstall"
            >
              <Download class="tw-size-4 tw-mr-1.5" />
              {{ t('host.autoRegister') }}
            </Button>

            <!-- Add (admin only) -->
            <Button
              v-if="isAdmin"
              size="sm"
              class="tw-h-8"
              @click="toEdit(null)"
            >
              <Plus class="tw-size-4 tw-mr-1.5" />
              {{ t('common.add') }}
            </Button>

            <!-- Refresh -->
            <Button
              variant="outline"
              size="sm"
              class="tw-h-8"
              :disabled="loading"
              @click="refresh"
            >
              <RefreshCw
                class="tw-size-4 tw-mr-1.5"
                :class="{ 'tw-animate-spin': loading }"
              />
              {{ t('common.refresh') }}
            </Button>
          </div>
        </div>
      </CardHeader>

      <CardContent class="tw-pt-0">
        <DataTable
          :columns="columns"
          :data="list"
          :loading="loading"
          :total="total"
          :page="page"
          :page-size="pageSize"
          @update:page="onPageChange"
          @update:page-size="onPageSizeChange"
        />
      </CardContent>
    </Card>

    <!-- ── Agent install dialog ──────────────────────────────────────────────── -->
    <Dialog v-model:open="agentDialogOpen">
      <DialogContent class="tw-max-w-2xl">
        <DialogHeader>
          <DialogTitle>{{ t('host.agentInstall') }}</DialogTitle>
        </DialogHeader>

        <!-- Loading spinner -->
        <div
          v-if="agentLoading"
          class="tw-flex tw-flex-col tw-items-center tw-justify-center tw-py-10 tw-gap-3"
        >
          <Loader2 class="tw-size-8 tw-animate-spin tw-text-muted-foreground" />
          <p class="tw-text-sm tw-text-muted-foreground">
            {{ t('common.loading') }}
          </p>
        </div>

        <!-- Content once token is available -->
        <div
          v-else-if="installCommand"
          class="tw-space-y-4"
        >
          <!-- Info banner -->
          <div
            class="tw-flex tw-gap-2 tw-rounded-md tw-border tw-border-blue-200 tw-bg-blue-50 tw-p-3 tw-text-sm tw-text-blue-800"
          >
            <Monitor class="tw-size-4 tw-mt-0.5 tw-shrink-0" />
            <span>{{ t('host.installTip') }}</span>
          </div>

          <!-- OS tabs -->
          <Tabs
            v-model="activeTab"
            class="tw-w-full"
          >
            <TabsList class="tw-grid tw-w-full tw-grid-cols-2">
              <TabsTrigger value="linux">Linux / macOS</TabsTrigger>
              <TabsTrigger value="windows">Windows</TabsTrigger>
            </TabsList>

            <!-- Linux / macOS -->
            <TabsContent
              value="linux"
              class="tw-mt-3"
            >
              <div class="tw-rounded-md tw-bg-muted tw-p-4 tw-space-y-3">
                <p class="tw-text-sm tw-text-muted-foreground">
                  {{ t('host.bashCommand') }}
                </p>
                <pre
                  class="tw-rounded tw-bg-background tw-border tw-p-3 tw-text-xs tw-font-mono tw-whitespace-pre-wrap tw-break-all"
                >{{ installCommand }}</pre>
                <div class="tw-flex tw-justify-end">
                  <Button
                    size="sm"
                    @click="copyCommand"
                  >
                    <ClipboardCopy class="tw-size-4 tw-mr-1.5" />
                    Copy
                  </Button>
                </div>
              </div>
            </TabsContent>

            <!-- Windows -->
            <TabsContent
              value="windows"
              class="tw-mt-3"
            >
              <div class="tw-space-y-3">
                <!-- Warning banner -->
                <div
                  class="tw-rounded-md tw-border tw-border-yellow-200 tw-bg-yellow-50 tw-p-3 tw-text-sm tw-text-yellow-800"
                >
                  <p class="tw-font-semibold tw-mb-1">
                    {{ t('host.windowsManualInstall') }}
                  </p>
                  <p>{{ t('host.windowsManualInstallTip') }}</p>
                </div>

                <!-- Numbered steps -->
                <ol class="tw-space-y-3 tw-pl-1">
                  <li
                    v-for="(step, idx) in [
                      { title: t('host.windowsStep1'), desc: t('host.windowsStep1Desc') },
                      { title: t('host.windowsStep2'), desc: t('host.windowsStep2Desc') },
                      { title: t('host.windowsStep3'), desc: t('host.windowsStep3Desc') }
                    ]"
                    :key="idx"
                    class="tw-flex tw-gap-3"
                  >
                    <span
                      class="tw-flex tw-size-6 tw-shrink-0 tw-items-center tw-justify-center tw-rounded-full tw-bg-primary tw-text-xs tw-font-bold tw-text-primary-foreground"
                    >
                      {{ idx + 1 }}
                    </span>
                    <div>
                      <p class="tw-text-sm tw-font-medium">{{ step.title }}</p>
                      <p class="tw-text-sm tw-text-muted-foreground">{{ step.desc }}</p>
                    </div>
                  </li>
                </ol>
              </div>
            </TabsContent>
          </Tabs>

          <!-- Token metadata -->
          <div class="tw-rounded-md tw-border tw-divide-y">
            <div class="tw-flex tw-items-center tw-gap-3 tw-px-4 tw-py-2.5 tw-text-sm">
              <span class="tw-w-32 tw-shrink-0 tw-font-medium">{{ t('host.tokenExpires') }}</span>
              <Badge
                variant="outline"
                class="tw-font-mono"
              >
                {{ expiresAt }}
              </Badge>
            </div>
            <div class="tw-flex tw-items-start tw-gap-3 tw-px-4 tw-py-2.5 tw-text-sm">
              <span class="tw-w-32 tw-shrink-0 tw-font-medium">{{ t('host.tokenUsage') }}</span>
              <span class="tw-text-green-600">{{ t('host.tokenReusable') }}</span>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
