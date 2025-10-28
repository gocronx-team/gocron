<template>
  <el-container >
    <task-sidebar></task-sidebar>
    <el-main>
      <el-form ref="form" :model="form" :rules="formRules" label-width="180px">
        <el-input v-model="form.id" type="hidden"></el-input>
        <el-row>
          <el-col :span="12">
            <el-form-item label="任务名称" prop="name">
              <el-input v-model.trim="form.name"></el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="标签">
              <el-input v-model.trim="form.tag" placeholder="通过标签将任务分组"></el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row v-if="form.level === 1">
          <el-col>
            <el-alert
              title="主任务可以配置多个子任务, 当主任务执行完成后，自动执行子任务
任务类型新增后不能变更"
              type="info"
              :closable="false">
            </el-alert>
            <el-alert
              title="强依赖: 主任务执行成功，才会运行子任务
弱依赖: 无论主任务执行是否成功，都会运行子任务"
              type="info"
              :closable="false">
            </el-alert> <br>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="7">
            <el-form-item label="任务类型">
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
            <el-form-item label="依赖关系">
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
            <el-form-item label="子任务ID" v-if="form.level === 1">
              <el-input v-model.trim="form.dependency_task_id" placeholder="多个ID逗号分隔"></el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row v-if="form.level === 1">
          <el-col :span="12">
            <el-form-item label="crontab表达式" prop="spec">
              <el-input v-model.trim="form.spec"
                        placeholder="秒 分 时 天 月 周"></el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="8">
            <el-form-item label="执行方式">
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
            <el-form-item label="请求方法">
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
            <el-form-item label="任务节点" prop="host_ids">
              <el-select
                key="shell"
                v-model="form.host_ids"
                filterable
                multiple
                placeholder="请选择任务节点">
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
            <el-form-item label="命令" prop="command">
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
              title="任务执行超时强制结束, 取值0-86400(秒), 默认0, 不限制"
              type="info"
              :closable="false">
            </el-alert>
            <el-alert
              title="单实例运行, 前次任务未执行完成，下次任务调度时间到了是否要执行, 即是否允许多进程执行同一任务"
              type="info"
              :closable="false">
            </el-alert> <br>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item label="任务超时时间" prop="timeout">
              <el-input v-model.number.trim="form.timeout"></el-input>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="单实例运行">
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
          <el-form-item label="任务失败重试次数" prop="retry_times">
            <el-input v-model.number.trim="form.retry_times"
                      placeholder="0 - 10, 默认0，不重试"></el-input>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="任务失败重试间隔时间" prop="retry_interval">
            <el-input v-model.number.trim="form.retry_interval" placeholder="0 - 3600 (秒), 默认0，执行系统默认策略"></el-input>
          </el-form-item>
        </el-col>
        </el-row>
        <el-row>
          <el-col :span="8">
            <el-form-item label="任务通知">
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
            <el-form-item label="通知类型">
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
            <el-form-item label="接收用户">
              <el-select key="notify-mail" v-model="selectedMailNotifyIds" filterable multiple placeholder="请选择">
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
            <el-form-item label="发送Channel">
              <el-select key="notify-slack" v-model="selectedSlackNotifyIds" filterable multiple placeholder="请选择">
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
            <el-form-item label="任务执行输出关键字" prop="notify_keyword">
              <el-input v-model.trim="form.notify_keyword" placeholder="任务执行输出中包含此关键字将触发通知"></el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="16">
            <el-form-item label="备注">
              <el-input
                type="textarea"
                :rows="3"
                v-model="form.remark">
              </el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item>
          <el-button type="primary" @click="submit">保存</el-button>
          <el-button @click="cancel">取消</el-button>
        </el-form-item>
      </el-form>
    </el-main>
  </el-container>
</template>


