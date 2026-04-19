<!-- Email notification configuration tab -->
<template>
  <ElCard shadow="never">
    <template #header>
      <span class="text-base font-medium">{{ t('notification.emailServerConfig') }}</span>
    </template>

    <ElForm
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="110px"
      style="max-width: 640px"
      @submit.prevent
    >
      <ElRow :gutter="16">
        <ElCol :span="14">
          <ElFormItem :label="t('notification.smtpHost')" prop="host">
            <ElInput v-model.trim="form.host" clearable />
          </ElFormItem>
        </ElCol>
        <ElCol :span="10">
          <ElFormItem :label="t('notification.smtpPort')" prop="port">
            <ElInputNumber
              v-model="form.port"
              :min="1"
              :max="65535"
              controls-position="right"
              style="width: 100%"
            />
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElRow :gutter="16">
        <ElCol :span="12">
          <ElFormItem :label="t('notification.smtpUser')" prop="user">
            <ElInput v-model.trim="form.user" clearable />
          </ElFormItem>
        </ElCol>
        <ElCol :span="12">
          <ElFormItem :label="t('notification.smtpPassword')" prop="password">
            <ElInput v-model="form.password" type="password" show-password />
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElFormItem :label="t('notification.template')" prop="template">
        <ElInput
          v-model="form.template"
          type="textarea"
          :rows="6"
          :placeholder="templatePlaceholder"
        />
      </ElFormItem>

      <ElFormItem>
        <ElButton type="primary" :loading="saving" @click="handleSave" v-ripple>
          {{ t('notification.save') }}
        </ElButton>
      </ElFormItem>
    </ElForm>

    <!-- Mail recipients -->
    <div class="recipients-section">
      <div class="section-header">
        <span class="text-sm font-medium">{{ t('notification.mailRecipients') }}</span>
        <ElButton type="primary" size="small" @click="openDialog" v-ripple>
          {{ t('notification.addRecipient') }}
        </ElButton>
      </div>
      <div class="tag-list">
        <ElTag
          v-for="item in mailUsers"
          :key="item.id"
          closable
          @close="handleRemoveUser(item.id)"
        >
          {{ item.username }} - {{ item.email }}
        </ElTag>
        <span v-if="mailUsers.length === 0" class="empty-hint">—</span>
      </div>
    </div>
  </ElCard>

  <!-- Add recipient dialog -->
  <ElDialog
    v-model="dialogVisible"
    :title="t('notification.addRecipient')"
    width="400px"
    @closed="resetDialog"
  >
    <ElForm :model="dialogForm" label-width="90px">
      <ElFormItem :label="t('notification.recipientUsername')">
        <ElInput v-model.trim="dialogForm.username" clearable />
      </ElFormItem>
      <ElFormItem :label="t('notification.recipientEmail')">
        <ElInput v-model.trim="dialogForm.email" clearable />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="dialogVisible = false">{{ t('notification.cancel') }}</ElButton>
      <ElButton type="primary" :loading="dialogSaving" @click="handleSaveUser" v-ripple>
        {{ t('notification.confirm') }}
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchMail, updateMail, createMailUser, removeMailUser } from '@/api/notification'
  import type { MailUser } from '@/api/notification'

  defineOptions({ name: 'EmailTab' })

  const { t } = useI18n()

  // ── State ─────────────────────────────────────────────────────────────────────

  const formRef = ref<FormInstance>()
  const saving = ref(false)
  const mailUsers = ref<MailUser[]>([])

  const form = reactive({
    host: '',
    port: 465,
    user: '',
    password: '',
    template: ''
  })

  const dialogVisible = ref(false)
  const dialogSaving = ref(false)
  const dialogForm = reactive({ username: '', email: '' })

  // ── Computed ──────────────────────────────────────────────────────────────────

  const templatePlaceholder = computed(
    () =>
      `${t('notification.taskIdVar')}: {{.TaskId}}\n${t('notification.taskNameVar')}: {{.TaskName}}\n${t('notification.statusVar')}: {{.Status}}\n${t('notification.resultVar')}: {{.Result}}\n${t('notification.remarkVar')}: {{.Remark}}`
  )

  const rules = computed<FormRules>(() => ({
    host: [{ required: true, message: t('notification.pleaseEnterEmailServer'), trigger: 'blur' }],
    port: [
      { required: true, type: 'number', message: t('notification.pleaseEnterValidPort'), trigger: 'blur' }
    ],
    user: [{ required: true, message: t('notification.pleaseEnterUserEmail'), trigger: 'blur' }],
    password: [{ required: true, message: t('notification.pleaseEnterPassword'), trigger: 'blur' }],
    template: [{ required: true, message: t('notification.pleaseEnterTemplate'), trigger: 'blur' }]
  }))

  // ── Methods ───────────────────────────────────────────────────────────────────

  async function loadData() {
    try {
      const data = await fetchMail()
      if (data) {
        form.host = data.host || ''
        form.port = data.port || 465
        form.user = data.user || ''
        form.password = data.password || ''
        form.template = data.template || ''
        mailUsers.value = data.mail_users || []
      }
    } catch {
      // error toast handled by http interceptor
    }
  }

  async function handleSave() {
    if (!formRef.value) return
    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    saving.value = true
    try {
      await updateMail({ ...form })
      ElMessage.success(t('notification.saveSuccess'))
      await loadData()
    } catch {
      // error toast handled by http interceptor
    } finally {
      saving.value = false
    }
  }

  function openDialog() {
    dialogVisible.value = true
  }

  function resetDialog() {
    dialogForm.username = ''
    dialogForm.email = ''
  }

  async function handleSaveUser() {
    if (!dialogForm.username || !dialogForm.email) {
      ElMessage.error(t('notification.incompleteParameters'))
      return
    }
    dialogSaving.value = true
    try {
      await createMailUser({ username: dialogForm.username, email: dialogForm.email })
      dialogVisible.value = false
      await loadData()
    } catch {
      // error toast handled by http interceptor
    } finally {
      dialogSaving.value = false
    }
  }

  async function handleRemoveUser(id: number) {
    try {
      await removeMailUser(id)
      await loadData()
    } catch {
      // error toast handled by http interceptor
    }
  }

  onMounted(loadData)
</script>

<style scoped>
  .recipients-section {
    margin-top: 8px;
  }

  .section-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;
  }

  .tag-list {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .empty-hint {
    color: var(--el-text-color-placeholder);
    font-size: 13px;
  }
</style>
