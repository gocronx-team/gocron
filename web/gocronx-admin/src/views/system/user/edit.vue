<!-- User create / edit form -->
<!-- Routes: /system/user/edit/0  (create)  or  /system/user/edit/:id  (edit) -->
<template>
  <div class="user-edit-page art-full-height">
    <ElCard shadow="never">
      <template #header>
        <span class="text-base font-medium">
          {{ isEdit ? t('user.editTitle') : t('user.createTitle') }}
        </span>
      </template>

      <ElForm
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        style="max-width: 720px"
        @submit.prevent
      >
        <!-- ── Basic Info ─────────────────────────────────────────────── -->
        <ElDivider content-position="left">{{ t('user.sectionBasic') }}</ElDivider>

        <ElRow :gutter="24">
          <ElCol :span="12">
            <ElFormItem :label="t('user.name')" prop="name">
              <ElInput
                v-model.trim="form.name"
                :placeholder="t('user.namePlaceholder')"
                autocomplete="username"
                clearable
              />
            </ElFormItem>
          </ElCol>
          <ElCol :span="12">
            <ElFormItem :label="t('user.email')" prop="email">
              <ElInput
                v-model.trim="form.email"
                type="email"
                :placeholder="t('user.emailPlaceholder')"
                autocomplete="email"
                clearable
              />
            </ElFormItem>
          </ElCol>
        </ElRow>

        <!-- ── Credentials (create only) ──────────────────────────────── -->
        <template v-if="!isEdit">
          <ElDivider content-position="left">{{ t('user.sectionCredentials') }}</ElDivider>

          <ElRow :gutter="24">
            <ElCol :span="12">
              <ElFormItem :label="t('user.password')" prop="password">
                <ElInput
                  v-model="form.password"
                  type="password"
                  :placeholder="t('user.passwordPlaceholder')"
                  autocomplete="new-password"
                  show-password
                >
                  <template #append>
                    <ElButton :icon="Refresh" @click="generateAndCopyPassword">
                      {{ t('user.generatePassword') }}
                    </ElButton>
                  </template>
                </ElInput>

                <!-- Strength meter -->
                <div v-if="form.password" class="pw-strength">
                  <div class="pw-strength-bar">
                    <div
                      class="pw-strength-fill"
                      :style="{ width: passwordStrength.pct + '%', background: passwordStrength.color }"
                    />
                  </div>
                  <span class="pw-strength-label" :style="{ color: passwordStrength.color }">
                    {{ t('user.passwordStrength') }}: {{ passwordStrength.label }}
                  </span>
                </div>
              </ElFormItem>
            </ElCol>

            <ElCol :span="12">
              <ElFormItem :label="t('user.confirmPassword')" prop="confirm_password">
                <ElInput
                  v-model="form.confirm_password"
                  type="password"
                  :placeholder="t('user.passwordPlaceholder')"
                  autocomplete="new-password"
                  show-password
                />
              </ElFormItem>
            </ElCol>
          </ElRow>
        </template>

        <!-- ── Access ─────────────────────────────────────────────────── -->
        <ElDivider content-position="left">{{ t('user.sectionAccess') }}</ElDivider>

        <ElRow :gutter="24">
          <ElCol :span="12">
            <ElFormItem :label="t('user.role')" prop="is_admin">
              <ElRadioGroup v-model="form.is_admin">
                <ElRadio :value="0">{{ t('user.normalUserLabel') }}</ElRadio>
                <ElRadio :value="1">{{ t('user.adminLabel') }}</ElRadio>
              </ElRadioGroup>
              <div class="role-hint">
                {{ form.is_admin === 1 ? t('user.roleAdminHint') : t('user.roleUserHint') }}
              </div>
            </ElFormItem>
          </ElCol>

          <ElCol :span="12">
            <ElFormItem :label="t('user.status')" prop="status">
              <ElRadioGroup v-model="form.status">
                <ElRadio :value="1">{{ t('user.enabled') }}</ElRadio>
                <ElRadio :value="0">{{ t('user.disabled') }}</ElRadio>
              </ElRadioGroup>
            </ElFormItem>
          </ElCol>
        </ElRow>

        <!-- ── Actions ────────────────────────────────────────────────── -->
        <ElFormItem>
          <ElButton type="primary" :loading="submitting" @click="handleSubmit" v-ripple>
            {{ t('user.save') }}
          </ElButton>
          <ElButton @click="handleCancel">{{ t('user.cancel') }}</ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { useRoute, useRouter } from 'vue-router'
  import { Refresh } from '@element-plus/icons-vue'
  import type { FormInstance, FormRules } from 'element-plus'
  import request from '@/utils/http'
  import { fetchUserDetail } from '@/api/user'

  defineOptions({ name: 'UserEdit' })

  const { t } = useI18n()
  const route = useRoute()
  const router = useRouter()

  // ── State ─────────────────────────────────────────────────────────────────────

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    id: 0,
    name: '',
    email: '',
    password: '',
    confirm_password: '',
    is_admin: 0,
    status: 1
  })

  // ── Computed ──────────────────────────────────────────────────────────────────

  const routeId = computed(() => {
    const id = route.params.id
    if (!id || id === '0') return 0
    return Number(id)
  })

  const isEdit = computed(() => routeId.value > 0)

  /**
   * Password strength scoring.
   * 0 points: too short
   * 1-2 points: weak (length only or single class)
   * 3 points: medium (2-3 classes)
   * 4+ points: strong (length + 3+ classes)
   */
  const passwordStrength = computed(() => {
    const pwd = form.password
    let score = 0
    if (pwd.length >= 8) score++
    if (pwd.length >= 12) score++
    if (/[a-z]/.test(pwd) && /[A-Z]/.test(pwd)) score++
    if (/\d/.test(pwd)) score++
    if (/[^a-zA-Z0-9]/.test(pwd)) score++

    if (score <= 2) return { label: t('user.strengthWeak'), pct: 33, color: '#F56C6C' }
    if (score <= 3) return { label: t('user.strengthMedium'), pct: 66, color: '#E6A23C' }
    return { label: t('user.strengthStrong'), pct: 100, color: '#67C23A' }
  })

  // ── Validation rules ──────────────────────────────────────────────────────────

  const rules = computed<FormRules>(() => {
    const base: FormRules = {
      name: [{ required: true, message: t('user.nameRequired'), trigger: 'blur' }],
      email: [
        { required: true, message: t('user.emailRequired'), trigger: 'blur' },
        { type: 'email', message: t('user.emailRequired'), trigger: 'blur' }
      ],
      is_admin: [{ required: true, trigger: 'change' }]
    }

    if (!isEdit.value) {
      base.password = [
        { required: true, message: t('user.passwordRequired'), trigger: 'blur' },
        { min: 8, message: t('user.passwordTooShort'), trigger: 'blur' }
      ]
      base.confirm_password = [
        { required: true, message: t('user.confirmPasswordRequired'), trigger: 'blur' },
        {
          validator: (_rule: unknown, value: string, callback: (e?: Error) => void) => {
            if (value !== form.password) {
              callback(new Error(t('user.passwordMismatch')))
            } else {
              callback()
            }
          },
          trigger: 'blur'
        }
      ]
    }

    return base
  })

  // Re-validate confirm when password itself changes; avoids the confusing state
  // where user changes password but confirm still shows "matched" from before.
  watch(
    () => form.password,
    () => {
      if (!isEdit.value && form.confirm_password) {
        formRef.value?.validateField('confirm_password').catch(() => {})
      }
    }
  )

  // ── Password generator ───────────────────────────────────────────────────────

  function generateAndCopyPassword() {
    const chars = {
      lower: 'abcdefghijklmnopqrstuvwxyz',
      upper: 'ABCDEFGHIJKLMNOPQRSTUVWXYZ',
      digit: '0123456789',
      symbol: '!@#$%^&*-_=+'
    }
    const pick = (s: string) => s[Math.floor(Math.random() * s.length)]
    // Guarantee at least one from each class, then fill to 14 chars total
    const out = [pick(chars.lower), pick(chars.upper), pick(chars.digit), pick(chars.symbol)]
    const pool = chars.lower + chars.upper + chars.digit + chars.symbol
    while (out.length < 14) out.push(pick(pool))
    // Shuffle so the forced chars aren't always at the start
    const pwd = out.sort(() => Math.random() - 0.5).join('')

    form.password = pwd
    form.confirm_password = pwd
    navigator.clipboard?.writeText(pwd).catch(() => {})
    ElMessage.success(t('user.generatedHint'))
  }

  // ── Data loading ──────────────────────────────────────────────────────────────

  async function loadDetail(id: number) {
    try {
      const data = await fetchUserDetail(id)
      if (!data) {
        ElMessage.error(t('user.notFound'))
        router.push('/system/user')
        return
      }
      form.id = data.id
      form.name = data.name
      form.email = data.email
      form.is_admin = data.is_admin
      form.status = data.status
    } catch {
      // error toast handled by http interceptor
      router.push('/system/user')
    }
  }

  // ── Submit ────────────────────────────────────────────────────────────────────

  async function handleSubmit() {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    submitting.value = true
    try {
      // Build form-urlencoded payload including status, which is not in
      // the typed UserStoreParams wrapper so we call the endpoint directly.
      const body = new URLSearchParams()
      if (isEdit.value) body.append('id', String(form.id))
      body.append('name', form.name)
      body.append('email', form.email)
      if (form.password) body.append('password', form.password)
      body.append('is_admin', String(form.is_admin))
      body.append('status', String(form.status))

      await request.post<null>({
        url: '/api/user/store',
        data: body,
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
      })
      ElMessage.success(isEdit.value ? t('user.updateSuccess') : t('user.createSuccess'))
      router.push('/system/user')
    } catch {
      // error toast handled by http interceptor
    } finally {
      submitting.value = false
    }
  }

  function handleCancel() {
    router.push('/system/user')
  }

  // ── Lifecycle ─────────────────────────────────────────────────────────────────

  onMounted(() => {
    if (isEdit.value) {
      loadDetail(routeId.value)
    }
  })

  watch(routeId, (newId) => {
    if (newId > 0) {
      loadDetail(newId)
    } else {
      Object.assign(form, {
        id: 0,
        name: '',
        email: '',
        password: '',
        confirm_password: '',
        is_admin: 0,
        status: 1
      })
      formRef.value?.clearValidate()
    }
  })
</script>

<style scoped>
  .user-edit-page {
    display: flex;
    flex-direction: column;
  }

  .role-hint {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    line-height: 1.5;
    margin-top: 4px;
  }

  .pw-strength {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 6px;
  }

  .pw-strength-bar {
    flex: 1;
    max-width: 200px;
    height: 6px;
    background: var(--el-fill-color);
    border-radius: 3px;
    overflow: hidden;
  }

  .pw-strength-fill {
    height: 100%;
    border-radius: 3px;
    transition: width 0.2s, background 0.2s;
  }

  .pw-strength-label {
    font-size: 12px;
    font-variant-numeric: tabular-nums;
  }
</style>
