<!-- Log retention settings page -->
<template>
  <div class="log-retention-page art-full-height">
    <ElCard shadow="never">
      <template #header>
        <span class="text-base font-medium">{{ t('logRetention.title') }}</span>
      </template>

      <p class="description">{{ t('logRetention.description') }}</p>

      <ElForm
        ref="formRef"
        :model="form"
        label-width="180px"
        style="max-width: 560px; margin-top: 16px"
        @submit.prevent
      >
        <!-- Task log retention days -->
        <ElFormItem :label="t('logRetention.taskLogDays')">
          <ElInputNumber
            v-model="form.days"
            :min="0"
            :max="3650"
            controls-position="right"
            class="retention-input"
          />
          <span class="unit-hint">{{ t('logRetention.daysPlaceholder') }}</span>
        </ElFormItem>

        <!-- Cleanup time -->
        <ElFormItem :label="t('logRetention.cleanupTime')">
          <ElTimePicker
            v-model="form.cleanup_time"
            format="HH:mm"
            value-format="HH:mm"
            :placeholder="t('logRetention.selectTime')"
            class="retention-input"
          />
        </ElFormItem>

        <!-- Log file size limit -->
        <ElFormItem :label="t('logRetention.logFileSizeLimit')">
          <ElInputNumber
            v-model="form.file_size_limit"
            :min="0"
            :max="10240"
            controls-position="right"
            class="retention-input"
          />
          <span class="unit-hint">MB</span>
        </ElFormItem>

        <ElFormItem>
          <ElButton type="primary" :loading="saving" @click="handleSave" v-ripple>
            {{ t('logRetention.save') }}
          </ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { fetchLogRetention, updateLogRetention } from '@/api/log-retention'

  defineOptions({ name: 'LogRetention' })

  const { t } = useI18n()

  const saving = ref(false)

  const form = reactive({
    days: 0,
    cleanup_time: '03:00',
    file_size_limit: 0
  })

  async function loadData() {
    try {
      const data = await fetchLogRetention()
      if (data) {
        form.days = data.days ?? 0
        form.cleanup_time = data.cleanup_time || '03:00'
        form.file_size_limit = data.file_size_limit ?? 0
      }
    } catch {
      // error toast handled by http interceptor
    }
  }

  async function handleSave() {
    saving.value = true
    try {
      await updateLogRetention({
        days: form.days,
        cleanup_time: form.cleanup_time,
        file_size_limit: form.file_size_limit
      })
      ElMessage.success(t('logRetention.saveSuccess'))
    } catch {
      // error toast handled by http interceptor
    } finally {
      saving.value = false
    }
  }

  onMounted(loadData)
</script>

<style scoped>
  .log-retention-page {
    display: flex;
    flex-direction: column;
  }

  .description {
    color: var(--el-text-color-secondary);
    font-size: 13px;
    margin: 0;
  }

  .unit-hint {
    margin-left: 8px;
    color: var(--el-text-color-secondary);
    font-size: 13px;
  }

  .retention-input {
    width: 200px;
    max-width: 100%;
  }

  @media (max-width: 768px) {
    .retention-input {
      width: 100%;
    }
    .unit-hint {
      display: block;
      margin-left: 0;
      margin-top: 4px;
    }
  }
</style>
