<template>
  <el-main>
    <el-form ref="form" :model="form" :rules="formRules" label-width="auto">
      <!-- 基本信息 -->
      <el-row>
        <el-col :span="8">
          <el-form-item :label="t('template.name')" prop="name">
            <el-input v-model.trim="form.name" :placeholder="t('template.templateNamePlaceholder')"></el-input>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item :label="t('template.category')" prop="category">
            <el-select v-model="form.category" filterable allow-create default-first-option
              :placeholder="t('template.selectCategory')" style="width: 100%">
              <el-option value="backup" :label="t('template.category_backup')"></el-option>
              <el-option value="cleanup" :label="t('template.category_cleanup')"></el-option>
              <el-option value="monitor" :label="t('template.category_monitor')"></el-option>
              <el-option value="deploy" :label="t('template.category_deploy')"></el-option>
              <el-option value="api" :label="t('template.category_api')"></el-option>
              <el-option value="custom" :label="t('template.category_custom')"></el-option>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item :label="t('task.tag')">
            <el-input v-model="form.tag" :placeholder="t('task.tagPlaceholder')"></el-input>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row>
        <el-col :span="16">
          <el-form-item :label="t('template.description')">
            <el-input v-model="form.description" :placeholder="t('template.templateDescPlaceholder')"></el-input>
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 调度配置 -->
      <el-row>
        <el-col :span="12">
          <el-form-item :label="t('task.cronExpression')">
            <CronInput v-model="form.spec" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item :label="t('task.timezone')">
            <el-select v-model="form.timezone" filterable clearable
              :placeholder="t('task.timezoneServer')" style="width: 100%;">
              <el-option-group v-for="group in timezoneGroups" :key="group.label" :label="group.label">
                <el-option v-for="tz in group.zones" :key="tz" :label="tz" :value="tz"></el-option>
              </el-option-group>
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 执行配置 -->
      <el-row>
        <el-col :span="8">
          <el-form-item :label="t('template.protocol')">
            <el-select v-model.trim="form.protocol">
              <el-option :value="1" label="HTTP"></el-option>
              <el-option :value="2" label="Shell"></el-option>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="8" v-if="form.protocol === 1">
          <el-form-item :label="t('task.httpMethod')">
            <el-select v-model.trim="form.http_method">
              <el-option :value="1" label="GET"></el-option>
              <el-option :value="2" label="POST"></el-option>
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row>
        <el-col :span="20">
          <el-form-item :label="t('template.command')" prop="command">
            <div style="width: 100%;">
              <MonacoEditor v-model="form.command" :language="editorLanguage" height="250px" />
              <div style="color: #909399; font-size: 12px; margin-top: 4px;">
                {{ t('template.templateVarTip') }}
              </div>
            </div>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row v-if="form.protocol === 1 && form.http_method === 2">
        <el-col :span="16">
          <el-form-item :label="t('task.httpBody')">
            <el-input type="textarea" :rows="4" v-model="form.http_body"></el-input>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row v-if="form.protocol === 1">
        <el-col :span="16">
          <el-form-item :label="t('task.httpHeaders')">
            <el-input type="textarea" :rows="3" v-model="form.http_headers"></el-input>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row v-if="form.protocol === 1">
        <el-col :span="12">
          <el-form-item :label="t('task.successPattern')">
            <el-input v-model.trim="form.success_pattern" :placeholder="t('task.successPatternPlaceholder')"></el-input>
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 超时与重试 -->
      <el-row>
        <el-col :span="6">
          <el-form-item :label="t('template.timeout')">
            <el-input-number v-model="form.timeout" :min="0" :max="86400" style="width: 100%;"></el-input-number>
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item :label="t('task.singleInstance')">
            <el-select v-model.trim="form.multi" style="width: 100%;">
              <el-option :value="0" :label="t('common.yes')"></el-option>
              <el-option :value="1" :label="t('common.no')"></el-option>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item :label="t('task.retryTimes')">
            <el-input-number v-model="form.retry_times" :min="0" :max="10" style="width: 100%;"></el-input-number>
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item :label="t('task.retryInterval')">
            <el-input-number v-model="form.retry_interval" :min="0" :max="3600" style="width: 100%;"></el-input-number>
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 通知策略 -->
      <el-row>
        <el-col :span="8">
          <el-form-item :label="t('task.notification')">
            <el-select v-model.trim="form.notify_status" style="width: 100%;">
              <el-option :value="0" :label="t('task.notifyDisabled')"></el-option>
              <el-option :value="1" :label="t('task.notifyOnFailure')"></el-option>
              <el-option :value="2" :label="t('task.notifyAlways')"></el-option>
              <el-option :value="3" :label="t('task.notifyKeywordMatch')"></el-option>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="8" v-if="form.notify_status !== 0">
          <el-form-item :label="t('task.notifyType')">
            <el-select v-model.trim="form.notify_type" style="width: 100%;">
              <el-option :value="0" :label="t('task.notifyEmail')"></el-option>
              <el-option :value="1" :label="t('task.notifySlack')"></el-option>
              <el-option :value="2" :label="t('task.notifyWebhook')"></el-option>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="8" v-if="form.notify_status === 3">
          <el-form-item :label="t('task.notifyKeyword')">
            <el-input v-model.trim="form.notify_keyword" :placeholder="t('task.notifyKeywordPlaceholder')"></el-input>
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 日志保留 -->
      <el-row>
        <el-col :span="12">
          <el-form-item :label="t('task.logRetentionDays')">
            <el-input-number v-model="form.log_retention_days" :min="0" :max="3650"></el-input-number>
            <span style="margin-left: 8px; color: #909399; font-size: 12px;">{{ t('task.logRetentionDaysTip') }}</span>
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
import templateService from '../../api/template'
import MonacoEditor from '../../components/common/MonacoEditor.vue'
import CronInput from '../../components/common/CronInput.vue'

