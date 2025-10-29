<template>
  <el-container>
    <system-sidebar></system-sidebar>
    <el-main>
      <h3>{{ t('system.logRetentionSettings') }}</h3>
      <el-form :model="form" :label-width="locale === 'zh-CN' ? '180px' : '220px'" style="width: 600px;">
        <el-form-item :label="t('system.dbLogRetentionDays')">
          <el-input-number v-model="form.days" :min="0" :max="3650" style="width: 200px;"></el-input-number>
          <div style="color: #909399; font-size: 12px; margin-top: 5px;">
            {{ t('system.dbLogRetentionTip') }}
          </div>
        </el-form-item>
        <el-form-item :label="t('system.cleanupTime')">
          <el-time-picker
            v-model="cleanupTime"
            format="HH:mm"
            value-format="HH:mm"
            :placeholder="t('system.selectTime')"
            style="width: 200px;">
          </el-time-picker>
          <div style="color: #909399; font-size: 12px; margin-top: 5px;">
            {{ t('system.cleanupTimeTip') }}
          </div>
        </el-form-item>
        <el-form-item :label="t('system.logFileSizeLimit')">
          <el-input-number v-model="form.fileSizeLimit" :min="0" :max="10240" style="width: 200px;"></el-input-number>
          <span style="margin-left: 10px;">MB</span>
          <div style="color: #909399; font-size: 12px; margin-top: 5px;">
            {{ t('system.logFileSizeLimitTip') }}
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="submit">{{ t('common.save') }}</el-button>
        </el-form-item>
      </el-form>
    </el-main>
  </el-container>
</template>

<script>
import { useI18n } from 'vue-i18n'
import systemSidebar from './sidebar.vue'
import httpClient from '../../utils/httpClient'

export default {
  name: 'log-retention',
  components: { systemSidebar },
  setup() {
    const { t, locale } = useI18n()
    return { t, locale }
  },
  data() {
    return {
      form: {
        days: 0,
        fileSizeLimit: 0
      },
      cleanupTime: '03:00'
    }
  },
  created() {
    this.loadData()
  },
  methods: {
    loadData() {
      httpClient.get('/system/log-retention', {}, (data) => {
        this.form.days = data.days
        this.form.fileSizeLimit = data.file_size_limit || 0
        this.cleanupTime = data.cleanup_time || '03:00'
      })
    },
    submit() {
      httpClient.postJson('/system/log-retention', { 
        days: this.form.days,
        cleanup_time: this.cleanupTime,
        file_size_limit: this.form.fileSizeLimit
      }, () => {
        this.$message.success(this.t('system.logRetentionSaveSuccess'))
      })
    }
  }
}
</script>
