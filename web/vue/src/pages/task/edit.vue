<template>
  <el-container >
    <task-sidebar></task-sidebar>
    <el-main>
      <el-form ref="form" :model="form" :rules="formRules" :label-width="locale === 'zh-CN' ? '180px' : '220px'">
        <el-input v-model="form.id" type="hidden"></el-input>
        <el-row>
          <el-col :span="12">
            <el-form-item :label="t('task.name')" prop="name">
              <el-input v-model.trim="form.name"></el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('task.tag')">
              <el-input v-model.trim="form.tag" :placeholder="t('task.tagPlaceholder')"></el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row v-if="form.level === 1">
          <el-col>
            <el-alert
              :title="t('task.mainTaskTip')"
              type="info"
              :closable="false">
            </el-alert>
            <el-alert
              :title="t('task.dependencyTip')"
              type="info"
              :closable="false">
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
                v-model="form.command">
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
          <el-col :span="8" v-if="form.notify_status !== 1">
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
                  v-if="form.notify_status !== 1 && form.notify_type === 2">
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
                  v-if="form.notify_status !== 1 && form.notify_type === 3">
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
        </el-row>
        <el-row v-if="form.notify_status === 4">
          <el-col :span="12">
            <el-form-item :label="t('task.notifyKeyword')" prop="notify_keyword">
              <el-input v-model.trim="form.notify_keyword" :placeholder="t('task.notifyKeywordPlaceholder')"></el-input>
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
  </el-container>
</template>


<script>
import { useI18n } from 'vue-i18n'
import taskSidebar from './sidebar.vue'
import taskService from '../../api/task'
import notificationService from '../../api/notification'
import { validateCronSpec, getCronExamples } from '../../utils/cronValidator'

const createDefaultForm = () => ({
  id: '',
  name: '',
  tag: '',
  level: 1,
  dependency_status: 1,
  dependency_task_id: '',
  spec: '',
  protocol: 2,
  http_method: 1,
  command: '',
  host_id: '',
  host_ids: [],
  timeout: 0,
  multi: 2,
  notify_status: 1,
  notify_type: 2,
  notify_receiver_id: '',
  notify_keyword: '',
  retry_times: 0,
  retry_interval: 0,
  remark: ''
})

