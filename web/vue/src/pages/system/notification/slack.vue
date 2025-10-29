<template>
  <el-container>
    <system-sidebar></system-sidebar>
    <el-main>
      <notification-tab></notification-tab>
      <el-form ref="form" :model="form" :rules="formRules" :label-width="locale === 'zh-CN' ? '180px' : '220px'" style="width: 700px;">
        <el-form-item :label="t('system.slackUrl')" prop="url">
          <el-input v-model="form.url"></el-input>
        </el-form-item>
        <el-form-item :label="t('system.template')" prop="template">
          <el-input
            type="textarea"
            :rows="8"
            :placeholder="slackPlaceholder"
            v-model="form.template">
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="submit">{{ t('common.save') }}</el-button>
        </el-form-item>
        <h3>{{ t('system.channels') }}</h3>
        <el-button type="primary" @click="createChannel">{{ t('system.addChannel') }}</el-button> <br><br>
        <el-tag
          v-for="item in channels"
          :key="item.id"
          closable
          @close="deleteChannel(item)"
        >
          {{item.name}}
        </el-tag>
      </el-form>
      <el-dialog
        :title="t('system.addChannel')"
        v-model="dialogVisible"
        width="30%">
        <el-form :model="form">
          <el-form-item :label="t('system.channelName')" >
            <el-input v-model.trim="channel" v-focus></el-input>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveChannel">{{ t('common.confirm') }}</el-button>
          </el-form-item>
        </el-form>
      </el-dialog>
    </el-main>
  </el-container>
</template>

<script>
import { useI18n } from 'vue-i18n'
import systemSidebar from '../sidebar.vue'
import notificationTab from './tab.vue'
import notificationService from '../../../api/notification'
export default {
  name: 'notification-slack',
  setup() {
    const { t, locale } = useI18n()
    return { t, locale }
  },
  data () {
    return {
      dialogVisible: false,
      form: {
        url: '',
        template: ''
      },
      formRules: {},
      channels: [],
      channel: ''
    }
  },
  computed: {
    slackPlaceholder() {
      return `${this.t('system.taskIdVar')}: {{.TaskId}}
${this.t('system.taskNameVar')}: {{.TaskName}}
${this.t('system.statusVar')}: {{.Status}}
${this.t('system.resultVar')}: {{.Result}}
${this.t('task.remark')}: {{.Remark}}`
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
    createChannel () {
      this.dialogVisible = true
    },
    submit () {
      this.$refs['form'].validate((valid) => {
        if (!valid) {
          return false
        }
        this.save()
      })
    },
    save () {
      notificationService.updateSlack(this.form, () => {
        this.$message.success(this.t('message.updateSuccess'))
        this.init()
      })
    },
    saveChannel () {
      if (this.channel === '') {
        this.$message.error(this.t('system.pleaseEnterChannelName'))
        return
      }
      notificationService.createSlackChannel(this.channel, () => {
        this.dialogVisible = false
        this.init()
      })
    },
    deleteChannel (item) {
      notificationService.removeSlackChannel(item.id, () => {
        this.init()
      })
    },
    init () {
      this.channel = ''
      notificationService.slack((data) => {
        this.form.url = data.url
        this.form.template = data.template
        this.channels = data.channels
      })
    }
  }
}
</script>

<style scoped>
  .el-tag + .el-tag {
    margin-left: 10px;
  }
</style>
