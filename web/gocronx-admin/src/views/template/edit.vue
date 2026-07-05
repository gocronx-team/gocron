<!-- Template create / edit form -->
<!-- Routes: /template/create  or  /template/edit/:id -->
<template>
  <div class="template-edit-page">
    <ElCard shadow="never">
      <template #header>
        <span class="text-base font-medium">
          {{ isEdit ? t('template.editTitle') : t('template.createTitle') }}
        </span>
      </template>

      <ElForm
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="130px"
        label-position="right"
        @submit.prevent
      >
        <!-- ── Basic Info ─────────────────────────────────────────────── -->
        <ElDivider content-position="left">{{ t('template.sectionBasic') }}</ElDivider>

        <ElRow :gutter="24">
          <ElCol :span="8">
            <ElFormItem :label="t('template.name')" prop="name">
              <ElInput
                v-model.trim="form.name"
                :placeholder="t('template.namePlaceholder')"
                clearable
              />
            </ElFormItem>
          </ElCol>
          <ElCol :span="8">
            <ElFormItem :label="t('template.category')" prop="category">
              <ElSelect
                v-model="form.category"
                filterable
                allow-create
                default-first-option
                :placeholder="t('template.selectCategory')"
                style="width: 100%"
              >
                <ElOption value="backup" :label="t('template.category_backup')" />
                <ElOption value="cleanup" :label="t('template.category_cleanup')" />
                <ElOption value="monitor" :label="t('template.category_monitor')" />
                <ElOption value="deploy" :label="t('template.category_deploy')" />
                <ElOption value="api" :label="t('template.category_api')" />
                <ElOption value="custom" :label="t('template.category_custom')" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
          <ElCol :span="8">
            <ElFormItem :label="t('task.tag')">
              <ElInput v-model.trim="form.tag" :placeholder="t('task.tagPlaceholder')" clearable />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="24">
          <ElCol :span="16">
            <ElFormItem :label="t('task.description')">
              <ElInput
                v-model="form.description"
                type="textarea"
                :rows="2"
                :placeholder="t('task.descriptionPlaceholder')"
              />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <!-- ── Schedule ───────────────────────────────────────────────── -->
        <ElDivider content-position="left">{{ t('template.sectionSchedule') }}</ElDivider>

        <ElRow :gutter="24">
          <ElCol :span="12">
            <ElFormItem :label="t('template.spec')">
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
                        <li
                          ><code>0 * * * * *</code>
                          <span>{{ t('template.cronEveryMinute') }}</span></li
                        >
                        <li
                          ><code>*/20 * * * * *</code>
                          <span>{{ t('template.cronEvery20Sec') }}</span></li
                        >
                        <li
                          ><code>0 30 21 * * *</code>
                          <span>{{ t('template.cronEveryDay2130') }}</span></li
                        >
                        <li
                          ><code>0 0 23 * * 6</code>
                          <span>{{ t('template.cronEverySat23') }}</span></li
                        >
                      </ul>
                      <h4>{{ t('template.cronShortcut') }}</h4>
                      <ul>
                        <li
                          ><code>@reboot</code> <span>{{ t('template.cronReboot') }}</span></li
                        >
                        <li
                          ><code>@yearly</code> <span>{{ t('template.cronYearly') }}</span></li
                        >
                        <li
                          ><code>@monthly</code> <span>{{ t('template.cronMonthly') }}</span></li
                        >
                        <li
                          ><code>@weekly</code> <span>{{ t('template.cronWeekly') }}</span></li
                        >
                        <li
                          ><code>@daily</code> <span>{{ t('template.cronDaily') }}</span></li
                        >
                        <li
                          ><code>@hourly</code> <span>{{ t('template.cronHourly') }}</span></li
                        >
                        <li
                          ><code>@every 30s</code> <span>{{ t('template.cronEvery30s') }}</span></li
                        >
                        <li
                          ><code>@every 1m20s</code>
                          <span>{{ t('template.cronEvery1m20s') }}</span></li
                        >
                      </ul>
                    </div>
                  </ElPopover>
                </template>
              </ElInput>
            </ElFormItem>
          </ElCol>
          <ElCol :span="8">
            <ElFormItem :label="t('template.timezone')">
              <ElSelect
                v-model="form.timezone"
                filterable
                clearable
                :placeholder="t('template.timezoneServer')"
                style="width: 100%"
                @change="previewCron"
              >
                <ElOptionGroup
                  v-for="group in timezoneGroups"
                  :key="group.label"
                  :label="group.label"
                >
                  <ElOption v-for="tz in group.zones" :key="tz" :label="tz" :value="tz" />
                </ElOptionGroup>
              </ElSelect>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="24">
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

        <!-- ── Execution ──────────────────────────────────────────────── -->
        <ElDivider content-position="left">{{ t('template.sectionExecution') }}</ElDivider>

        <ElRow :gutter="24">
          <ElCol :span="8">
            <ElFormItem :label="t('template.protocol')">
              <ElSelect v-model="form.protocol" style="width: 100%">
                <ElOption :value="1" label="HTTP" />
                <ElOption :value="2" label="Shell / RPC" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
          <ElCol :span="8" v-if="form.protocol === 1">
            <ElFormItem :label="t('template.httpMethod')">
              <ElSelect v-model="form.http_method" style="width: 100%">
                <ElOption :value="1" label="GET" />
                <ElOption :value="2" label="POST" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow :gutter="24">
          <ElCol :span="20">
            <ElFormItem :label="commandLabel" prop="command">
              <ElInput
                v-model="form.command"
                type="textarea"
                :rows="6"
                :placeholder="commandPlaceholder"
                style="font-family: monospace"
              />
              <div class="form-help">{{ t('template.templateVarTip') }}</div>
              <div class="form-help">{{ t('template.varSyntaxTip') }}</div>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow v-if="form.protocol === 1 && form.http_method === 2" :gutter="24">
          <ElCol :span="20">
            <ElFormItem :label="t('template.httpBody')">
              <ElInput
                v-model="form.http_body"
                type="textarea"
                :rows="4"
                placeholder='{"key": "value"}'
                style="font-family: monospace"
              />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow v-if="form.protocol === 1" :gutter="24">
          <ElCol :span="20">
            <ElFormItem :label="t('template.httpHeaders')">
              <ElInput
                v-model="form.http_headers"
                type="textarea"
                :rows="3"
                placeholder="X-Key: value"
                style="font-family: monospace"
              />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <ElRow v-if="form.protocol === 1" :gutter="24">
          <ElCol :span="12">
            <ElFormItem :label="t('template.successPattern')">
              <ElInput
                v-model.trim="form.success_pattern"
                :placeholder="t('template.successPatternPlaceholder')"
                clearable
              />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <!-- ── Runtime ────────────────────────────────────────────────── -->
        <ElDivider content-position="left">{{ t('template.sectionRuntime') }}</ElDivider>

        <ElRow :gutter="24">
          <ElCol :span="6">
            <ElFormItem :label="t('task.timeout')">
              <ElInputNumber v-model="form.timeout" :min="0" :max="86400" style="width: 100%" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="6">
            <ElFormItem :label="t('template.singleInstance')">
              <ElSelect v-model="form.multi" style="width: 100%">
                <ElOption :value="0" :label="t('template.singleInstanceYes')" />
                <ElOption :value="1" :label="t('template.singleInstanceNo')" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
          <ElCol :span="6">
            <ElFormItem :label="t('task.retryCount')">
              <ElInputNumber v-model="form.retry_times" :min="0" :max="10" style="width: 100%" />
            </ElFormItem>
          </ElCol>
          <ElCol :span="6">
            <ElFormItem :label="t('task.retryInterval')">
              <ElInputNumber
                v-model="form.retry_interval"
                :min="0"
                :max="3600"
                style="width: 100%"
              />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <!-- ── Notification ───────────────────────────────────────────── -->
        <ElDivider content-position="left">{{ t('template.sectionNotify') }}</ElDivider>

        <ElRow :gutter="24">
          <ElCol :span="8">
            <ElFormItem :label="t('task.notifyStatus')">
              <ElCheckboxGroup v-model="notifyConditions">
                <ElCheckbox :value="1">{{ t('task.notifyStatusFailed') }}</ElCheckbox>
                <ElCheckbox :value="2">{{ t('task.notifyStatusSuccess') }}</ElCheckbox>
                <ElCheckbox :value="4">{{ t('task.notifyKeyword') }}</ElCheckbox>
              </ElCheckboxGroup>
            </ElFormItem>
          </ElCol>
          <ElCol :span="8" v-if="form.notify_status > 0">
            <ElFormItem :label="t('task.notifyType')">
              <ElSelect v-model="form.notify_type" style="width: 100%">
                <ElOption :value="0" :label="t('task.notifyTypeEmail')" />
                <ElOption :value="1" :label="t('task.notifyTypeSlack')" />
                <ElOption :value="2" :label="t('task.notifyTypeWebhook')" />
              </ElSelect>
            </ElFormItem>
          </ElCol>
          <ElCol :span="8" v-if="(form.notify_status & 4) !== 0">
            <ElFormItem :label="t('task.notifyKeyword')">
              <ElInput
                v-model.trim="form.notify_keyword"
                :placeholder="t('template.notifyKeywordPlaceholder')"
                clearable
              />
            </ElFormItem>
          </ElCol>
          <ElCol :span="8" v-if="(form.notify_status & 4) !== 0">
            <ElFormItem :label="t('task.notifyKeywordRegex')">
              <ElSwitch v-model="form.notify_keyword_regex" :active-value="1" :inactive-value="0" />
              <span class="regex-hint">{{ t('task.notifyKeywordRegexHint') }}</span>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <!-- ── Log retention ──────────────────────────────────────────── -->
        <ElDivider content-position="left">{{ t('template.sectionLogRetention') }}</ElDivider>

        <ElRow :gutter="24">
          <ElCol :span="12">
            <ElFormItem :label="t('template.logRetentionDays')">
              <ElInputNumber v-model="form.log_retention_days" :min="0" :max="3650" />
              <span class="form-help inline-help">{{ t('template.logRetentionTip') }}</span>
            </ElFormItem>
          </ElCol>
        </ElRow>
      </ElForm>
    </ElCard>

    <!-- Sticky bottom action bar -->
    <div class="form-actions-bar">
      <ElButton size="large" @click="handleCancel">
        {{ t('template.cancel') }}
      </ElButton>
      <ElButton type="primary" size="large" :loading="submitting" @click="handleSubmit">
        <ElIcon class="mr-1"><Check /></ElIcon>
        {{ t('template.save') }}
      </ElButton>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { useRoute, useRouter } from 'vue-router'
  import { Check, Clock, InfoFilled, WarningFilled } from '@element-plus/icons-vue'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchTemplateDetail, fetchTemplateStore } from '@/api/template'
  import { fetchCronPreview } from '@/api/task'

  defineOptions({ name: 'TemplateEdit' })

  const { t } = useI18n()
  const route = useRoute()
  const router = useRouter()

  interface CronRun {
    iso: string
    unix: number
    weekday: number
  }

  // ── State ────────────────────────────────────────────────────────────────────
  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const nextRuns = ref<CronRun[]>([])
  const previewError = ref('')
  const previewTz = ref('')
  let cronDebounce: ReturnType<typeof setTimeout> | null = null

  const emptyForm = () => ({
    id: 0,
    name: '',
    category: 'custom',
    tag: '',
    description: '',
    spec: '',
    timezone: '',
    protocol: 2,
    http_method: 1,
    command: '',
    http_body: '',
    http_headers: '',
    success_pattern: '',
    timeout: 300,
    multi: 1,
    retry_times: 0,
    retry_interval: 0,
    notify_status: 0,
    notify_type: 0,
    notify_keyword: '',
    notify_keyword_regex: 0,
    log_retention_days: 0
  })

  const form = reactive(emptyForm())

  // 通知触发条件多选(位掩码 1=失败 2=成功 4=关键字)UI 中间态,同步到 form.notify_status
  const notifyConditions = ref<number[]>([])
  watch(notifyConditions, (vals) => {
    form.notify_status = vals.reduce((sum, v) => sum + v, 0)
    if (form.notify_status === 0) {
      form.notify_type = 0
    }
    if ((form.notify_status & 4) === 0) {
      form.notify_keyword = ''
    }
  })

  // ── Computed ─────────────────────────────────────────────────────────────────
  const routeId = computed(() => {
    const id = route.params.id
    if (!id || id === '0') return 0
    return Number(id)
  })

  const isEdit = computed(() => routeId.value > 0)

  const commandLabel = computed(() => (form.protocol === 1 ? 'URL' : t('template.command')))

  const commandPlaceholder = computed(() =>
    form.protocol === 1 ? 'https://example.com/api/hook' : t('template.commandPlaceholder')
  )

  const timezoneGroups = computed(() => {
    try {
      const zones = (Intl as any).supportedValuesOf?.('timeZone') as string[] | undefined
      if (!zones || !zones.length) throw new Error('no zones')
      const groups: Record<string, string[]> = { UTC: ['UTC'] }
      for (const tz of zones) {
        const region = tz.split('/')[0]
        if (!groups[region]) groups[region] = []
        groups[region].push(tz)
      }
      const priority = ['UTC', 'Asia', 'America', 'Europe', 'Pacific', 'Australia', 'Africa']
      const sorted = Object.keys(groups).sort((a, b) => {
        const ai = priority.indexOf(a)
        const bi = priority.indexOf(b)
        if (ai !== -1 && bi !== -1) return ai - bi
        if (ai !== -1) return -1
        if (bi !== -1) return 1
        return a.localeCompare(b)
      })
      return sorted.map((region) => ({ label: region, zones: groups[region] }))
    } catch {
      return [
        {
          label: 'All',
          zones: ['UTC', 'Asia/Shanghai', 'America/New_York', 'Europe/London']
        }
      ]
    }
  })

  // ── Validation rules ─────────────────────────────────────────────────────────
  const rules = computed<FormRules>(() => ({
    name: [{ required: true, message: t('template.nameRequired'), trigger: 'blur' }],
    category: [{ required: true, message: t('template.categoryRequired'), trigger: 'blur' }],
    command: [{ required: true, message: t('template.commandRequired'), trigger: 'blur' }]
  }))

  // ── Cron preview ─────────────────────────────────────────────────────────────
  function previewCron() {
    if (cronDebounce) clearTimeout(cronDebounce)
    cronDebounce = setTimeout(runPreview, 300)
  }

  async function runPreview() {
    const spec = form.spec.trim()
    if (!spec) {
      nextRuns.value = []
      previewError.value = ''
      previewTz.value = ''
      return
    }
    try {
      const res: any = await fetchCronPreview({
        spec,
        timezone: form.timezone || undefined,
        count: 6
      })
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

  // ── Data loading ─────────────────────────────────────────────────────────────
  async function loadDetail(id: number) {
    try {
      const data = await fetchTemplateDetail(id)
      if (!data) {
        ElMessage.error(t('template.notFound'))
        router.push('/template/list')
        return
      }
      form.id = data.id
      form.name = data.name ?? ''
      form.category = data.category || 'custom'
      form.tag = data.tag ?? ''
      form.description = data.description ?? ''
      form.spec = data.spec ?? ''
      form.timezone = data.timezone ?? ''
      // backend binds oneof=1 2 strictly — coerce missing/0 to defaults
      form.protocol = data.protocol === 1 || data.protocol === 2 ? data.protocol : 2
      form.http_method = data.http_method === 2 ? 2 : 1
      form.command = data.command ?? ''
      form.http_body = data.http_body ?? ''
      form.http_headers = data.http_headers ?? ''
      form.success_pattern = data.success_pattern ?? ''
      form.timeout = data.timeout ?? 300
      form.multi = data.multi === 0 ? 0 : 1
      form.retry_times = data.retry_times ?? 0
      form.retry_interval = data.retry_interval ?? 0
      form.notify_status = data.notify_status ?? 0
      notifyConditions.value = [1, 2, 4].filter((b) => (form.notify_status & b) !== 0)
      form.notify_type = data.notify_type ?? 0
      form.notify_keyword = data.notify_keyword ?? ''
      form.notify_keyword_regex = data.notify_keyword_regex ?? 0
      form.log_retention_days = data.log_retention_days ?? 0

      previewCron()
    } catch {
      router.push('/template/list')
    }
  }

  // ── Submit ───────────────────────────────────────────────────────────────────
  async function handleSubmit() {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    submitting.value = true
    try {
      await fetchTemplateStore({
        ...(isEdit.value ? { id: form.id } : {}),
        name: form.name,
        category: form.category,
        tag: form.tag,
        description: form.description,
        spec: form.spec,
        timezone: form.timezone,
        protocol: form.protocol,
        http_method: form.http_method,
        command: form.command,
        http_body: form.http_body,
        http_headers: form.http_headers,
        success_pattern: form.success_pattern,
        timeout: form.timeout,
        multi: form.multi,
        retry_times: form.retry_times,
        retry_interval: form.retry_interval,
        notify_status: form.notify_status,
        notify_type: form.notify_type,
        notify_keyword: form.notify_keyword,
        notify_keyword_regex: form.notify_keyword_regex,
        log_retention_days: form.log_retention_days
      })
      ElMessage.success(isEdit.value ? t('template.updateSuccess') : t('template.createSuccess'))
      router.push('/template/list')
    } catch {
      // error toast handled by http interceptor
    } finally {
      submitting.value = false
    }
  }

  function handleCancel() {
    router.push('/template/list')
  }

  // ── Lifecycle ─────────────────────────────────────────────────────────────────
  onMounted(() => {
    if (isEdit.value) {
      loadDetail(routeId.value)
    }
  })

  watch(routeId, (newId) => {
    if (newId > 0) {
      loadDetail(newId)
    } else {
      Object.assign(form, emptyForm())
      nextRuns.value = []
      previewError.value = ''
      previewTz.value = ''
      formRef.value?.clearValidate()
    }
  })
</script>

<style scoped>
  .template-edit-page {
    display: flex;
    flex-direction: column;
    padding-bottom: 72px; /* leave room for sticky action bar */
  }

  .form-help {
    margin-top: 4px;
    font-size: 12px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
  }

  .inline-help {
    margin-left: 12px;
  }

  /* Cron expression help popover */
  .cron-help h4 {
    margin: 0 0 6px;
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .cron-help ul {
    padding: 0;
    margin: 0 0 12px;
    list-style: none;
  }

  .cron-help li {
    display: flex;
    gap: 10px;
    align-items: center;
    padding: 3px 0;
    font-size: 13px;
    color: var(--el-text-color-regular);
  }

  .cron-help code {
    display: inline-block;
    min-width: 110px;
    padding: 2px 8px;
    font-family: monospace;
    color: var(--el-color-primary);
    background: var(--el-fill-color-light);
    border-radius: 4px;
  }

  /* Cron preview box */
  .cron-preview {
    width: 100%;
    min-height: 60px;
    padding: 12px 16px;
    background: var(--el-fill-color-blank);
    border: 1px solid var(--el-border-color-light);
    border-radius: 6px;
    transition: border-color 0.2s;
  }

  .cron-preview.is-empty {
    background: var(--el-fill-color-lighter);
  }

  .cron-preview.is-invalid {
    background: var(--el-color-danger-light-9);
    border-color: var(--el-color-danger);
  }

  .preview-state {
    display: flex;
    gap: 6px;
    align-items: center;
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
    gap: 6px;
    align-items: center;
    margin-bottom: 10px;
    font-size: 13px;
    font-weight: 500;
    color: var(--el-text-color-primary);
  }

  .tz-tag {
    margin-left: 4px;
    font-weight: normal;
  }

  .run-list {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 4px 20px;
    padding: 0;
    margin: 0;
    list-style: none;
  }

  .run-list li {
    display: flex;
    gap: 8px;
    align-items: baseline;
    font-size: 13px;
    font-variant-numeric: tabular-nums;
  }

  .run-list .idx {
    min-width: 22px;
    font-size: 11px;
    color: var(--el-text-color-placeholder);
  }

  .run-list .ts {
    color: var(--el-text-color-primary);
  }

  .run-list .rel {
    margin-left: auto;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  @media (width <= 768px) {
    .run-list {
      grid-template-columns: 1fr;
    }

    /* Sticky action bar: full width, centred buttons on mobile */
    .form-actions-bar {
      gap: 8px;
      justify-content: center;
      padding: 12px 16px;
    }

    /* Give the form extra bottom clearance for the taller bar */
    .template-edit-page {
      padding-bottom: 80px;
    }
  }

  /* Sticky action bar */
  .form-actions-bar {
    position: sticky;
    bottom: 0;
    z-index: 10;
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: flex-end;
    padding: 14px 28px;
    margin: 16px -20px -20px;
    background: var(--el-bg-color);
    border-top: 1px solid var(--el-border-color-light);
  }

  :deep(.el-divider__text) {
    padding: 0 12px;
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    background: var(--el-bg-color);
  }
</style>
