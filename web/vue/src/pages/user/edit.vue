<template>
  <el-container>
    <el-main>
      <el-form ref="form" :model="form" :rules="formRules" :label-width="locale === 'zh-CN' ? '100px' : '180px'" style="width: 500px;">
        <el-form-item>
          <el-input v-model="form.id" type="hidden"></el-input>
        </el-form-item>
        <el-form-item :label="t('user.username')" prop="name">
          <el-input v-model="form.name"></el-input>
        </el-form-item>
        <el-form-item :label="t('user.email')" prop="email">
          <el-input v-model="form.email"></el-input>
        </el-form-item>
        <template v-if="!form.id">
          <el-form-item :label="t('user.password')" prop="password">
            <el-input v-model="form.password" type="password"></el-input>
          </el-form-item>
          <el-form-item :label="t('user.confirmPassword')" prop="confirm_password">
            <el-input v-model="form.confirm_password" type="password"></el-input>
          </el-form-item>
        </template>
        <el-form-item :label="t('user.role')" prop="is_admin">
          <el-radio-group v-model="form.is_admin">
            <el-radio :label="0">{{ t('user.normalUser') }}</el-radio>
            <el-radio :label="1">{{ t('user.admin') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="t('common.status')" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio :label="1">{{ t('common.enabled') }}</el-radio>
            <el-radio :label="0">{{ t('common.disabled') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="submit()">{{ t('common.save') }}</el-button>
          <el-button @click="cancel">{{ t('common.cancel') }}</el-button>
        </el-form-item>
      </el-form>
    </el-main>
  </el-container>
</template>

<script>
import { useI18n } from 'vue-i18n'
import userService from '../../api/user'
export default {
  name: 'user-edit',
  setup() {
    const { t, locale } = useI18n()
    return { t, locale }
  },
  data: function () {
    return {
      form: {
        id: '',
        name: '',
        email: '',
        is_admin: 0,
        password: '',
        confirm_password: '',
        status: 1
      },
      formRules: {}
    }
  },
  computed: {
    computedFormRules() {
      return {
        name: [
          {required: true, message: this.t('user.usernameRequired'), trigger: 'blur'}
        ],
        email: [
          {type: 'email', required: true, message: this.t('user.emailRequired'), trigger: 'blur'}
        ],
        password: [
          {required: true, message: this.t('user.passwordRequired'), trigger: 'blur'}
        ],
        confirm_password: [
          {required: true, message: this.t('user.confirmPasswordRequired'), trigger: 'blur'}
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
  created () {
    const id = this.$route.params.id
    if (!id) {
      return
    }
    userService.detail(id, (data) => {
      if (!data) {
        this.$message.error(this.t('message.dataNotFound'))
        return
      }
      this.form.id = data.id
      this.form.name = data.name
      this.form.email = data.email
      this.form.is_admin = data.is_admin
      this.form.status = data.status
    })
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
      userService.update(this.form, () => {
        this.$router.push('/user')
      })
    },
    cancel () {
      this.$router.push('/user')
    }
  }
}
</script>
