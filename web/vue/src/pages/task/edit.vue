<template>
  <el-main>
    <el-form ref="form" :model="form" :rules="formRules" label-width="auto">
        <el-input v-model="form.id" type="hidden"></el-input>
        <el-row>
          <el-col :span="12">
            <el-form-item :label="t('task.name')" prop="name">
              <el-input v-model.trim="form.name"></el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('task.tag')">
              <el-select
                v-model="form.tags"
                multiple
                filterable
                allow-create
                default-first-option
                collapse-tags
                collapse-tags-tooltip
                :placeholder="t('task.tagPlaceholder')"
                style="width: 100%">
                <el-option
                  v-for="tag in tagOptions"
                  :key="tag"
                  :label="tag"
                  :value="tag">
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row v-if="form.level === 1">
          <el-col>
            <el-alert type="info" :closable="false">
              <span v-html="t('task.mainTaskTip')"></span>
            </el-alert>
            <el-alert type="info" :closable="false">
              <span v-html="t('task.dependencyTip')"></span>
            </el-alert> <br>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="7">
            <el-form-item :label="t('task.type')">
              <el-select v-model.trim="form.level" :disabled="form.id !== '' ">
                <el-option
                  v-for="item in levelList"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value">
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="7" v-if="form.level === 1">
            <el-form-item :label="t('task.dependency')">
              <el-select v-model.trim="form.dependency_status">
                <el-option
                  v-for="item in dependencyStatusList"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value">
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="10">
            <el-form-item :label="t('task.childTaskId')" v-if="form.level === 1">
              <el-input v-model.trim="form.dependency_task_id" :placeholder="t('task.childTaskIdPlaceholder')"></el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row v-if="form.level === 1">
          <el-col :span="12">
            <el-form-item :label="t('task.cronExpression')" prop="spec">
              <el-input v-model.trim="form.spec"
                        :placeholder="t('task.cronPlaceholder')">
                <template #append>
                  <el-popover
                    placement="bottom"
                    :width="500"
                    trigger="click">
                    <template #reference>
                      <el-button>{{ t('task.cronExample') }}</el-button>
                    </template>
                    <div>
                      <h4>{{ t('task.cronStandard') }}</h4>
                      <ul style="padding-left: 20px; margin: 10px 0;">
                        <li>0 * * * * * - {{ t('message.everyMinute') }}</li>
                        <li>*/20 * * * * * - {{ t('message.every20Seconds') }}</li>
                        <li>0 30 21 * * * - {{ t('message.everyDay21_30') }}</li>
                        <li>0 0 23 * * 6 - {{ t('message.everySaturday23') }}</li>
                      </ul>
                      <h4>{{ t('task.cronShortcut') }}</h4>
                      <ul style="padding-left: 20px; margin: 10px 0;">
                        <li>@reboot - {{ t('message.reboot') }}</li>
                        <li>@yearly - {{ t('message.yearly') }}</li>
                        <li>@monthly - {{ t('message.monthly') }}</li>
                        <li>@weekly - {{ t('message.weekly') }}</li>
                        <li>@daily - {{ t('message.daily') }}</li>
                        <li>@hourly - {{ t('message.hourly') }}</li>
                        <li>@every 30s - {{ t('message.every30s') }}</li>
                        <li>@every 1m20s - {{ t('message.every1m20s') }}</li>
                      </ul>
                    </div>
                  </el-popover>
                </template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('task.timezone')">
              <el-select
                v-model="form.timezone"
                filterable
                clearable
                :placeholder="t('task.timezoneServer')">
                <el-option-group
                  v-for="group in timezoneGroups"
                  :key="group.label"
                  :label="group.label">
                  <el-option
                    v-for="tz in group.zones"
                    :key="tz"
                    :label="tz"
                    :value="tz">
                  </el-option>
                </el-option-group>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="8">
            <el-form-item :label="t('task.protocol')">
              <el-select v-model.trim="form.protocol" @change="handleProtocolChange">
                <el-option
                  v-for="item in protocolList"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value">
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8" v-if="form.protocol === 1 ">
            <el-form-item :label="t('task.httpMethod')">
              <el-select key="http-method" v-model.trim="form.http_method">
                <el-option
                  v-for="item in httpMethods"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value">
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8" v-else>
            <el-form-item :label="t('task.taskNode')" prop="host_ids">
              <el-select
                key="shell"
                v-model="form.host_ids"
                filterable
                multiple
                :placeholder="t('task.taskNodePlaceholder')">
                <el-option
                  v-for="item in hosts"
                  :key="item.id"
                  :label="item.alias + ' - ' + item.name"
                  :value="item.id">
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="16">
            <el-form-item :label="t('task.command')" prop="command">
              <el-input
                type="textarea"
                :rows="5"
                :placeholder="commandPlaceholder"
                v-model="form.command"
                @blur="validateCommand">
              </el-input>
              <div v-if="commandWarning" class="command-warning" style="color: #E6A23C; font-size: 12px; margin-top: 4px;">
                {{ commandWarning }}
              </div>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row v-if="Number(form.protocol) === 1 && Number(form.http_method) === 2">
          <el-col :span="16">
            <el-form-item :label="t('task.httpBody')">
              <el-input
                type="textarea"
                :rows="4"
                :placeholder="t('task.httpBodyPlaceholder')"
                v-model="form.http_body">
              </el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row v-if="Number(form.protocol) === 1">
          <el-col :span="16">
            <el-form-item :label="t('task.httpHeaders')">
              <el-input
                type="textarea"
                :rows="3"
                :placeholder="t('task.httpHeadersPlaceholder')"
                v-model="form.http_headers">
              </el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row v-if="Number(form.protocol) === 1">
          <el-col :span="12">
            <el-form-item :label="t('task.successPattern')">
              <el-input
                v-model.trim="form.success_pattern"
                :placeholder="t('task.successPatternPlaceholder')">
              </el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col>
            <el-alert
              :title="t('task.timeoutTip')"
              type="info"
              :closable="false">
            </el-alert>
            <el-alert
              :title="t('task.singleInstanceTip')"
              type="info"
              :closable="false">
            </el-alert> <br>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item :label="t('task.timeout')" prop="timeout">
              <el-input v-model.number.trim="form.timeout"></el-input>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('task.singleInstance')">
              <el-select v-model.trim="form.multi">
                <el-option
                  v-for="item in runStatusList"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value">
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
        <el-col :span="12">
          <el-form-item :label="t('task.retryTimes')" prop="retry_times">
            <el-input v-model.number.trim="form.retry_times"
                      :placeholder="t('task.retryTimesPlaceholder')"></el-input>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="t('task.retryInterval')" prop="retry_interval">
            <el-input v-model.number.trim="form.retry_interval" :placeholder="t('task.retryIntervalPlaceholder')"></el-input>
          </el-form-item>
        </el-col>
        </el-row>
        <el-row>
          <el-col :span="8">
            <el-form-item :label="t('task.notification')">
              <el-select v-model.trim="form.notify_status">
                <el-option
                  v-for="item in notifyStatusList"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value">
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8" v-if="form.notify_status !== 0">
            <el-form-item :label="t('task.notifyType')">
              <el-select v-model.trim="form.notify_type">
                <el-option
                  v-for="item in notifyTypes"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                  >
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8"
                  v-if="form.notify_status !== 0 && form.notify_type === 0">
            <el-form-item :label="t('task.notifyReceiver')">
              <el-select key="notify-mail" v-model="selectedMailNotifyIds" filterable multiple :placeholder="t('task.notifyReceiverPlaceholder')">
                <el-option
                  v-for="item in mailUsers"
                  :key="item.id"
                  :label="item.username"
                  :value="item.id">
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>

          <el-col :span="8"
                  v-if="form.notify_status !== 0 && form.notify_type === 1">
            <el-form-item :label="t('task.notifyChannel')">
              <el-select key="notify-slack" v-model="selectedSlackNotifyIds" filterable multiple :placeholder="t('task.notifyReceiverPlaceholder')">
                <el-option
                  v-for="item in slackChannels"
                  :key="item.id"
                  :label="item.name"
                  selected="true"
                  :value="item.id">
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>

          <el-col :span="8"
                  v-if="form.notify_status !== 0 && form.notify_type === 2">
            <el-form-item :label="t('task.notifyReceiver')">
              <el-select key="notify-webhook" v-model="selectedWebhookNotifyIds" filterable multiple :placeholder="t('task.notifyReceiverPlaceholder')">
                <el-option
                  v-for="item in webhookUrls"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id">
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row v-if="form.notify_status === 3">
          <el-col :span="12">
            <el-form-item :label="t('task.notifyKeyword')" prop="notify_keyword">
              <el-input v-model.trim="form.notify_keyword" :placeholder="t('task.notifyKeywordPlaceholder')"></el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item :label="t('task.logRetentionDays')">
              <el-input-number v-model="form.log_retention_days" :min="0" :max="3650"></el-input-number>
              <span style="margin-left: 8px; color: #909399; font-size: 12px;">{{ t('task.logRetentionDaysTip') }}</span>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="16">
            <el-form-item :label="t('task.remark')">
              <el-input
                type="textarea"
                :rows="3"
                v-model="form.remark">
              </el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item>
          <el-button type="primary" @click="submit">{{ t('common.save') }}</el-button>
          <el-button @click="cancel">{{ t('common.cancel') }}</el-button>
        </el-form-item>
      </el-form>
    </el-main>
