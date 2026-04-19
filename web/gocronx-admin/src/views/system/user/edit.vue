<!-- User create / edit form -->
<!-- Routes: /system/user/edit/0  (create)  or  /system/user/edit/:id  (edit) -->
<template>
  <div class="user-edit-page art-full-height">
    <ElCard class="p-2" shadow="never">
      <template #header>
        <span class="text-base font-medium">
          {{ isEdit ? t('user.editTitle') : t('user.createTitle') }}
        </span>
      </template>

      <ElForm
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="130px"
        style="max-width: 560px"
        @submit.prevent
      >
        <!-- hidden id -->
        <input type="hidden" :value="form.id" />

        <!-- username -->
        <ElFormItem :label="t('user.name')" prop="name">
          <ElInput v-model="form.name" :placeholder="t('user.namePlaceholder')" clearable />
        </ElFormItem>

        <!-- email -->
        <ElFormItem :label="t('user.email')" prop="email">
          <ElInput v-model="form.email" :placeholder="t('user.emailPlaceholder')" clearable />
        </ElFormItem>

        <!-- password — required on create, hidden on edit -->
        <template v-if="!isEdit">
          <ElFormItem :label="t('user.password')" prop="password">
            <ElInput
              v-model="form.password"
              type="password"
              :placeholder="t('user.passwordPlaceholder')"
              show-password
            />
          </ElFormItem>

          <ElFormItem :label="t('user.confirmPassword')" prop="confirm_password">
            <ElInput
              v-model="form.confirm_password"
              type="password"
              :placeholder="t('user.passwordPlaceholder')"
              show-password
            />
          </ElFormItem>
        </template>

        <!-- role -->
        <ElFormItem :label="t('user.role')" prop="is_admin">
          <ElRadioGroup v-model="form.is_admin">
            <ElRadio :value="0">{{ t('user.normalUserLabel') }}</ElRadio>
            <ElRadio :value="1">{{ t('user.adminLabel') }}</ElRadio>
          </ElRadioGroup>
        </ElFormItem>

        <!-- status -->
        <ElFormItem :label="t('user.status')" prop="status">
          <ElRadioGroup v-model="form.status">
            <ElRadio :value="1">{{ t('user.enabled') }}</ElRadio>
            <ElRadio :value="0">{{ t('user.disabled') }}</ElRadio>
          </ElRadioGroup>
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
</style>
