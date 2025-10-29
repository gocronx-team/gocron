<template>
  <el-container>
    <system-sidebar></system-sidebar>
    <el-main>
      <notification-tab></notification-tab>
      <el-form ref="form" :model="form" :rules="formRules" :label-width="locale === 'zh-CN' ? '150px' : '180px'" style="width: 800px;">
        <h3>{{ t('system.emailServerConfig') }}</h3>
        <el-row>
          <el-col :span="12">
            <el-form-item :label="t('system.smtpHost')" prop="host">
              <el-input v-model="form.host"></el-input>
            </el-form-item>
          </el-col>
          <el-col :span="10">
            <el-form-item :label="t('host.port')" prop="port">
              <el-input v-model.number="form.port"></el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="12">
            <el-form-item :label="t('user.username')" prop="user">
              <el-input v-model="form.user"></el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('user.password')" prop="password">
              <el-input v-model="form.password" type="password"></el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item :label="t('system.template')" prop="template">
          <el-input
            type="textarea"
            :rows="6"
            :placeholder="emailPlaceholder"
            v-model="form.template">
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="submit()">{{ t('common.save') }}</el-button>
        </el-form-item>
        <el-button type="primary" @click="createUser">{{ t('system.addUser') }}</el-button> <br><br>
        <h3>{{ t('system.notificationUsers') }}</h3>
        <el-tag
          v-for="item in receivers"
          :key="item.email"
          closable
          @close="deleteUser(item)">
          {{item.username}} - {{item.email}}
        </el-tag>
      </el-form>
      <el-dialog
        :title="t('system.addUser')"
        v-model="dialogVisible"
        width="30%">
        <el-form :model="form">
          <el-form-item :label="t('user.username')" >
            <el-input v-model.trim="username"></el-input>
          </el-form-item>
          <el-form-item :label="t('system.emailAddress')" >
            <el-input v-model.trim="email"></el-input>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveUser">{{ t('common.confirm') }}</el-button>
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
  name: 'notification-email',
  setup() {
    const { t, locale } = useI18n()
    return { t, locale }
  },
  data () {
    return {
      form: {
        host: '',
        port: 465,
        user: '',
        password: '',
        template: ''
      },
      formRules: {},
      receivers: [],
      username: '',
      email: '',
      dialogVisible: false
    }
  },
  computed: {
    emailPlaceholder() {
      return `${this.t('system.taskIdVar')}: {{.TaskId}}
${this.t('system.taskNameVar')}: {{.TaskName}}
${this.t('system.statusVar')}: {{.Status}}
${this.t('system.resultVar')}: {{.Result}}
${this.t('task.remark')}: {{.Remark}}`
    },
    computedFormRules() {
      return {
        host: [
          {required: true, message: this.t('system.pleaseEnterEmailServer'), trigger: 'blur'}
        ],
        port: [
          {type: 'number', required: true, message: this.t('system.pleaseEnterValidPort'), trigger: 'blur'}
        ],
        user: [
          {required: true, message: this.t('system.pleaseEnterUserEmail'), trigger: 'blur'}
        ],
        password: [
          {required: true, message: this.t('user.passwordRequired'), trigger: 'blur'}
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
    createUser () {
      this.dialogVisible = true
    },
    saveUser () {
      if (this.username === '' || this.email === '') {
        this.$message.error(this.t('system.incompleteParameters'))
        return
      }
      notificationService.createMailUser({
        username: this.username,
        email: this.email
      }, () => {
        this.dialogVisible = false
        this.init()
      })
    },
    deleteUser (item) {
      notificationService.removeMailUser(item.id, () => {
        this.init()
      })
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
      notificationService.updateMail(this.form, () => {
        this.$message.success(this.t('message.updateSuccess'))
        this.init()
      })
    },
    init () {
      this.username = ''
      this.email = ''
      notificationService.mail((data) => {
        this.form.host = data.host
        if (data.port) {
          this.form.port = data.port
        }
        this.form.user = data.user
        this.form.password = data.password
        this.form.template = data.template
        this.receivers = data.mail_users
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