<script>
import taskSidebar from './sidebar.vue'
import taskService from '../../api/task'
import notificationService from '../../api/notification'

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
  data () {
    return {
      form: createDefaultForm(),
      formRules: {
        name: [
          {required: true, message: '请输入任务名称', trigger: 'blur'}
        ],
        spec: [
          {required: true, message: '请输入crontab表达式', trigger: 'blur'},
          {validator: () => true, trigger: 'blur'},
          {validator: () => true, trigger: 'change'}
        ],
        command: [
          {required: true, message: '请输入命令', trigger: 'blur'}
        ],
        timeout: [
          {type: 'number', required: true, message: '请输入有效的任务超时时间', trigger: 'blur'}
        ],
        retry_times: [
          {type: 'number', required: true, message: '请输入有效的任务执行失败重试次数', trigger: 'blur'}
        ],
        retry_interval: [
          {type: 'number', required: true, message: '请输入有效的任务执行失败，重试间隔时间', trigger: 'blur'}
        ],
        notify_keyword: [
          {required: true, message: '请输入要匹配的任务执行输出关键字', trigger: 'blur'}
        ],
        host_ids: [
          {validator: () => true, trigger: 'change'}
        ]
      },
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
      levelList: [
        {
          value: 1,
          label: '主任务'
        },
        {
          value: 2,
          label: '子任务'
        }
      ],
      dependencyStatusList: [
        {
          value: 1,
          label: '强依赖'
        },
        {
          value: 2,
          label: '弱依赖'
        }
      ],
      runStatusList: [
        {
          value: 2,
          label: '是'
        },
        {
          value: 1,
          label: '否'
        }
      ],
      notifyStatusList: [
        {
          value: 1,
          label: '不通知'
        },
        {
          value: 2,
          label: '失败通知'
        },
        {
          value: 3,
          label: '总是通知'
        },
        {
          value: 4,
          label: '关键字匹配通知'
        }
      ],
      notifyTypes: [
        {
          value: 2,
          label: '邮件'
        },
        {
          value: 3,
          label: 'Slack'
        },
        {
          value: 4,
          label: 'WebHook'
        }
      ],
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
        return '请输入URL地址'
      }

      return '请输入shell命令'
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
    this.formRules.host_ids[0].validator = (rule, value, callback) => this.validateHostIds(rule, value, callback)
    this.formRules.spec[1].validator = (rule, value, callback) => this.validateCronSpecField(rule, value, callback)
    this.formRules.spec[2].validator = (rule, value, callback) => this.validateCronSpecField(rule, value, callback)
    this.loadNotificationOptions()
    this.initializeForm()
  },
  methods: {
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
        callback(new Error('请选择任务节点'))
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
      const specValue = value ? value.trim() : ''
      if (!specValue) {
        callback(new Error('请输入crontab表达式'))
        return
      }
      const segments = specValue.split(/\s+/)
      if (segments.length !== 6) {
        callback(new Error('crontab表达式需包含6段（秒 分 时 日 月 周）'))
        return
      }
      const allowedPattern = /^[0-9A-Za-z*\/,\-?#]+$/
      const numericRanges = [
        { min: 0, max: 59 },
        { min: 0, max: 59 },
        { min: 0, max: 23 },
        { min: 1, max: 31 },
        { min: 1, max: 12 },
        { min: 0, max: 7 }
      ]
      const invalid = segments.some((part, index) => {
        if (!allowedPattern.test(part)) {
          return true
        }
        if (/^\d+$/.test(part)) {
          const num = Number(part)
          const { min, max } = numericRanges[index]
          return Number.isNaN(num) || num < min || num > max
        }
        return false
      })
      if (invalid) {
        callback(new Error('crontab表达式格式不正确，请检查输入'))
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
            this.$message.error('数据不存在')
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
            this.$message.error('请选择邮件接收用户')
            return false
          }
          if (this.form.notify_type === 3 && this.selectedSlackNotifyIds.length === 0) {
            this.$message.error('请选择Slack Channel')
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