export default {
  name: 'task-edit',
  components: {taskSidebar},
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
      selectedMailNotifyIds: [],
      selectedSlackNotifyIds: []
    }
  },
  computed: {
    commandPlaceholder () {
      if (this.form.protocol === 1) {
        return this.t('message.pleaseEnterUrl')
      }
      return this.t('message.pleaseEnterShellCommand')
    }
  },
  watch: {
    $route () {
      this.initializeForm()
    },
    'form.notify_status' () {
      this.updateNotifyKeywordRule()
    },
    'form.level' () {
      this.updateSpecRule()
    }
  },
  created () {
    this.initFormRules()
    this.initSelectOptions()
    this.loadNotificationOptions()
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
        { value: 2, label: this.t('common.yes') },
        { value: 1, label: this.t('common.no') }
      ]
      this.notifyStatusList = [
        { value: 1, label: this.t('task.notifyDisabled') },
        { value: 2, label: this.t('task.notifyOnFailure') },
        { value: 3, label: this.t('task.notifyAlways') },
        { value: 4, label: this.t('task.notifyKeywordMatch') }
      ]
      this.notifyTypes = [
        { value: 2, label: this.t('task.notifyEmail') },
        { value: 3, label: this.t('task.notifySlack') },
        { value: 4, label: this.t('task.notifyWebhook') }
      ]
    },
    updateNotifyKeywordRule () {
      const keywordRules = this.formRules.notify_keyword
      const needKeyword = this.form.notify_status === 4
      if (!keywordRules || !keywordRules.length) {
        return
      }
      keywordRules[0].required = needKeyword
      if (!needKeyword) {
        this.form.notify_keyword = ''
        if (this.$refs.form) {
          this.$refs.form.clearValidate('notify_keyword')
        }
      } else if (this.$refs.form) {
        this.$nextTick(() => {
          this.$refs.form && this.$refs.form.validateField('notify_keyword')
        })
      }
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
      } else if (needSpec && this.$refs.form) {
        this.$nextTick(() => {
          this.$refs.form && this.$refs.form.validateField('spec')
        })
      }
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
            this.$refs.form && this.$refs.form.validateField('host_ids')
          })
        }
        return
      }
      this.form.host_ids = []
      this.form.host_id = ''
      if (this.$refs.form) {
        this.$refs.form.clearValidate('host_ids')
      }
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
    resetForm () {
      if (this.$refs.form) {
        this.$refs.form.clearValidate()
      }
      const defaults = createDefaultForm()
      Object.assign(this.form, defaults)
      this.selectedMailNotifyIds = []
      this.selectedSlackNotifyIds = []
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
      Object.assign(this.form, {
        id: taskData.id,
        name: taskData.name,
        tag: taskData.tag,
        level: taskData.level,
        dependency_status: taskData.dependency_status || 1,
        dependency_task_id: taskData.dependency_task_id || '',
        spec: taskData.spec,
        protocol: taskData.protocol,
        http_method: taskData.http_method || 1,
        command: taskData.command,
        timeout: taskData.timeout,
        multi: taskData.multi ? 1 : 2,
        notify_keyword: taskData.notify_keyword,
        notify_status: taskData.notify_status + 1,
        notify_type: taskData.notify_type ? taskData.notify_type + 1 : 2,
        notify_receiver_id: taskData.notify_receiver_id,
        retry_times: taskData.retry_times,
        retry_interval: taskData.retry_interval,
        remark: taskData.remark || ''
      })
      const taskHosts = taskData.hosts || []
      this.form.host_ids = Number(this.form.protocol) === 2 ? taskHosts.map(v => v.host_id) : []
      this.handleProtocolChange(this.form.protocol, true)
      this.updateNotifyKeywordRule()
      this.updateSpecRule()

      this.selectedMailNotifyIds = []
      this.selectedSlackNotifyIds = []
      if (this.form.notify_status > 1 && this.form.notify_receiver_id) {
        const notifyReceiverIds = this.form.notify_receiver_id.split(',').filter(Boolean)
        if (this.form.notify_type === 2) {
          this.selectedMailNotifyIds = notifyReceiverIds.map(v => parseInt(v))
        } else if (this.form.notify_type === 3) {
          this.selectedSlackNotifyIds = notifyReceiverIds.map(v => parseInt(v))
        }
      }
    },
    loadNotificationOptions () {
      if (!this.mailUsers.length) {
        notificationService.mail((data) => {
          this.mailUsers = data.mail_users
        })
      }
      if (!this.slackChannels.length) {
        notificationService.slack((data) => {
          this.slackChannels = data.channels
        })
      }
    },
    submit () {
      this.$refs.form.validate((valid) => {
        if (!valid) {
          return false
        }
        if (this.form.notify_status > 1) {
          if (this.form.notify_type === 2 && this.selectedMailNotifyIds.length === 0) {
            this.$message.error(this.t('message.selectMailReceiver'))
            return false
          }
          if (this.form.notify_type === 3 && this.selectedSlackNotifyIds.length === 0) {
            this.$message.error(this.t('message.selectSlackChannel'))
            return false
          }
        }

        this.save()
      })
    },
    save () {
      if (Number(this.form.protocol) === 2) {
        this.form.host_id = this.form.host_ids.join(',')
      } else {
        this.form.host_id = ''
        this.form.host_ids = []
      }
      if (this.form.notify_status > 1 && this.form.notify_type === 2) {
        this.form.notify_receiver_id = this.selectedMailNotifyIds.join(',')
      }
      if (this.form.notify_status > 1 && this.form.notify_type === 3) {
        this.form.notify_receiver_id = this.selectedSlackNotifyIds.join(',')
      }
      taskService.update(this.form, () => {
        this.$router.push('/task')
      })
    },
    cancel () {
      this.$router.push('/task')
    }
  }
}
</script>