export default {
  name: 'template-edit',
  components: { MonacoEditor, CronInput },
  setup() {
    const { t } = useI18n()
    return { t }
  },
  computed: {
    editorLanguage() {
      return this.form.protocol === 1 ? 'plaintext' : 'shell'
    },
    timezoneGroups() {
      try {
        const zones = Intl.supportedValuesOf('timeZone')
        const groups = { UTC: ['UTC'] }
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
        return sorted.map(region => ({ label: region, zones: groups[region] }))
      } catch {
        return [{ label: 'All', zones: ['UTC', 'Asia/Shanghai', 'America/New_York', 'Europe/London'] }]
      }
    }
  },
  data() {
    return {
      form: {
        id: '',
        name: '',
        description: '',
        category: 'custom',
        protocol: 2,
        command: '',
        http_method: 1,
        http_body: '',
        http_headers: '',
        success_pattern: '',
        tag: '',
        spec: '',
        timeout: 300,
        multi: 1,
        retry_times: 0,
        retry_interval: 0,
        timezone: '',
        notify_status: 0,
        notify_type: 0,
        notify_keyword: '',
        log_retention_days: 0
      },
      formRules: {
        name: [
          { required: true, message: '', trigger: 'blur' }
        ],
        category: [
          { required: true, message: '', trigger: 'blur' }
        ],
        command: [
          { required: true, message: '', trigger: 'blur' }
        ]
      }
    }
  },
  watch: {
    $route() {
      this.loadForm()
    }
  },
  created() {
    this.formRules.name[0].message = this.t('template.templateNamePlaceholder')
    this.formRules.category[0].message = this.t('template.selectCategory')
    this.formRules.command[0].message = this.t('message.pleaseEnterCommand')
    this.loadForm()
  },
  methods: {
    loadForm() {
      // 重置表单
      this.form = {
        id: '',
        name: '',
        description: '',
        category: 'custom',
        protocol: 2,
        command: '',
        http_method: 1,
        http_body: '',
        http_headers: '',
        success_pattern: '',
        tag: '',
        spec: '',
        timeout: 300,
        multi: 1,
        retry_times: 0,
        retry_interval: 0,
        timezone: '',
        notify_status: 0,
        notify_type: 0,
        notify_keyword: '',
        log_retention_days: 0
      }
      if (this.$refs.form) {
        this.$refs.form.clearValidate()
      }

      const id = this.$route.params.id
      if (id) {
        templateService.detail(id, (data) => {
          if (data) {
            this.form = {
              id: data.id,
              name: data.name,
              description: data.description || '',
              category: data.category,
              protocol: data.protocol,
              command: data.command,
              http_method: data.http_method || 1,
              http_body: data.http_body || '',
              http_headers: data.http_headers || '',
              success_pattern: data.success_pattern || '',
              tag: data.tag || '',
              spec: data.spec || '',
              timeout: data.timeout || 300,
              multi: data.multi ?? 1,
              retry_times: data.retry_times || 0,
              retry_interval: data.retry_interval || 0,
              timezone: data.timezone || '',
              notify_status: data.notify_status || 0,
              notify_type: data.notify_type || 0,
              notify_keyword: data.notify_keyword || '',
              log_retention_days: data.log_retention_days || 0
            }
          }
        })
      }
    },
    submit() {
      this.$refs.form.validate().then((valid) => {
        if (!valid) return false
        templateService.store(this.form, () => {
          this.$router.push('/template')
        })
      }).catch(() => {})
    },
    cancel() {
      this.$router.push('/template')
    }
  }
}
</script>
