<template>
  <el-dialog
    :title="t('login.title')"
    v-model="dialogVisible"
    :close-on-click-modal="false"
    :show-close="false"
    :close-on-press-escape="false"
    width="450px">
    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      :closable="false"
      style="margin-bottom: 20px;"
    />
    <el-form ref="formRef" :model="form" :label-width="locale === 'zh-CN' ? '80px' : '150px'" :rules="formRules">
      <el-form-item :label="t('login.username')" prop="username">
        <el-input v-model.trim="form.username" :placeholder="t('login.usernamePlaceholder')" />
      </el-form-item>
      <el-form-item :label="t('login.password')" prop="password">
        <el-input v-model.trim="form.password" type="password" :placeholder="t('login.passwordPlaceholder')" @keyup.enter="submit" />
      </el-form-item>
      <el-form-item :label="t('login.verifyCode')" prop="twoFactorCode" v-if="require2FA">
        <el-input v-model.trim="form.twoFactorCode" :placeholder="t('login.verifyCodePlaceholder')" maxlength="6" @keyup.enter="submit" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="submit" :loading="loading" style="width: 100%;">{{ t('login.login') }}</el-button>
      </el-form-item>
    </el-form>
    <div style="text-align: center; margin-top: 10px;">
      <LanguageSwitcher />
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '../../stores/user'
import { useLoading } from '../../composables/useLoading'
import userService from '../../api/user'
import LanguageSwitcher from '../../components/common/LanguageSwitcher.vue'

const { t, locale } = useI18n()

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const { loading, withLoading } = useLoading()

const dialogVisible = ref(true)
const require2FA = ref(false)
const formRef = ref()
const errorMessage = ref('')

const form = reactive({
  username: '',
  password: '',
  twoFactorCode: ''
})

const formRules = computed(() => ({
  username: [{ required: true, message: t('login.usernameRequired'), trigger: 'blur' }],
  password: [{ required: true, message: t('login.passwordRequired'), trigger: 'blur' }],
  twoFactorCode: [{ required: true, message: t('login.verifyCodeRequired'), trigger: 'blur' }]
}))

const submit = async () => {
  if (!formRef.value) return
  
  errorMessage.value = ''
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    if (require2FA.value && !form.twoFactorCode) {
      errorMessage.value = t('login.verifyCodeRequired')
      return
    }
    
    await withLoading(async () => {
      const params = {
        username: form.username,
        password: form.password,
        two_factor_code: form.twoFactorCode || undefined
      }
      
      userService.login(
        params.username, 
        params.password, 
        params.two_factor_code, 
        (data) => {
          if (data.require_2fa) {
            require2FA.value = true
            errorMessage.value = ''
            return
          }
          
          userStore.setUser({
            token: data.token,
            uid: data.uid,
            username: data.username,
            isAdmin: data.is_admin
          })
          
          router.push(route.query.redirect || '/')
        },
        (code, message) => {
          errorMessage.value = message || '登录失败'
        }
      )
    })
  })
}
</script>
