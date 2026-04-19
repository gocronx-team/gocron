<!-- Webhook notification configuration tab -->
<template>
  <ElCard shadow="never">
    <template #header>
      <span class="text-base font-medium">Webhook</span>
    </template>

    <ElAlert
      :title="t('notification.webhookDescription')"
      type="info"
      :closable="false"
      style="margin-bottom: 16px; max-width: 640px"
    />

    <ElForm
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="110px"
      style="max-width: 640px"
      @submit.prevent
    >
      <ElFormItem :label="t('notification.template')" prop="template">
        <ElInput
          v-model.trim="form.template"
          type="textarea"
          :rows="8"
          :placeholder="webhookPlaceholder"
        />
      </ElFormItem>

      <ElFormItem>
        <ElButton type="primary" :loading="saving" @click="handleSave" v-ripple>
          {{ t('notification.save') }}
        </ElButton>
      </ElFormItem>
    </ElForm>

    <!-- Webhook URLs -->
    <div class="urls-section">
      <div class="section-header">
        <span class="text-sm font-medium">{{ t('notification.webhookUrls') }}</span>
        <ElButton type="primary" size="small" @click="openDialog" v-ripple>
          {{ t('notification.addWebhookUrl') }}
        </ElButton>
      </div>
      <div class="tag-list">
        <ElTag
          v-for="item in webhookUrls"
          :key="item.id"
          closable
          @close="handleRemoveUrl(item.id)"
        >
          {{ item.name }} - {{ item.url }}
        </ElTag>
        <span v-if="webhookUrls.length === 0" class="empty-hint">—</span>
      </div>
    </div>
  </ElCard>

  <!-- Add webhook URL dialog -->
  <ElDialog
    v-model="dialogVisible"
    :title="t('notification.addWebhookUrl')"
    width="480px"
    @closed="resetDialog"
  >
    <ElForm :model="dialogForm" label-width="90px">
      <ElFormItem :label="t('notification.webhookName')">
        <ElInput v-model.trim="dialogForm.name" clearable />
      </ElFormItem>
      <ElFormItem label="URL">
        <ElInput v-model.trim="dialogForm.url" clearable />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="dialogVisible = false">{{ t('notification.cancel') }}</ElButton>
      <ElButton type="primary" :loading="dialogSaving" @click="handleSaveUrl" v-ripple>
        {{ t('notification.confirm') }}
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchWebhook, updateWebhook, createWebhookUrl, removeWebhookUrl } from '@/api/notification'
  import type { WebhookUrl } from '@/api/notification'

  defineOptions({ name: 'WebhookTab' })

  const { t } = useI18n()

  // ── State ─────────────────────────────────────────────────────────────────────

  const formRef = ref<FormInstance>()
  const saving = ref(false)
  const webhookUrls = ref<WebhookUrl[]>([])

  const form = reactive({
    template: ''
  })

  const dialogVisible = ref(false)
  const dialogSaving = ref(false)
  const dialogForm = reactive({ name: '', url: '' })

  // ── Computed ──────────────────────────────────────────────────────────────────

  const webhookPlaceholder = computed(
    () =>
      `{"task_id": "{{.TaskId}}", "task_name": "{{.TaskName}}", "status": "{{.Status}}", "result": "{{.Result}}", "remark": "{{.Remark}}"}`
  )

  const rules = computed<FormRules>(() => ({
    template: [{ required: true, message: t('notification.pleaseEnterTemplate'), trigger: 'blur' }]
  }))

  // ── Methods ───────────────────────────────────────────────────────────────────

  async function loadData() {
    try {
      const data = await fetchWebhook()
      if (data) {
        form.template = data.template || ''
        webhookUrls.value = data.webhook_urls || []
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
      await updateWebhook({ template: form.template })
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
    dialogForm.name = ''
    dialogForm.url = ''
  }

  async function handleSaveUrl() {
    if (!dialogForm.name || !dialogForm.url) {
      ElMessage.error(t('notification.incompleteParameters'))
      return
    }
    dialogSaving.value = true
    try {
      await createWebhookUrl({ name: dialogForm.name, url: dialogForm.url })
      dialogVisible.value = false
      await loadData()
    } catch {
      // error toast handled by http interceptor
    } finally {
      dialogSaving.value = false
    }
  }

  async function handleRemoveUrl(id: number) {
    try {
      await removeWebhookUrl(id)
      await loadData()
    } catch {
      // error toast handled by http interceptor
    }
  }

  onMounted(loadData)
</script>

<style scoped>
  .urls-section {
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
