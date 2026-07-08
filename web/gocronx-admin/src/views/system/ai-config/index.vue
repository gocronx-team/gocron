<!-- AI (LLM) configuration page -->
<template>
  <div class="ai-config-page art-full-height">
    <ElCard shadow="never">
      <template #header>
        <span class="text-base font-medium">{{ t('aiConfig.title') }}</span>
      </template>

      <ElAlert :closable="false" type="info" show-icon style="margin-bottom: 16px">
        <template #title>{{ t('aiConfig.intro') }}</template>
      </ElAlert>

      <ElForm
        :model="form"
        label-width="160px"
        style="max-width: 620px; margin-top: 8px"
        @submit.prevent
      >
        <ElFormItem :label="t('aiConfig.enable')">
          <ElSwitch v-model="form.enable" />
        </ElFormItem>

        <ElFormItem :label="t('aiConfig.baseUrl')">
          <ElInput v-model="form.base_url" placeholder="https://api.openai.com/v1" clearable />
          <div class="hint">{{ t('aiConfig.baseUrlHint') }}</div>
        </ElFormItem>

        <ElFormItem :label="t('aiConfig.apiKey')">
          <ElInput
            v-model="form.api_key"
            type="password"
            show-password
            :placeholder="apiKeyPlaceholder"
            clearable
          />
        </ElFormItem>

        <ElFormItem :label="t('aiConfig.model')">
          <ElInput v-model="form.model" placeholder="gpt-4o-mini" clearable />
        </ElFormItem>

        <ElFormItem>
          <ElButton type="primary" :loading="saving" @click="handleSave" v-ripple>
            {{ t('aiConfig.save') }}
          </ElButton>
          <ElButton :loading="testing" @click="handleTest">
            {{ t('aiConfig.test') }}
          </ElButton>
          <div class="hint">{{ t('aiConfig.testHint') }}</div>
        </ElFormItem>
      </ElForm>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage } from 'element-plus'
  import { fetchLLMConfig, updateLLMConfig, testLLMConfig } from '@/api/ai'

  defineOptions({ name: 'AiConfig' })

  const { t } = useI18n()

  const saving = ref(false)
  const testing = ref(false)
  const apiKeySet = ref(false)

  const form = reactive({
    enable: false,
    base_url: '',
    api_key: '',
    model: ''
  })

  const apiKeyPlaceholder = computed(() =>
    apiKeySet.value ? t('aiConfig.apiKeyConfigured') : t('aiConfig.apiKeyPlaceholder')
  )

  async function loadData() {
    try {
      const data = await fetchLLMConfig()
      if (data) {
        form.enable = !!data.enable
        form.base_url = data.base_url || ''
        form.model = data.model || ''
        apiKeySet.value = !!data.api_key_set
      }
    } catch {
      // error toast handled by http interceptor
    }
  }

  async function handleSave() {
    saving.value = true
    try {
      await updateLLMConfig({
        enable: form.enable,
        base_url: form.base_url.trim(),
        api_key: form.api_key, // 留空 = 不修改
        model: form.model.trim()
      })
      ElMessage.success(t('aiConfig.saveSuccess'))
      form.api_key = ''
      loadData()
    } catch {
      // error toast handled by http interceptor
    } finally {
      saving.value = false
    }
  }

  // 测试连接:用当前表单值发一句极短对话。api_key 留空则后端用已保存的值。
  // 失败时 http 拦截器会弹出后端带回的真实原因(如 invalid api key / connection refused)。
  async function handleTest() {
    testing.value = true
    try {
      await testLLMConfig({
        enable: form.enable,
        base_url: form.base_url.trim(),
        api_key: form.api_key,
        model: form.model.trim()
      })
      ElMessage.success(t('aiConfig.testSuccess'))
    } catch {
      // 失败详情由 http 拦截器按后端 message 弹出
    } finally {
      testing.value = false
    }
  }

  onMounted(loadData)
</script>

<style scoped>
  .hint {
    margin-top: 4px;
    font-size: 12px;
    line-height: 1.5;
    color: var(--el-text-color-secondary);
  }
</style>
