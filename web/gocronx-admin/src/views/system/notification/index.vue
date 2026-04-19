<!-- Notification configuration page — Email / Slack / Webhook -->
<template>
  <div class="notification-page art-full-height">
    <!-- Template variables info alert -->
    <ElAlert type="info" :closable="false" style="margin-bottom: 16px">
      <template #title>
        <div style="font-weight: 600; margin-bottom: 6px">
          {{ t('notification.templateVariables') }}
        </div>
        <div style="font-size: 13px; line-height: 1.9">
          <div v-for="v in templateVars" :key="v.key">
            <code>{{ v.key }}</code> — {{ v.label }}
          </div>
        </div>
      </template>
    </ElAlert>

    <!-- Tabs -->
    <ElTabs v-model="activeTab" type="border-card" class="notification-tabs">
      <ElTabPane :label="t('notification.tabEmail')" name="email">
        <EmailTab />
      </ElTabPane>
      <ElTabPane label="Slack" name="slack">
        <SlackTab />
      </ElTabPane>
      <ElTabPane label="Webhook" name="webhook">
        <WebhookTab />
      </ElTabPane>
    </ElTabs>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed } from 'vue'
  import { useI18n } from 'vue-i18n'
  import EmailTab from './modules/email-tab.vue'
  import SlackTab from './modules/slack-tab.vue'
  import WebhookTab from './modules/webhook-tab.vue'

  defineOptions({ name: 'Notification' })

  const { t } = useI18n()

  const activeTab = ref('email')

  const templateVars = computed(() => [
    { key: '{{.TaskId}}', label: t('notification.taskIdVar') },
    { key: '{{.TaskName}}', label: t('notification.taskNameVar') },
    { key: '{{.Status}}', label: t('notification.statusVar') },
    { key: '{{.Result}}', label: t('notification.resultVar') },
    { key: '{{.Remark}}', label: t('notification.remarkVar') }
  ])
</script>

<style scoped>
  .notification-page {
    display: flex;
    flex-direction: column;
  }

  .notification-tabs :deep(.el-tabs__content) {
    padding: 16px;
  }
</style>
