<template>
  <div>
    <el-tabs v-model="activeName">
      <el-tab-pane :label="t('system.email')" name="email"></el-tab-pane>
      <el-tab-pane label="Slack" name="slack"></el-tab-pane>
      <el-tab-pane label="Webhook" name="webhook"></el-tab-pane>
    </el-tabs>
    <el-alert type="info" :closable="false" style="margin-bottom: 15px;">
      <template #title>
        <div style="font-weight: bold; margin-bottom: 8px;">{{ t('system.templateVariables') }}</div>
        <div style="font-size: 13px; line-height: 1.8;">
          <div><code>{{'{{'}}TaskId{{}}}}</code> - {{ t('system.taskIdVar') }}</div>
          <div><code>{{'{{'}}TaskName{{}}}}</code> - {{ t('system.taskNameVar') }}</div>
          <div><code>{{'{{'}}Status{{}}}}</code> - {{ t('system.statusVar') }}</div>
          <div><code>{{'{{'}}Result{{}}}}</code> - {{ t('system.resultVar') }}</div>
          <div><code>{{'{{'}}Remark{{}}}}</code> - {{ t('task.remark') }}</div>
        </div>
      </template>
    </el-alert>
  </div>
</template>

<script>
import { useI18n } from 'vue-i18n'
export default {
  name: 'notification-tab',
  setup() {
    const { t } = useI18n()
    return { t }
  },
  data () {
    return {
      activeName: ''
    }
  },
  created () {
    const segments = this.$route.path.split('/')
    if (segments.length !== 4) {
      this.activeName = 'email'
      return
    }
    this.activeName = segments[3]
  },
  watch: {
    activeName (newVal) {
      if (newVal && this.$route.path !== `/system/notification/${newVal}`) {
        this.$router.push(`/system/notification/${newVal}`)
      }
    }
  }
}
</script>
