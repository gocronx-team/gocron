<!-- Task Create / Edit form -->
<!-- Routes: /task/create  or  /task/edit/:id -->
<template>
  <div class="task-edit-page">
    <ElCard shadow="never">
      <template #header>
        <span class="text-base font-medium">
          {{ isEdit ? t('task.editTitle') : t('task.createTitle') }}
        </span>
      </template>

      <ElForm
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="130px"
        @submit.prevent
      >
        <!-- ── Basic Info ──────────────────────────────────────────────── -->
        <ElCard shadow="never" class="section-card mb-4">
          <ElRow :gutter="24">
            <!-- name -->
            <ElCol :span="12">
              <ElFormItem :label="t('task.name')" prop="name">
                <ElInput
                  v-model.trim="form.name"
                  :placeholder="t('task.namePlaceholder')"
                  clearable
                />
              </ElFormItem>
            </ElCol>

            <!-- tag -->
            <ElCol :span="12">
              <ElFormItem :label="t('task.tag')">
                <ElSelect
                  v-model="form.tags"
                  multiple
                  filterable
                  allow-create
                  default-first-option
                  collapse-tags
                  collapse-tags-tooltip
                  :placeholder="t('task.tagPlaceholder')"
                  style="width: 100%"
                >
                  <ElOption
                    v-for="tag in tagOptions"
                    :key="tag"
                    :label="tag"
                    :value="tag"
                  />
                </ElSelect>
              </ElFormItem>
            </ElCol>
          </ElRow>

          <!-- description / remark -->
          <ElRow :gutter="24">
            <ElCol :span="18">
              <ElFormItem :label="t('task.description')">
                <ElInput
                  v-model="form.remark"
                  type="textarea"
                  :rows="2"
                  :placeholder="t('task.descriptionPlaceholder')"
                />
              </ElFormItem>
            </ElCol>
          </ElRow>
        </ElCard>

        <!-- ── Schedule ───────────────────────────────────────────────── -->
        <ElCard shadow="never" class="section-card mb-4">
          <template #header>
            <span class="section-title">{{ t('task.spec') }}</span>
          </template>

          <ElRow :gutter="24">
            <!-- level -->
            <ElCol :span="8">
              <ElFormItem :label="t('task.levelLabel')">
                <ElSelect v-model="form.level" :disabled="isEdit" style="width: 100%">
                  <ElOption :label="t('task.levelMaster')" :value="1" />
                  <ElOption :label="t('task.levelChild')" :value="2" />
                </ElSelect>
              </ElFormItem>
            </ElCol>

            <!-- cron spec — only for master tasks -->
            <ElCol :span="16" v-if="form.level === 1">
              <ElFormItem :label="t('task.spec')" prop="spec">
                <ElInput
                  v-model.trim="form.spec"
                  :placeholder="t('template.cronPlaceholder')"
                  clearable
                  style="font-family: monospace"
                  @change="previewCron"
                  @blur="previewCron"
                >
                  <template #append>
                    <ElPopover placement="bottom-end" :width="460" trigger="click">
                      <template #reference>
                        <ElButton>{{ t('template.cronExample') }}</ElButton>
                      </template>
                      <div class="cron-help">
                        <h4>{{ t('template.cronStandard') }}</h4>
                        <ul>
                          <li><code>0 * * * *</code> <span>{{ t('template.cronEveryMinute') }}</span></li>
                          <li><code>*/20 * * * *</code> <span>{{ t('template.cronEvery20Sec') }}</span></li>
                          <li><code>30 21 * * *</code> <span>{{ t('template.cronEveryDay2130') }}</span></li>
                          <li><code>0 23 * * 6</code> <span>{{ t('template.cronEverySat23') }}</span></li>
                        </ul>
                        <h4>{{ t('template.cronShortcut') }}</h4>
                        <ul>
                          <li><code>@reboot</code> <span>{{ t('template.cronReboot') }}</span></li>
                          <li><code>@yearly</code> <span>{{ t('template.cronYearly') }}</span></li>
                          <li><code>@monthly</code> <span>{{ t('template.cronMonthly') }}</span></li>
                          <li><code>@weekly</code> <span>{{ t('template.cronWeekly') }}</span></li>
                          <li><code>@daily</code> <span>{{ t('template.cronDaily') }}</span></li>
                          <li><code>@hourly</code> <span>{{ t('template.cronHourly') }}</span></li>
                          <li><code>@every 30s</code> <span>{{ t('template.cronEvery30s') }}</span></li>
                          <li><code>@every 1m20s</code> <span>{{ t('template.cronEvery1m20s') }}</span></li>
                        </ul>
                      </div>
                    </ElPopover>
                  </template>
                </ElInput>
              </ElFormItem>
            </ElCol>
          </ElRow>

          <!-- cron preview (rich panel: empty / error / next-runs) -->
          <ElRow v-if="form.level === 1" :gutter="24">
            <ElCol :span="20">
              <ElFormItem label=" ">
                <div
                  class="cron-preview"
                  :class="{ 'is-invalid': !!previewError, 'is-empty': !form.spec.trim() }"
                >
                  <div v-if="!form.spec.trim()" class="preview-state muted">
                    <ElIcon><InfoFilled /></ElIcon>
                    <span>{{ t('template.previewWaiting') }}</span>
                  </div>
                  <div v-else-if="previewError" class="preview-state error">
                    <ElIcon><WarningFilled /></ElIcon>
                    <span>{{ previewError }}</span>
                  </div>
                  <div v-else-if="nextRuns.length === 0" class="preview-state muted">
                    <ElIcon><InfoFilled /></ElIcon>
                    <span>{{ t('template.previewNoRuns') }}</span>
                  </div>
                  <div v-else>
                    <div class="preview-title">
                      <ElIcon><Clock /></ElIcon>
                      <span>{{ t('template.previewNextRuns', { count: nextRuns.length }) }}</span>
                      <ElTag v-if="previewTz" size="small" type="info" class="tz-tag">
                        {{ previewTz }}
                      </ElTag>
                    </div>
                    <ul class="run-list">
                      <li v-for="(run, idx) in nextRuns" :key="run.unix">
                        <span class="idx">#{{ idx + 1 }}</span>
                        <span class="ts">{{ formatRun(run) }}</span>
                        <span class="rel">{{ relativeTime(run.unix) }}</span>
                      </li>
                    </ul>
                  </div>
                </div>
              </ElFormItem>
            </ElCol>
          </ElRow>

          <!-- child task: dependency fields -->
          <template v-if="form.level === 2">
            <ElRow :gutter="24">
              <ElCol :span="8">
                <ElFormItem :label="t('task.dependencyTasks')">
                  <ElInput
                    v-model.trim="form.dependency_task_id"
                    placeholder="e.g. 1,2,3"
                    clearable
                  />
                </ElFormItem>
              </ElCol>
              <ElCol :span="8">
                <ElFormItem :label="t('task.dependencyStatus')">
                  <ElSelect v-model="form.dependency_status" style="width: 100%">
                    <ElOption :label="t('task.dependencyAll')" :value="1" />
                    <ElOption :label="t('task.dependencyAny')" :value="2" />
                  </ElSelect>
                </ElFormItem>
              </ElCol>
            </ElRow>
          </template>
        </ElCard>

        <!-- ── Execution ──────────────────────────────────────────────── -->
        <ElCard shadow="never" class="section-card mb-4">
          <template #header>
            <span class="section-title">{{ t('task.protocol') }}</span>
          </template>

          <ElRow :gutter="24">
            <!-- protocol -->
            <ElCol :span="8">
              <ElFormItem :label="t('task.protocol')">
                <ElSelect v-model="form.protocol" @change="handleProtocolChange" style="width: 100%">
                  <ElOption :label="t('task.protocolHttp')" :value="1" />
                  <ElOption :label="t('task.protocolRpc')" :value="2" />
                </ElSelect>
              </ElFormItem>
            </ElCol>

            <!-- HTTP method -->
            <ElCol :span="8" v-if="form.protocol === 1">
              <ElFormItem label="HTTP Method">
                <ElSelect v-model="form.http_method" style="width: 100%">
                  <ElOption label="GET" :value="1" />
                  <ElOption label="POST" :value="2" />
                </ElSelect>
              </ElFormItem>
            </ElCol>

            <!-- Shell: host selector -->
            <ElCol :span="16" v-if="form.protocol === 2">
              <ElFormItem :label="t('task.selectHosts')" prop="host_ids">
                <ElSelect
                  v-model="form.host_ids"
                  multiple
                  filterable
                  :placeholder="t('task.selectHosts')"
                  style="width: 100%"
                >
                  <ElOption
                    v-for="host in hostOptions"
                    :key="host.id"
                    :label="`${host.alias} - ${host.name}:${host.port}`"
                    :value="host.id"
                  />
                </ElSelect>
              </ElFormItem>
            </ElCol>
          </ElRow>

          <!-- command / URL -->
          <ElRow :gutter="24">
            <ElCol :span="20">
              <ElFormItem :label="form.protocol === 1 ? t('task.url') : t('task.command')" prop="command">
                <ElInput
                  v-model="form.command"
                  type="textarea"
                  :rows="5"
                  :placeholder="form.protocol === 1 ? t('task.urlPlaceholder') : t('task.commandPlaceholder')"
                />
              </ElFormItem>
            </ElCol>
          </ElRow>

          <!-- HTTP body (POST only) -->
          <ElRow :gutter="24" v-if="form.protocol === 1 && form.http_method === 2">
            <ElCol :span="18">
              <ElFormItem label="HTTP Body">
                <ElInput
                  v-model="form.http_body"
                  type="textarea"
                  :rows="4"
                  placeholder="Request body (JSON, form data, etc.)"
                />
              </ElFormItem>
            </ElCol>
          </ElRow>

          <!-- HTTP headers -->
          <ElRow :gutter="24" v-if="form.protocol === 1">
            <ElCol :span="18">
              <ElFormItem label="HTTP Headers">
                <ElInput
                  v-model="form.http_headers"
                  type="textarea"
                  :rows="3"
                  placeholder="Key: Value (one per line)"
                />
              </ElFormItem>
            </ElCol>
          </ElRow>

          <!-- Success pattern (HTTP only) -->
          <ElRow :gutter="24" v-if="form.protocol === 1">
            <ElCol :span="12">
              <ElFormItem label="Success Pattern">
                <ElInput
                  v-model.trim="form.success_pattern"
                  placeholder="Regex to match success response"
                  clearable
                />
              </ElFormItem>
            </ElCol>
          </ElRow>
        </ElCard>

        <!-- ── Concurrency & Retry ─────────────────────────────────────── -->
        <ElCard shadow="never" class="section-card mb-4">
          <template #header>
            <span class="section-title">{{ t('task.timeout') }} &amp; {{ t('task.retryCount') }}</span>
          </template>

          <ElRow :gutter="24">
            <!-- timeout -->
            <ElCol :span="8">
              <ElFormItem :label="t('task.timeout')" prop="timeout">
                <ElInputNumber
                  v-model="form.timeout"
                  :min="0"
                  controls-position="right"
                  style="width: 100%"
                />
              </ElFormItem>
            </ElCol>

            <!-- multi (single instance) -->
            <ElCol :span="8">
              <ElFormItem :label="t('task.multi')">
                <ElSelect v-model="form.multi" style="width: 100%">
                  <ElOption :label="t('task.multiYes')" :value="0" />
                  <ElOption :label="t('task.multiNo')" :value="1" />
                </ElSelect>
              </ElFormItem>
            </ElCol>
          </ElRow>

          <ElRow :gutter="24">
            <!-- retry count -->
            <ElCol :span="8">
              <ElFormItem :label="t('task.retryCount')" prop="retry_times">
                <ElInputNumber
                  v-model="form.retry_times"
                  :min="0"
                  :max="10"
                  controls-position="right"
                  style="width: 100%"
                />
              </ElFormItem>
            </ElCol>

            <!-- retry interval -->
            <ElCol :span="8">
              <ElFormItem :label="t('task.retryInterval')" prop="retry_interval">
                <ElInputNumber
                  v-model="form.retry_interval"
                  :min="0"
                  controls-position="right"
                  style="width: 100%"
                />
              </ElFormItem>
            </ElCol>
          </ElRow>
        </ElCard>

        <!-- ── Notification ───────────────────────────────────────────── -->
        <ElCard shadow="never" class="section-card mb-4">
          <template #header>
            <span class="section-title">{{ t('task.notifyStatus') }}</span>
          </template>

          <ElRow :gutter="24">
            <!-- notify status -->
            <ElCol :span="8">
              <ElFormItem :label="t('task.notifyStatus')">
                <ElSelect v-model="form.notify_status" @change="handleNotifyStatusChange" style="width: 100%">
                  <ElOption :label="t('task.notifyStatusNone')" :value="0" />
                  <ElOption :label="t('task.notifyStatusFailed')" :value="1" />
                  <ElOption :label="t('task.notifyStatusAll')" :value="2" />
                  <ElOption :label="t('task.notifyKeyword')" :value="3" />
                </ElSelect>
              </ElFormItem>
            </ElCol>

            <!-- notify type (when enabled) -->
            <ElCol :span="8" v-if="form.notify_status !== 0">
              <ElFormItem :label="t('task.notifyType')">
                <ElSelect v-model="form.notify_type" style="width: 100%">
                  <ElOption :label="t('task.notifyTypeEmail')" :value="0" />
                  <ElOption :label="t('task.notifyTypeSlack')" :value="1" />
                  <ElOption :label="t('task.notifyTypeWebhook')" :value="2" />
                </ElSelect>
              </ElFormItem>
            </ElCol>

            <!-- notify receiver: email -->
            <ElCol :span="8" v-if="form.notify_status !== 0 && form.notify_type === 0">
              <ElFormItem :label="t('task.notifyReceiver')">
                <ElSelect
                  v-model="selectedMailIds"
                  multiple
                  filterable
                  style="width: 100%"
                >
                  <ElOption
                    v-for="u in mailUsers"
                    :key="u.id"
                    :label="u.username"
                    :value="u.id"
                  />
                </ElSelect>
              </ElFormItem>
            </ElCol>

            <!-- notify receiver: slack -->
            <ElCol :span="8" v-if="form.notify_status !== 0 && form.notify_type === 1">
              <ElFormItem :label="t('task.notifyReceiver')">
                <ElSelect
                  v-model="selectedSlackIds"
                  multiple
                  filterable
                  style="width: 100%"
                >
                  <ElOption
                    v-for="c in slackChannels"
                    :key="c.id"
                    :label="c.name"
                    :value="c.id"
                  />
                </ElSelect>
              </ElFormItem>
            </ElCol>

            <!-- notify receiver: webhook -->
            <ElCol :span="8" v-if="form.notify_status !== 0 && form.notify_type === 2">
              <ElFormItem :label="t('task.notifyReceiver')">
                <ElSelect
                  v-model="selectedWebhookIds"
                  multiple
                  filterable
                  style="width: 100%"
                >
                  <ElOption
                    v-for="w in webhookUrls"
                    :key="w.id"
                    :label="w.name"
                    :value="w.id"
                  />
                </ElSelect>
              </ElFormItem>
            </ElCol>
          </ElRow>

          <!-- keyword match -->
          <ElRow :gutter="24" v-if="form.notify_status === 3">
            <ElCol :span="12">
              <ElFormItem :label="t('task.notifyKeyword')" prop="notify_keyword">
                <ElInput v-model.trim="form.notify_keyword" clearable />
              </ElFormItem>
            </ElCol>
          </ElRow>
        </ElCard>

        <!-- ── Template ───────────────────────────────────────────────── -->
        <ElCard shadow="never" class="section-card mb-4">
          <template #header>
            <span class="section-title">{{ t('task.template') }}</span>
          </template>

          <ElRow :gutter="24">
            <ElCol :span="12">
              <ElFormItem :label="t('task.template')">
                <ElSelect
                  v-model="selectedTemplateId"
                  filterable
                  clearable
                  :placeholder="t('task.templateHelp')"
                  style="width: 100%"
                  @change="handleTemplateChange"
                >
                  <ElOption
                    v-for="tpl in templateOptions"
                    :key="tpl.id"
                    :label="tpl.name"
                    :value="tpl.id"
                  />
                </ElSelect>
              </ElFormItem>
            </ElCol>
          </ElRow>
        </ElCard>

        <!-- ── Actions ────────────────────────────────────────────────── -->
        <ElFormItem>
          <ElButton type="primary" :loading="submitting" @click="handleSubmit" v-ripple>
            {{ t('task.save') }}
          </ElButton>
          <ElButton @click="handleCancel">{{ t('common.cancel') }}</ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, watch, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { useRoute, useRouter } from 'vue-router'
  import { Clock, InfoFilled, WarningFilled } from '@element-plus/icons-vue'
  import type { FormInstance, FormRules } from 'element-plus'
  import {
    fetchTaskDetail,
    fetchTaskStore,
    fetchTaskTags,
    fetchCronPreview,
    type CronRun
  } from '@/api/task'
  import { fetchHostList, type HostItem } from '@/api/host'
  import { fetchTemplateList, fetchTemplateDetail } from '@/api/template'
  import { fetchMail, fetchSlack, fetchWebhook } from '@/api/notification'
  import type { MailUser, SlackChannel, WebhookUrl } from '@/api/notification'

  defineOptions({ name: 'TaskEdit' })

  const { t } = useI18n()
  const route = useRoute()
  const router = useRouter()

  // ── State ─────────────────────────────────────────────────────────────────────

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    id: 0,
    name: '',
    tags: [] as string[],
    remark: '',
    level: 1,
    spec: '',
    dependency_status: 1,
    dependency_task_id: '',
    protocol: 2,
    http_method: 1,
    http_body: '',
    http_headers: '',
    success_pattern: '',
    command: '',
    host_ids: [] as number[],
    timeout: 3600,
    multi: 0,
    retry_times: 0,
    retry_interval: 0,
    notify_status: 0,
    notify_type: 0,
    notify_keyword: '',
    notify_receiver_id: ''
  })

  // Notification receiver selection (separated by type, like old frontend)
  const selectedMailIds = ref<number[]>([])
  const selectedSlackIds = ref<number[]>([])
  const selectedWebhookIds = ref<number[]>([])

  // Drop-down data sources
  const tagOptions = ref<string[]>([])
  const hostOptions = ref<HostItem[]>([])
  const mailUsers = ref<MailUser[]>([])
  const slackChannels = ref<SlackChannel[]>([])
  const webhookUrls = ref<WebhookUrl[]>([])
  const templateOptions = ref<{ id: number; name: string }[]>([])
  const selectedTemplateId = ref<number | null>(null)

  // Cron preview
  const nextRuns = ref<CronRun[]>([])
  const previewError = ref('')
  const previewTz = ref('')
  let cronDebounce: ReturnType<typeof setTimeout> | null = null

  // ── Computed ──────────────────────────────────────────────────────────────────

  const routeId = computed(() => {
    const id = route.params.id
    if (!id || id === '0') return 0
    return Number(id)
  })

  const isEdit = computed(() => routeId.value > 0)

  // ── Validation rules ──────────────────────────────────────────────────────────

  const rules = computed<FormRules>(() => {
    const r: FormRules = {
      name: [{ required: true, message: t('task.nameRequired'), trigger: 'blur' }],
      command: [{ required: true, message: t('task.commandRequired'), trigger: 'blur' }],
      timeout: [{ required: true, type: 'number', message: t('task.timeoutRequired'), trigger: 'blur' }],
      retry_times: [{ required: true, type: 'number', message: t('task.retryCountRequired'), trigger: 'blur' }],
      retry_interval: [{ required: true, type: 'number', message: t('task.retryIntervalRequired'), trigger: 'blur' }]
    }

    if (form.level === 1) {
      r.spec = [{ required: true, message: t('task.specRequired'), trigger: 'blur' }]
    }

    if (form.protocol === 2) {
      r.host_ids = [
        {
          required: true,
          type: 'array',
          min: 1,
          message: t('task.hostsRequired'),
          trigger: 'change'
        }
      ]
    }

    if (form.notify_status === 3) {
      r.notify_keyword = [{ required: true, message: t('task.notifyKeywordRequired'), trigger: 'blur' }]
    }

    return r
  })

  // ── Helpers ───────────────────────────────────────────────────────────────────

  function parseCronSpec(spec: string): { expr: string; tz: string } {
    if (!spec) return { expr: '', tz: '' }
    const match = spec.match(/^(?:CRON_TZ|TZ)=(\S+)\s+(.+)$/)
    if (match) return { tz: match[1], expr: match[2] }
    return { expr: spec, tz: '' }
  }

  // ── Data loading ──────────────────────────────────────────────────────────────

  async function loadTagOptions() {
    try {
      const data = await fetchTaskTags()
      tagOptions.value = Array.isArray(data) ? data : []
    } catch {
      // ignore
    }
  }

  async function loadHostOptions() {
    try {
      const res = await fetchHostList({ page: 1, page_size: 999 })
      const list = (res as any)?.data ?? res
      hostOptions.value = Array.isArray(list) ? list : (list?.data ?? [])
    } catch {
      // ignore
    }
  }

  async function loadNotificationOptions() {
    try {
      const mailRes = await fetchMail()
      mailUsers.value = mailRes?.mail_users ?? []
    } catch {
      // ignore
    }
    try {
      const slackRes = await fetchSlack()
      slackChannels.value = slackRes?.channels ?? []
    } catch {
      // ignore
    }
    try {
      const webhookRes = await fetchWebhook()
      webhookUrls.value = webhookRes?.webhook_urls ?? []
    } catch {
      // ignore
    }
  }

  async function loadTemplateOptions() {
    try {
      const res = await fetchTemplateList({ page: 1, page_size: 200 })
      const list = (res as any)?.data ?? []
      templateOptions.value = list.map((tpl: any) => ({ id: tpl.id, name: tpl.name }))
    } catch {
      // ignore
    }
  }

  async function loadDetail(id: number) {
    try {
      const data = await fetchTaskDetail(id)
      if (!data) {
        ElMessage.error(t('task.notFound'))
        router.push('/task/list')
        return
      }
      populateForm(data)
    } catch {
      router.push('/task/list')
    }
  }

  function populateForm(data: any) {
    // Strip CRON_TZ prefix from spec
    const { expr: specExpr } = parseCronSpec(data.spec || '')

    form.id = data.id
    form.name = data.name
    form.tags = data.tag ? data.tag.split(',').filter(Boolean) : []
    form.remark = data.remark || ''
    form.level = data.level ?? 1
    form.spec = specExpr
    form.dependency_status = data.dependency_status ?? 1
    form.dependency_task_id = data.dependency_task_id || ''
    form.protocol = data.protocol
    form.http_method = data.http_method ?? 1
    form.http_body = data.http_body || ''
    form.http_headers = data.http_headers || ''
    form.success_pattern = data.success_pattern || ''
    form.command = data.command || ''
    form.timeout = data.timeout ?? 3600
    form.multi = data.multi ?? 0
    form.retry_times = data.retry_times ?? 0
    form.retry_interval = data.retry_interval ?? 0
    form.notify_status = data.notify_status ?? 0
    form.notify_type = data.notify_type ?? 0
    form.notify_keyword = data.notify_keyword || ''
    form.notify_receiver_id = data.notify_receiver_id || ''

    // Shell host IDs
    const taskHosts: any[] = data.hosts || []
    form.host_ids = form.protocol === 2 ? taskHosts.map((h: any) => h.host_id) : []

    // Notify receivers
    selectedMailIds.value = []
    selectedSlackIds.value = []
    selectedWebhookIds.value = []
    if (form.notify_status > 0 && form.notify_receiver_id) {
      const ids = form.notify_receiver_id.split(',').filter(Boolean).map(Number)
      if (form.notify_type === 0) selectedMailIds.value = ids
      else if (form.notify_type === 1) selectedSlackIds.value = ids
      else if (form.notify_type === 2) selectedWebhookIds.value = ids
    }

    // Trigger cron preview if spec present
    if (specExpr) previewCron()
  }

  // ── Cron preview ──────────────────────────────────────────────────────────────

  function previewCron() {
    if (cronDebounce) clearTimeout(cronDebounce)
    cronDebounce = setTimeout(runPreview, 300)
  }

  async function runPreview() {
    const spec = (form.spec || '').trim()
    if (!spec || form.level !== 1) {
      nextRuns.value = []
      previewError.value = ''
      previewTz.value = ''
      return
    }
    try {
      const res: any = await fetchCronPreview({ spec, count: 6 })
      if (!res || res.valid === false) {
        previewError.value = res?.error || t('template.previewInvalid')
        nextRuns.value = []
        return
      }
      previewError.value = ''
      previewTz.value = res.timezone || ''
      nextRuns.value = Array.isArray(res.next_runs) ? res.next_runs : []
    } catch {
      previewError.value = t('template.previewInvalid')
      nextRuns.value = []
    }
  }

  function formatRun(run: CronRun): string {
    const d = new Date(run.iso || run.unix * 1000)
    if (isNaN(d.getTime())) return String(run.iso || '')
    const pad = (n: number) => String(n).padStart(2, '0')
    return (
      `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ` +
      `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
    )
  }

  function relativeTime(unix: number): string {
    const diffSec = unix - Math.floor(Date.now() / 1000)
    if (diffSec <= 0) return ''
    if (diffSec < 60) return t('template.inSeconds', { n: diffSec })
    const diffMin = Math.floor(diffSec / 60)
    if (diffMin < 60) return t('template.inMinutes', { n: diffMin })
    const diffH = Math.floor(diffMin / 60)
    if (diffH < 24) return t('template.inHours', { n: diffH })
    return t('template.inDays', { n: Math.floor(diffH / 24) })
  }

  // ── Event handlers ────────────────────────────────────────────────────────────

  function handleProtocolChange(val: number) {
    if (val === 1) {
      form.host_ids = []
      // Clear host_ids validation error
      formRef.value?.clearValidate('host_ids')
    }
  }

  function handleNotifyStatusChange(val: number) {
    if (val === 0) {
      form.notify_type = 0
      form.notify_keyword = ''
      formRef.value?.clearValidate('notify_keyword')
    }
  }

  async function handleTemplateChange(id: number | null) {
    if (!id) return
    try {
      const tpl = await fetchTemplateDetail(id)
      if (!tpl) return
      form.protocol = tpl.protocol
      form.command = tpl.command || ''
      form.http_method = tpl.http_method ?? 1
      form.http_body = tpl.http_body || ''
      form.http_headers = tpl.http_headers || ''
      form.success_pattern = tpl.success_pattern || ''
      if (tpl.spec) form.spec = tpl.spec
      if (tpl.tag) form.tags = tpl.tag.split(',').filter(Boolean)
      if (tpl.timeout && tpl.timeout > 0) form.timeout = tpl.timeout
      if (tpl.multi !== undefined) form.multi = tpl.multi
      if (tpl.retry_times && tpl.retry_times > 0) form.retry_times = tpl.retry_times
      if (tpl.retry_interval && tpl.retry_interval > 0) form.retry_interval = tpl.retry_interval
      // Notify fields are intentionally NOT copied from templates.
      // Templates don't store notify_receiver_id (receivers are task-level),
      // so importing notify_status=1 without receivers would leave the form
      // in an inconsistent state (notify enabled, receiver empty). Users
      // configure notifications explicitly per task.
      if (tpl.description) form.remark = tpl.description

      // Update editor language and clear host_ids if switching to HTTP
      handleProtocolChange(tpl.protocol)
      if (tpl.spec) previewCron()

      ElMessage.success(t('task.templateApplied'))
    } catch {
      // error handled by http interceptor
    } finally {
      selectedTemplateId.value = null
    }
  }

  // ── Submit ────────────────────────────────────────────────────────────────────

  async function handleSubmit() {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    // Validate notify receivers
    if (form.notify_status > 0) {
      if (form.notify_type === 0 && selectedMailIds.value.length === 0) {
        ElMessage.error(t('task.selectNotifyReceiver'))
        return
      }
      if (form.notify_type === 1 && selectedSlackIds.value.length === 0) {
        ElMessage.error(t('task.selectNotifyReceiver'))
        return
      }
      if (form.notify_type === 2 && selectedWebhookIds.value.length === 0) {
        ElMessage.error(t('task.selectNotifyReceiver'))
        return
      }
    }

    submitting.value = true
    try {
      // Build spec with CRON_TZ prefix stripped (we stored just the expr)
      const specToSave = form.spec

      // Build notify_receiver_id
      let notifyReceiverIds = ''
      if (form.notify_status > 0) {
        if (form.notify_type === 0) notifyReceiverIds = selectedMailIds.value.join(',')
        else if (form.notify_type === 1) notifyReceiverIds = selectedSlackIds.value.join(',')
        else if (form.notify_type === 2) notifyReceiverIds = selectedWebhookIds.value.join(',')
      }

      // Build host_id: comma-joined string for shell protocol
      const hostIdString = form.protocol === 2 ? form.host_ids.join(',') : ''

      await fetchTaskStore({
        ...(isEdit.value ? { id: form.id } : {}),
        name: form.name,
        tag: form.tags.join(','),
        spec: specToSave,
        level: form.level,
        dependency_status: form.dependency_status,
        dependency_task_id: form.dependency_task_id,
        protocol: form.protocol,
        http_method: form.http_method,
        http_body: form.http_body,
        http_headers: form.http_headers,
        success_pattern: form.success_pattern,
        command: form.command,
        host_id: hostIdString,
        timeout: form.timeout,
        multi: form.multi,
        retry_times: form.retry_times,
        retry_interval: form.retry_interval,
        notify_status: form.notify_status,
        notify_type: form.notify_type,
        notify_keyword: form.notify_keyword,
        notify_receiver_id: notifyReceiverIds,
        remark: form.remark
      })

      ElMessage.success(isEdit.value ? t('task.updateSuccess') : t('task.createSuccess'))
      router.push('/task/list')
    } catch {
      // error handled by http interceptor
    } finally {
      submitting.value = false
    }
  }

  function handleCancel() {
    router.push('/task/list')
  }

  // ── Watchers ──────────────────────────────────────────────────────────────────

  // When switching level, clear spec validation if not needed
  watch(
    () => form.level,
    (val) => {
      if (val !== 1) {
        nextRuns.value = []
        previewError.value = ''
        previewTz.value = ''
        formRef.value?.clearValidate('spec')
      }
    }
  )

  // Re-run spec preview when spec changes (debounced via blur/change in template)
  watch(
    () => form.spec,
    () => {
      if (!form.spec) {
        nextRuns.value = []
        previewError.value = ''
        previewTz.value = ''
      }
    }
  )

  // ── Lifecycle ─────────────────────────────────────────────────────────────────

  onMounted(async () => {
    await Promise.all([
      loadHostOptions(),
      loadTagOptions(),
      loadNotificationOptions(),
      loadTemplateOptions()
    ])

    if (isEdit.value) {
      loadDetail(routeId.value)
    } else {
      // Create mode: auto-apply template if ?template_id=X query is present
      const queryTplId = route.query.template_id
      if (queryTplId) {
        const tplId = Number(queryTplId)
        if (tplId > 0) handleTemplateChange(tplId)
      }
    }
  })

  watch(routeId, (newId) => {
    if (newId > 0) {
      loadDetail(newId)
    } else {
      // Reset form for create mode
      Object.assign(form, {
        id: 0,
        name: '',
        tags: [],
        remark: '',
        level: 1,
        spec: '',
        dependency_status: 1,
        dependency_task_id: '',
        protocol: 2,
        http_method: 1,
        http_body: '',
        http_headers: '',
        success_pattern: '',
        command: '',
        host_ids: [],
        timeout: 3600,
        multi: 0,
        retry_times: 0,
        retry_interval: 0,
        notify_status: 0,
        notify_type: 0,
        notify_keyword: '',
        notify_receiver_id: ''
      })
      selectedMailIds.value = []
      selectedSlackIds.value = []
      selectedWebhookIds.value = []
      nextRuns.value = []
      previewError.value = ''
      previewTz.value = ''
      selectedTemplateId.value = null
      formRef.value?.clearValidate()
    }
  })
</script>

<style scoped>
  .task-edit-page {
    display: flex;
    flex-direction: column;
  }

  .section-card {
    border: 1px solid var(--el-border-color-light);
  }

  .section-title {
    font-size: 14px;
    font-weight: 500;
    color: var(--el-text-color-primary);
  }

  /* Cron expression help popover */
  .cron-help h4 {
    margin: 0 0 6px;
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }
  .cron-help ul {
    list-style: none;
    padding: 0;
    margin: 0 0 12px;
  }
  .cron-help li {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 3px 0;
    font-size: 13px;
    color: var(--el-text-color-regular);
  }
  .cron-help code {
    font-family: monospace;
    background: var(--el-fill-color-light);
    padding: 2px 8px;
    border-radius: 4px;
    color: var(--el-color-primary);
    min-width: 110px;
    display: inline-block;
  }

  /* Cron preview box */
  .cron-preview {
    border: 1px solid var(--el-border-color-light);
    border-radius: 6px;
    padding: 12px 16px;
    background: var(--el-fill-color-blank);
    min-height: 60px;
    width: 100%;
    transition: border-color 0.2s;
  }
  .cron-preview.is-empty {
    background: var(--el-fill-color-lighter);
  }
  .cron-preview.is-invalid {
    border-color: var(--el-color-danger);
    background: var(--el-color-danger-light-9);
  }
  .preview-state {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
  }
  .preview-state.muted {
    color: var(--el-text-color-secondary);
  }
  .preview-state.error {
    color: var(--el-color-danger);
  }
  .preview-title {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    font-weight: 500;
    color: var(--el-text-color-primary);
    margin-bottom: 10px;
  }
  .tz-tag {
    margin-left: 4px;
    font-weight: normal;
  }
  .run-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 4px 20px;
  }
  .run-list li {
    display: flex;
    align-items: baseline;
    gap: 8px;
    font-size: 13px;
    font-variant-numeric: tabular-nums;
  }
  .run-list .idx {
    color: var(--el-text-color-placeholder);
    font-size: 11px;
    min-width: 22px;
  }
  .run-list .ts {
    color: var(--el-text-color-primary);
  }
  .run-list .rel {
    margin-left: auto;
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }
  @media (max-width: 768px) {
    .run-list {
      grid-template-columns: 1fr;
    }
  }
</style>
