<template>
  <el-container>
    <system-sidebar></system-sidebar>
    <el-main>
      <notification-tab></notification-tab>
      <el-form ref="form" :model="form" :rules="formRules" :label-width="locale === 'zh-CN' ? '100px' : '120px'" style="width: 700px;">
        <el-form-item label="URL" prop="url">
          <el-input v-model.trim="form.url"></el-input>
        </el-form-item>
        <el-alert
          :title="t('system.webhookTip')"
          type="info"
          :closable="false"
          style="margin-bottom: 15px;">
        </el-alert>
        <el-form-item :label="t('system.template')" prop="template">
          <el-input
            type="textarea"
            :rows="8"
            :placeholder="webhookPlaceholder"
            v-model.trim="form.template">
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="submit()">{{ t('common.save') }}</el-button>
        </el-form-item>
      </el-form>
    </el-main>
  </el-container>
</template>

<script>
import { useI18n } from 'vue-i18n'
import systemSidebar from '../sidebar.vue'
import notificationTab from './tab.vue'
import notificationService from '../../../api/notification'
export default {
  name: 'notification-webhook',
  setup() {
    const { t, locale } = useI18n()
    return { t, locale }
  },
  data () {
    return {
      form: {
        url: '',
        template: ''
      },
      formRules: {}
    }
  },
  computed: {
    webhookPlaceholder() {
      return `{"task_id": "{{.TaskId}}", "task_name": "{{.TaskName}}", "status": "{{.Status}}", "result": "{{.Result}}", "remark": "{{.Remark}}"}`
    },
    computedFormRules() {
      return {
        url: [
          {type: 'url', required: true, message: this.t('system.pleaseEnterValidUrl'), trigger: 'blur'}
        ],
        template: [
          {required: true, message: this.t('system.pleaseEnterTemplate'), trigger: 'blur'}
        ]
      }
    }
  },
  watch: {
    computedFormRules: {
      handler(newVal) {
        this.formRules = newVal
      },
      immediate: true
    }
  },
  components: {notificationTab, systemSidebar},
  created () {
    this.init()
  },
  methods: {
    submit () {
      this.$refs['form'].validate((valid) => {
        if (!valid) {
          return false
        }
        this.save()
      })
    },
    save () {
      notificationService.updateWebHook(this.form, () => {
        this.$message.success(this.t('message.updateSuccess'))
        this.init()
      })
    },
    init () {
      notificationService.webhook((data) => {
        this.form.url = data.url
        this.form.template = data.template
      })
    }
  }
}
</script>
