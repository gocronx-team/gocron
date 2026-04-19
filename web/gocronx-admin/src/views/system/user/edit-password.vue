<!-- Admin resets another user's password -->
<!-- Route: /system/user/edit-password/:id -->
<template>
  <div class="user-edit-password-page art-full-height">
    <ElCard class="p-2" shadow="never">
      <template #header>
        <span class="text-base font-medium">{{ t('user.editPasswordTitle') }}</span>
      </template>

      <ElForm
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="130px"
        style="max-width: 560px"
        @submit.prevent
      >
        <!-- new password -->
        <ElFormItem :label="t('user.newPasswordLabel')" prop="new_password">
          <ElInput
            v-model="form.new_password"
            type="password"
            :placeholder="t('user.passwordPlaceholder')"
            show-password
          />
        </ElFormItem>

        <!-- confirm new password -->
        <ElFormItem :label="t('user.confirmPassword')" prop="confirm_new_password">
          <ElInput
            v-model="form.confirm_new_password"
            type="password"
            :placeholder="t('user.passwordPlaceholder')"
            show-password
          />
        </ElFormItem>

        <!-- actions -->
        <ElFormItem>
          <ElButton type="primary" :loading="submitting" @click="handleSubmit" v-ripple>
            {{ t('user.save') }}
          </ElButton>
          <ElButton @click="handleCancel">
            {{ t('user.cancel') }}
          </ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { useRoute, useRouter } from 'vue-router'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchUserEditPassword } from '@/api/user'

  defineOptions({ name: 'UserEditPassword' })

  const { t } = useI18n()
  const route = useRoute()
  const router = useRouter()

  // ── State ─────────────────────────────────────────────────────────────────────

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    new_password: '',
    confirm_new_password: ''
  })

  // ── Computed ──────────────────────────────────────────────────────────────────

  const userId = computed(() => Number(route.params.id))

  // ── Validation rules ──────────────────────────────────────────────────────────

  const rules = computed<FormRules>(() => ({
    new_password: [
      { required: true, message: t('user.passwordRequired'), trigger: 'blur' },
      { min: 8, message: t('user.passwordTooShort'), trigger: 'blur' }
    ],
    confirm_new_password: [
      { required: true, message: t('user.confirmPasswordRequired'), trigger: 'blur' },
      {
        validator: (_rule: unknown, value: string, callback: (e?: Error) => void) => {
          if (value !== form.new_password) {
            callback(new Error(t('user.passwordMismatch')))
          } else {
            callback()
          }
        },
        trigger: 'blur'
      }
    ]
  }))

  // ── Submit ────────────────────────────────────────────────────────────────────

  async function handleSubmit() {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    submitting.value = true
    try {
      await fetchUserEditPassword({
        id: userId.value,
        new_password: form.new_password,
        confirm_new_password: form.confirm_new_password
      })
      ElMessage.success(t('user.resetPasswordSuccess'))
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
</script>

<style scoped>
  .user-edit-password-page {
    display: flex;
    flex-direction: column;
  }
</style>