</template>


<script>
import { useI18n } from 'vue-i18n'
import taskService from '../../api/task'
import notificationService from '../../api/notification'
import { validateCronSpec, getCronExamples, extractTimezone } from '../../utils/cronValidator'

const createDefaultForm = () => ({
  id: '',
  name: '',
  tag: '',
  tags: [],
  level: 1,
  dependency_status: 1,
  dependency_task_id: '',
  spec: '',
  timezone: '',
  protocol: 2,
  http_method: 1,
  http_body: '',
  http_headers: '',
  success_pattern: '',
  command: '',
  host_id: '',
  host_ids: [],
  timeout: 3600,
  multi: 0,
  notify_status: 0,
  notify_type: 0,
  notify_receiver_id: '',
  notify_keyword: '',
  retry_times: 0,
  retry_interval: 0,
  log_retention_days: 0,
  remark: ''
})

export default {
  name: 'task-edit',
  setup() {
    const { t, locale } = useI18n()
    return { t, locale }
  },
  data () {
    return {
      form: createDefaultForm(),
      formRules: {},
      httpMethods: [
        {
          value: 1,
          label: 'get'
        },
        {
          value: 2,
          label: 'post'
        }
      ],
      protocolList: [
        {
          value: 1,
          label: 'http'
        },
        {
          value: 2,
          label: 'shell'
        }
      ],
      levelList: [],
      dependencyStatusList: [],
      runStatusList: [],
      notifyStatusList: [],
      notifyTypes: [],
      hosts: [],
      mailUsers: [],
      slackChannels: [],
      webhookUrls: [],
      selectedMailNotifyIds: [],
      selectedSlackNotifyIds: [],
      selectedWebhookNotifyIds: [],
      tagOptions: []
    }
  },
  computed: {
    timezoneGroups () {
      try {
        const zones = Intl.supportedValuesOf('timeZone')
        const groups = { UTC: ['UTC'] }
        for (const tz of zones) {
          const region = tz.split('/')[0]
          if (!groups[region]) {
            groups[region] = []
          }
          groups[region].push(tz)
        }
        // Sort regions, put common ones first
        const priority = ['UTC', 'Asia', 'America', 'Europe', 'Pacific', 'Australia', 'Africa']
        const sorted = Object.keys(groups).sort((a, b) => {
          const ai = priority.indexOf(a)
          const bi = priority.indexOf(b)
          if (ai !== -1 && bi !== -1) return ai - bi
          if (ai !== -1) return -1
          if (bi !== -1) return 1
          return a.localeCompare(b)
        })
        return sorted.map(region => ({ label: region, zones: groups[region] }))
      } catch {
        // Fallback for browsers without Intl.supportedValuesOf
        const fallback = [
          'UTC',
          'Asia/Shanghai', 'Asia/Tokyo', 'Asia/Seoul', 'Asia/Singapore',
          'Asia/Hong_Kong', 'Asia/Kolkata', 'Asia/Dubai',
          'America/New_York', 'America/Chicago', 'America/Denver',
          'America/Los_Angeles', 'America/Sao_Paulo',
          'Europe/London', 'Europe/Paris', 'Europe/Berlin', 'Europe/Moscow',
          'Australia/Sydney', 'Australia/Perth',
          'Pacific/Auckland', 'Pacific/Honolulu'
        ]
        return [{ label: 'All', zones: fallback }]
      }
    },
    commandPlaceholder () {
      if (this.form.protocol === 1) {
        return this.t('message.pleaseEnterUrl')
      }
      return this.t('message.pleaseEnterShellCommand')
    },
    commandWarning () {
      if (!this.form.command) return ''
      if (this.form.command.includes('&quot;')) {
        return this.t('message.htmlEntityDetected') || 'HTML 实体编码已检测到，将自动转换为正确的引号'
      }
      return ''
    }
  },
  watch: {
    $route () {
      this.initializeForm()
    },
    'form.notify_status' () {
      this.updateNotifyKeywordRule()
      if (this.form.notify_status === 0) {
        this.form.notify_type = 0
      }
    },
    'form.level' () {
      this.updateSpecRule()
    }
  },
  created () {
    this.initFormRules()
    this.initSelectOptions()
    this.loadNotificationOptions()
    this.loadTagOptions()
    this.initializeForm()
  },
  methods: {
    initFormRules() {
      this.formRules = {
        name: [
          {required: true, message: this.t('message.pleaseEnterTaskName'), trigger: 'blur'}
        ],
        spec: [
          {required: true, message: this.t('message.pleaseEnterCronExpression'), trigger: 'blur'},
          {validator: (rule, value, callback) => this.validateCronSpecField(rule, value, callback), trigger: 'blur'},
          {validator: (rule, value, callback) => this.validateCronSpecField(rule, value, callback), trigger: 'change'}
        ],
        command: [
          {required: true, message: this.t('message.pleaseEnterCommand'), trigger: 'blur'}
        ],
        timeout: [
          {type: 'number', required: true, message: this.t('message.pleaseEnterValidTimeout'), trigger: 'blur'}
        ],
        retry_times: [
          {type: 'number', required: true, message: this.t('message.pleaseEnterValidRetryTimes'), trigger: 'blur'}
        ],
        retry_interval: [
          {type: 'number', required: true, message: this.t('message.pleaseEnterValidRetryInterval'), trigger: 'blur'}
        ],
        notify_keyword: [
          {required: true, message: this.t('message.pleaseEnterNotifyKeyword'), trigger: 'blur'}
        ],
        host_ids: [
          {required: true, message: this.t('message.selectTaskNode'), trigger: 'blur'},
          {validator: (rule, value, callback) => this.validateHostIds(rule, value, callback), trigger: 'change'}
        ]
      }
    },
    initSelectOptions() {
      this.levelList = [
        { value: 1, label: this.t('task.mainTask') },
        { value: 2, label: this.t('task.childTask') }
      ]
      this.dependencyStatusList = [
        { value: 1, label: this.t('task.strongDependency') },
        { value: 2, label: this.t('task.weakDependency') }
      ]
      this.runStatusList = [
        { value: 0, label: this.t('common.yes') },
        { value: 1, label: this.t('common.no') }
      ]
      this.notifyStatusList = [
        { value: 0, label: this.t('task.notifyDisabled') },
        { value: 1, label: this.t('task.notifyOnFailure') },
        { value: 2, label: this.t('task.notifyAlways') },
        { value: 3, label: this.t('task.notifyKeywordMatch') }
      ]
      this.notifyTypes = [
        { value: 0, label: this.t('task.notifyEmail') },
        { value: 1, label: this.t('task.notifySlack') },
        { value: 2, label: this.t('task.notifyWebhook') }
      ]
    },
    updateNotifyKeywordRule () {
      const keywordRules = this.formRules.notify_keyword
      const needKeyword = this.form.notify_status === 3
      if (!keywordRules || !keywordRules.length) {
        return
      }
      keywordRules[0].required = needKeyword
      if (!needKeyword) {
        this.form.notify_keyword = ''
        if (this.$refs.form) {
          this.$refs.form.clearValidate('notify_keyword')
        }
      }
      // 移除主动验证，只在用户交互时才验证
    },
    updateSpecRule () {
      const specRules = this.formRules.spec
      if (!specRules || !specRules.length) {
        return
      }
      const needSpec = this.form.level === 1
      specRules[0].required = needSpec
      if (!needSpec && this.$refs.form) {
        this.$refs.form.clearValidate('spec')
      }
      // 移除主动验证，只在用户交互时才验证
    },
    validateHostIds (rule, value, callback) {
      if (Number(this.form.protocol) === 2 && (!value || value.length === 0)) {
        callback(new Error(this.t('message.selectTaskNode')))
        return
      }
      callback()
    },
    handleProtocolChange (value, skipValidation = false) {
      const protocolValue = Number(value)
      if (Number.isNaN(protocolValue)) {
        return
      }
      this.form.protocol = protocolValue
      if (protocolValue === 2) {
        if (!skipValidation) {
          this.$nextTick(() => {
            if (this.$refs.form) {
              const p = this.$refs.form.validateField('host_ids')
              if (p && p.catch) p.catch(() => {})
            }
          })
        }
        return
      }
      this.form.host_ids = []
      this.form.host_id = ''
      this.$nextTick(() => {
        if (this.$refs.form) {
          try { this.$refs.form.clearValidate('host_ids') } catch { /* ignore */ }
        }
      })
    },
    validateCronSpecField (rule, value, callback) {
      if (this.form.level !== 1) {
        callback()
        return
      }
      const result = validateCronSpec(value)
      if (!result.valid) {
        callback(new Error(result.message))
        return
      }
      callback()
    },
    validateCommand () {
      if (this.form.command && this.form.command.includes('&quot;')) {
        // 自动修复 HTML 实体编码
        this.form.command = this.form.command
          .replace(/&quot;/g, '"')
          .replace(/&apos;/g, "'")
          .replace(/&lt;/g, '<')
          .replace(/&gt;/g, '>')
          .replace(/&amp;/g, '&')
      }
    },
    resetForm () {
      if (this.$refs.form) {
        this.$refs.form.clearValidate()
      }
      const defaults = createDefaultForm()
      Object.assign(this.form, defaults)
      this.selectedMailNotifyIds = []
      this.selectedSlackNotifyIds = []
      this.selectedWebhookNotifyIds = []
      this.handleProtocolChange(this.form.protocol, true)
      this.updateNotifyKeywordRule()
      this.updateSpecRule()
    },
    initializeForm () {
      this.resetForm()
      const id = this.$route.params.id
      if (id) {
        taskService.detail(id, (taskData, hosts) => {
          this.hosts = hosts || []
          if (!taskData) {
            this.$message.error(this.t('message.dataNotFound'))
            this.cancel()
            return
          }
          this.populateForm(taskData)
        })
        return
      }
      taskService.detail(null, (...args) => {
        const hosts = args.length > 1 ? args[1] : args[0]
        this.hosts = hosts || []
        this.handleProtocolChange(this.form.protocol, true)
        this.updateSpecRule()
      })
    },
    populateForm (taskData) {
      const { timezone, spec: cronSpec } = extractTimezone(taskData.spec)
      Object.assign(this.form, {
        id: taskData.id,
        name: taskData.name,
        tag: taskData.tag,
        tags: taskData.tag ? taskData.tag.split(',').filter(Boolean) : [],
        level: taskData.level,
        dependency_status: taskData.dependency_status || 1,
        dependency_task_id: taskData.dependency_task_id || '',
        spec: cronSpec,
        timezone: timezone,
        protocol: taskData.protocol,
        http_method: taskData.http_method || 1,
        http_body: taskData.http_body || '',
        http_headers: taskData.http_headers || '',
        success_pattern: taskData.success_pattern || '',
        command: taskData.command,
        timeout: taskData.timeout,
        multi: taskData.multi,
        notify_keyword: taskData.notify_keyword,
        notify_status: taskData.notify_status,
        notify_type: taskData.notify_type,
        notify_receiver_id: taskData.notify_receiver_id,
        retry_times: taskData.retry_times,
        retry_interval: taskData.retry_interval,
        log_retention_days: taskData.log_retention_days || 0,
        remark: taskData.remark || ''
      })
      const taskHosts = taskData.hosts || []
      this.form.host_ids = Number(this.form.protocol) === 2 ? taskHosts.map(v => v.host_id) : []
      this.handleProtocolChange(this.form.protocol, true)
      this.updateNotifyKeywordRule()
      this.updateSpecRule()


      this.selectedMailNotifyIds = []
      this.selectedSlackNotifyIds = []
      this.selectedWebhookNotifyIds = []
      if (this.form.notify_status > 0 && this.form.notify_receiver_id) {
        const notifyReceiverIds = this.form.notify_receiver_id.split(',').filter(Boolean)
        if (this.form.notify_type === 0) {
          this.selectedMailNotifyIds = notifyReceiverIds.map(v => parseInt(v))
        } else if (this.form.notify_type === 1) {
          this.selectedSlackNotifyIds = notifyReceiverIds.map(v => parseInt(v))
        } else if (this.form.notify_type === 2) {
          this.selectedWebhookNotifyIds = notifyReceiverIds.map(v => parseInt(v))
        }
      }
    },
    loadTagOptions () {
      taskService.allTags((tags) => {
        this.tagOptions = tags || []
      })
    },
    loadNotificationOptions () {
      notificationService.mail((data) => {
        this.mailUsers = data.mail_users || []
      })
      notificationService.slack((data) => {
        this.slackChannels = data.channels || []
      })
      notificationService.webhook((data) => {
        this.webhookUrls = data.webhook_urls || []
      })
    },
    submit () {
      this.$refs.form.validate().then((valid) => {
        if (!valid) {
          return false
        }
        if (this.form.notify_status > 0) {
          if (this.form.notify_type === 0 && this.selectedMailNotifyIds.length === 0) {
            this.$message.error(this.t('message.selectMailReceiver'))
            return false
          }
          if (this.form.notify_type === 1 && this.selectedSlackNotifyIds.length === 0) {
            this.$message.error(this.t('message.selectSlackChannel'))
            return false
          }
          if (this.form.notify_type === 2 && this.selectedWebhookNotifyIds.length === 0) {
            this.$message.error(this.t('message.selectWebhookUrl'))
            return false
          }
        }

        this.save()
      }).catch(() => {})
    },
    save () {
      // 构建提交用的 spec，不修改 form.spec 避免重试时双重前缀
      let specToSave = this.form.spec
      if (this.form.level === 1 && this.form.timezone && specToSave) {
        specToSave = 'CRON_TZ=' + this.form.timezone + ' ' + specToSave
      }

      // 将标签数组转换为逗号分隔的字符串
      this.form.tag = (this.form.tags || []).join(',')

      // 清理命令中的 HTML 实体编码
      let command = this.form.command || ''
      if (command) {
        command = command
          .replace(/&quot;/g, '"')
          .replace(/&apos;/g, "'")
          .replace(/&lt;/g, '<')
          .replace(/&gt;/g, '>')
          .replace(/&amp;/g, '&')
      }

      const payload = { ...this.form, spec: specToSave, command: command }

      if (Number(payload.protocol) === 2) {
        payload.host_id = this.form.host_ids.join(',')
      } else {
        payload.host_id = ''
      }
      if (payload.notify_status > 0) {
        if (payload.notify_type === 0) {
          payload.notify_receiver_id = this.selectedMailNotifyIds.join(',')
        } else if (payload.notify_type === 1) {
          payload.notify_receiver_id = this.selectedSlackNotifyIds.join(',')
        } else if (payload.notify_type === 2) {
          payload.notify_receiver_id = this.selectedWebhookNotifyIds.join(',')
        }
      } else {
        payload.notify_receiver_id = ''
      }
      taskService.update(payload, () => {
        this.$router.push('/task')
      })
    },
    cancel () {
      this.$router.push('/task')
    }
  }
}
</script>

<style scoped>
:deep(.el-form-item__error) {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
