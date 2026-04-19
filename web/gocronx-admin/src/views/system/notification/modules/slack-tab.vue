<!-- Slack notification configuration tab -->
<template>
  <ElCard shadow="never">
    <template #header>
      <span class="text-base font-medium">Slack</span>
    </template>

    <ElForm
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="110px"
      style="max-width: 640px"
      @submit.prevent
    >
      <ElFormItem :label="t('notification.slackUrl')" prop="url">
        <ElInput v-model.trim="form.url" clearable />
      </ElFormItem>

      <ElFormItem :label="t('notification.template')" prop="template">
        <ElInput
          v-model="form.template"
          type="textarea"
          :rows="8"
          :placeholder="templatePlaceholder"
        />
      </ElFormItem>

      <ElFormItem>
        <ElButton type="primary" :loading="saving" @click="handleSave" v-ripple>
          {{ t('notification.save') }}
        </ElButton>
      </ElFormItem>
    </ElForm>

    <!-- Channels -->
    <div class="channels-section">
      <div class="section-header">
        <span class="text-sm font-medium">{{ t('notification.channels') }}</span>
        <ElButton type="primary" size="small" @click="openDialog" v-ripple>
          {{ t('notification.addChannel') }}
        </ElButton>
      </div>
      <div class="tag-list">
        <ElTag
          v-for="item in channels"
          :key="item.id"
          closable
          @close="handleRemoveChannel(item.id)"
        >
          {{ item.name }}
        </ElTag>
        <span v-if="channels.length === 0" class="empty-hint">—</span>
      </div>
    </div>
  </ElCard>

  <!-- Add channel dialog -->
  <ElDialog
    v-model="dialogVisible"
    :title="t('notification.addChannel')"
    width="400px"
    @closed="resetDialog"
  >
    <ElForm :model="dialogForm" label-width="90px">
      <ElFormItem :label="t('notification.channelName')">
        <ElInput v-model.trim="dialogForm.channel" clearable />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="dialogVisible = false">{{ t('notification.cancel') }}</ElButton>
      <ElButton type="primary" :loading="dialogSaving" @click="handleSaveChannel" v-ripple>
        {{ t('notification.confirm') }}
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchSlack, updateSlack, createSlackChannel, removeSlackChannel } from '@/api/notification'
  import type { SlackChannel } from '@/api/notification'

  defineOptions({ name: 'SlackTab' })

  const { t } = useI18n()

  // ── State ─────────────────────────────────────────────────────────────────────

  const formRef = ref<FormInstance>()
  const saving = ref(false)
  const channels = ref<SlackChannel[]>([])

  const form = reactive({
    url: '',
    template: ''
  })

  const dialogVisible = ref(false)
  const dialogSaving = ref(false)
  const dialogForm = reactive({ channel: '' })

  // ── Computed ──────────────────────────────────────────────────────────────────

  const templatePlaceholder = computed(
    () =>
      `${t('notification.taskIdVar')}: {{.TaskId}}\n${t('notification.taskNameVar')}: {{.TaskName}}\n${t('notification.statusVar')}: {{.Status}}\n${t('notification.resultVar')}: {{.Result}}\n${t('notification.remarkVar')}: {{.Remark}}`
  )

  const rules = computed<FormRules>(() => ({
    url: [
      { required: true, type: 'url', message: t('notification.pleaseEnterValidUrl'), trigger: 'blur' }
    ],
    template: [{ required: true, message: t('notification.pleaseEnterTemplate'), trigger: 'blur' }]
  }))

  // ── Methods ───────────────────────────────────────────────────────────────────

  async function loadData() {
    try {
      const data = await fetchSlack()
      if (data) {
        form.url = data.url || ''
        form.template = data.template || ''
        channels.value = data.channels || []
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
      await updateSlack({ ...form })
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
    dialogForm.channel = ''
  }

  async function handleSaveChannel() {
    if (!dialogForm.channel) {
      ElMessage.error(t('notification.pleaseEnterChannelName'))
      return
    }
    dialogSaving.value = true
    try {
      await createSlackChannel(dialogForm.channel)
      dialogVisible.value = false
      await loadData()
    } catch {
      // error toast handled by http interceptor
    } finally {
      dialogSaving.value = false
    }
  }

  async function handleRemoveChannel(id: number) {
    try {
      await removeSlackChannel(id)
      await loadData()
    } catch {
      // error toast handled by http interceptor
    }
  }

  onMounted(loadData)
</script>

<style scoped>
  .channels-section {
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
