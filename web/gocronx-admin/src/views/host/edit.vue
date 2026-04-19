<!-- 主机新建 / 编辑表单 -->
<!-- Routes: /host/create  or  /host/edit/:id (registered by HOST-LIST agent) -->
<template>
  <div class="host-edit-page art-full-height">
    <ElCard class="p-2" shadow="never">
      <template #header>
        <span class="text-base font-medium">
          {{ isEdit ? t('host.editTitle') : t('host.createTitle') }}
        </span>
      </template>

      <ElForm
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
        style="max-width: 560px"
        @submit.prevent
      >
        <!-- hidden id -->
        <input type="hidden" :value="form.id" />

        <!-- alias -->
        <ElFormItem :label="t('host.alias')" prop="alias">
          <ElInput v-model="form.alias" :placeholder="t('host.aliasPlaceholder')" clearable />
        </ElFormItem>

        <!-- name (host / IP) -->
        <ElFormItem :label="t('host.name')" prop="name">
          <ElInput v-model="form.name" :placeholder="t('host.namePlaceholder')" clearable />
        </ElFormItem>

        <!-- port -->
        <ElFormItem :label="t('host.port')" prop="port">
          <ElInputNumber
            v-model="form.port"
            :min="1"
            :max="65535"
            :placeholder="t('host.portPlaceholder')"
            controls-position="right"
            style="width: 100%"
          />
        </ElFormItem>

        <!-- remark -->
        <ElFormItem :label="t('host.remark')">
          <ElInput v-model="form.remark" type="textarea" :rows="4" />
        </ElFormItem>

        <!-- actions -->
        <ElFormItem>
          <ElButton type="primary" :loading="submitting" @click="handleSubmit" v-ripple>
            {{ t('host.save') }}
          </ElButton>
          <ElButton @click="handleCancel">
            {{ t('host.cancel') }}
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
  import { fetchHostDetail, saveHost } from '@/api/host'

  defineOptions({ name: 'HostEdit' })

  const { t } = useI18n()
  const route = useRoute()
  const router = useRouter()

  // ── State ────────────────────────────────────────────────────────────────────

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    id: 0,
    name: '',
    alias: '',
    port: 5921,
    remark: ''
  })

  // ── Computed ─────────────────────────────────────────────────────────────────

  const routeId = computed(() => {
    const id = route.params.id
    if (!id || id === '0') return 0
    return Number(id)
  })

  const isEdit = computed(() => routeId.value > 0)

  // ── Validation rules ─────────────────────────────────────────────────────────

  const rules = computed<FormRules>(() => ({
    alias: [{ required: true, message: t('host.aliasRequired'), trigger: 'blur' }],
    name: [{ required: true, message: t('host.nameRequired'), trigger: 'blur' }],
    port: [
      { required: true, message: t('host.portRequired'), trigger: 'blur' },
      {
        type: 'number',
        min: 1,
        max: 65535,
        message: t('host.portRange'),
        trigger: 'blur'
      }
    ]
  }))

  // ── Data loading ─────────────────────────────────────────────────────────────

  async function loadDetail(id: number) {
    try {
      const data = await fetchHostDetail(id)
      if (!data) {
        ElMessage.error(t('host.notFound'))
        router.push('/host')
        return
      }
      form.id = data.id
      form.name = data.name
      form.alias = data.alias
      form.port = data.port
      form.remark = data.remark ?? ''
    } catch {
      // error toast handled by http interceptor
      router.push('/host')
    }
  }

  // ── Submit ───────────────────────────────────────────────────────────────────

  async function handleSubmit() {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    submitting.value = true
    try {
      await saveHost({
        ...(isEdit.value ? { id: form.id } : {}),
        name: form.name,
        alias: form.alias,
        port: form.port,
        remark: form.remark
      })
      ElMessage.success(isEdit.value ? t('host.updateSuccess') : t('host.createSuccess'))
      router.push('/host')
    } catch {
      // error toast handled by http interceptor
    } finally {
      submitting.value = false
    }
  }

  function handleCancel() {
    router.push('/host')
  }

  // ── Lifecycle ─────────────────────────────────────────────────────────────────

  onMounted(() => {
    if (isEdit.value) {
      loadDetail(routeId.value)
    }
  })

  // Re-load when route param changes (e.g. navigating between create and edit)
  watch(routeId, (newId) => {
    if (newId > 0) {
      loadDetail(newId)
    } else {
      // Reset to blank form for create mode
      Object.assign(form, { id: 0, name: '', alias: '', port: 5921, remark: '' })
      formRef.value?.clearValidate()
    }
  })
</script>

<style scoped>
  .host-edit-page {
    display: flex;
    flex-direction: column;
  }
</style>
