<!-- 登录页面 -->
<template>
  <div class="flex w-full h-screen">
    <LoginLeftView />

    <div class="relative flex-1">
      <AuthTopBar />

      <div class="auth-right-wrap">
        <div class="form">
          <h3 class="title">{{ $t('login.title') }}</h3>
          <p class="sub-title">{{ $t('login.subTitle') }}</p>

          <!-- Step 1: username + password -->
          <ElForm
            v-if="!require2FA"
            ref="formRef"
            :model="formData"
            :rules="rules"
            :key="formKey"
            @keyup.enter="handleSubmit"
            style="margin-top: 25px"
          >
            <ElFormItem prop="username">
              <ElInput
                class="custom-height"
                :placeholder="$t('login.placeholder.username')"
                v-model.trim="formData.username"
                autocomplete="username"
              />
            </ElFormItem>
            <ElFormItem prop="password">
              <ElInput
                class="custom-height"
                :placeholder="$t('login.placeholder.password')"
                v-model.trim="formData.password"
                type="password"
                autocomplete="current-password"
                show-password
              />
            </ElFormItem>

            <div style="margin-top: 30px">
              <ElButton
                class="w-full custom-height"
                type="primary"
                @click="handleSubmit"
                :loading="loading"
                v-ripple
              >
                {{ $t('login.btnText') }}
              </ElButton>
            </div>
          </ElForm>

          <!-- Step 2: 2FA code -->
          <ElForm
            v-else
            ref="form2FARef"
            :model="formData"
            :rules="rules2FA"
            :key="'2fa-' + formKey"
            @keyup.enter="handleSubmit"
            style="margin-top: 25px"
          >
            <p class="sub-title" style="margin-bottom: 16px">
              {{ $t('login.placeholder.twoFactorCode') }}
            </p>
            <ElFormItem prop="two_factor_code">
              <ElInput
                class="custom-height"
                :placeholder="$t('login.placeholder.twoFactorCode')"
                v-model.trim="formData.two_factor_code"
                maxlength="6"
                autocomplete="one-time-code"
              />
            </ElFormItem>

            <div style="margin-top: 30px">
              <ElButton
                class="w-full custom-height"
                type="primary"
                @click="handleSubmit"
                :loading="loading"
                v-ripple
              >
                {{ $t('login.btnText') }}
              </ElButton>
            </div>

            <div class="mt-4">
              <ElButton
                class="w-full"
                @click="cancelTwoFactor"
                :disabled="loading"
                link
              >
                {{ $t('common.cancel') }}
              </ElButton>
            </div>
          </ElForm>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import AppConfig from '@/config'
  import { useUserStore } from '@/store/modules/user'
  import { useI18n } from 'vue-i18n'
  import { HttpError } from '@/utils/http/error'
  import { fetchLogin } from '@/api/auth'
  import { ElNotification, type FormInstance, type FormRules } from 'element-plus'

  defineOptions({ name: 'Login' })

  const { t, locale } = useI18n()
  const formKey = ref(0)

  // 监听语言切换，重置表单
  watch(locale, () => {
    formKey.value++
  })

  const userStore = useUserStore()
  const router = useRouter()
  const route = useRoute()

  const systemName = AppConfig.systemInfo.name
  const formRef = ref<FormInstance>()
  const form2FARef = ref<FormInstance>()

  // 是否处于 2FA 第二步
  const require2FA = ref(false)

  const formData = reactive({
    username: '',
    password: '',
    two_factor_code: ''
  })

  const rules = computed<FormRules>(() => ({
    username: [{ required: true, message: t('login.placeholder.username'), trigger: 'blur' }],
    password: [{ required: true, message: t('login.placeholder.password'), trigger: 'blur' }]
  }))

  const rules2FA = computed<FormRules>(() => ({
    two_factor_code: [
      { required: true, message: t('login.placeholder.twoFactorCode'), trigger: 'blur' },
      { len: 6, message: t('login.placeholder.twoFactorCode'), trigger: 'blur' }
    ]
  }))

  const loading = ref(false)

  // 取消 2FA，回到第一步
  const cancelTwoFactor = () => {
    require2FA.value = false
    formData.two_factor_code = ''
  }

  // 登录
  const handleSubmit = async () => {
    // validate current step's form
    const activeForm = require2FA.value ? form2FARef.value : formRef.value
    if (!activeForm) return

    try {
      const valid = await activeForm.validate()
      if (!valid) return

      loading.value = true

      const params: Api.Auth.LoginParams = {
        username: formData.username,
        password: formData.password
      }
      if (require2FA.value && formData.two_factor_code) {
        params.two_factor_code = formData.two_factor_code
      }

      // fetchLogin returns either LoginResponse or Login2FARequired
      const result = await fetchLogin(params) as any

      // Check for 2FA intermediate step:
      // gocron returns { code:0, message:"2fa_code_required", data:{ require_2fa: true } }
      // The HTTP interceptor lets code=0 through; the 'data' field here is the parsed data.
      if (result && (result as any).require_2fa === true) {
        require2FA.value = true
        loading.value = false
        return
      }

      // Normal login success — result is LoginResponse { token, uid, username, is_admin }
      const loginData = result as Api.Auth.LoginResponse
      if (!loginData.token) {
        throw new Error('Login failed - no token received')
      }

      // Determine roles from is_admin
      const roles = loginData.is_admin === 1 ? ['R_SUPER', 'R_ADMIN'] : ['R_USER']

      // Store token (no refresh token in gocron)
      userStore.setToken(loginData.token)
      userStore.setLoginStatus(true)
      userStore.setUserInfo({
        userId: loginData.uid,
        userName: loginData.username,
        isAdmin: loginData.is_admin,
        roles,
        buttons: [],
        email: ''
      })

      // 登录成功处理
      showLoginSuccessNotice()

      // 获取 redirect 参数，如果存在则跳转到指定页面，否则跳转到首页
      const redirect = route.query.redirect as string
      router.push(redirect || '/')
    } catch (error) {
      if (error instanceof HttpError) {
        // HTTP errors (non-zero code) are already shown by the error interceptor
        // unless showErrorMessage: false — since we set that in fetchLogin, show manually
        ElMessage.error(error.message)
      } else if (error instanceof Error) {
        ElMessage.error(error.message)
        console.error('[Login] Error:', error)
      }
    } finally {
      loading.value = false
    }
  }

  // 登录成功提示
  const showLoginSuccessNotice = () => {
    setTimeout(() => {
      ElNotification({
        title: t('login.success.title'),
        type: 'success',
        duration: 2500,
        zIndex: 10000,
        message: `${t('login.success.message')}, ${systemName}!`
      })
    }, 1000)
  }
</script>

<style scoped>
  @import './style.css';
</style>

<style lang="scss" scoped>
  /* no account-select overrides needed */
</style>
