<template>
  <div class="login-container">
    <div class="login-box">
      <div class="language-switcher">
        <LanguageSwitcher />
      </div>
      <h2 class="login-title">{{ t('login.title') }}</h2>
      <el-alert
        v-if="errorMessage"
        :title="errorMessage"
        type="error"
        :closable="false"
        style="margin-bottom: 20px;"
      />
      <el-form ref="formRef" :model="form" label-width="120px" :rules="formRules">
        <el-form-item :label="t('login.username')" prop="username">
          <el-input v-model.trim="form.username" :placeholder="t('login.usernamePlaceholder')" size="large" />
        </el-form-item>
        <el-form-item :label="t('login.password')" prop="password">
          <el-input v-model.trim="form.password" type="password" :placeholder="t('login.passwordPlaceholder')" @keyup.enter="submit" size="large" />
        </el-form-item>
        <el-form-item :label="t('login.verifyCode')" prop="twoFactorCode" v-if="require2FA">
          <el-input v-model.trim="form.twoFactorCode" :placeholder="t('login.verifyCodePlaceholder')" maxlength="6" @keyup.enter="submit" size="large" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="submit" :loading="loading" class="login-button" size="large">{{ t('login.login') }}</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
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


<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  position: relative;
  overflow: hidden;
}

.login-container::before {
  content: '';
  position: absolute;
  top: -50%;
  right: -10%;
  width: 600px;
  height: 600px;
  background: rgba(99, 102, 241, 0.1);
  border-radius: 50%;
  filter: blur(80px);
}

.login-container::after {
  content: '';
  position: absolute;
  bottom: -30%;
  left: -10%;
  width: 500px;
  height: 500px;
  background: rgba(168, 85, 247, 0.08);
  border-radius: 50%;
  filter: blur(80px);
}

.login-box {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  padding: 48px 40px;
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
  width: 100%;
  max-width: 420px;
  position: relative;
  z-index: 1;
  border: 1px solid rgba(255, 255, 255, 0.8);
}

.language-switcher {
  position: absolute;
  top: 16px;
  left: 16px;
}

.login-title {
  text-align: center;
  margin: 0 0 32px 0;
  font-size: 26px;
  color: #1f2937;
  font-weight: 600;
  letter-spacing: -0.5px;
}

.el-button--large {
  height: 40px;
  line-height: 40px;
  padding: 0 15px;
}

.login-button {
  width: calc(100% + 120px);
  margin-left: -90px;
  margin-right: -20px;
}
</style>
