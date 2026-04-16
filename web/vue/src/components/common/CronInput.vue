<template>
  <el-input
    v-model.trim="innerValue"
    :placeholder="t('task.cronPlaceholder')"
    @input="onInput">
    <template #append>
      <el-popover
        placement="bottom"
        :width="500"
        trigger="click">
        <template #reference>
          <el-button>{{ t('task.cronExample') }}</el-button>
        </template>
        <div>
          <h4>{{ t('task.cronStandard') }}</h4>
          <ul style="padding-left: 20px; margin: 10px 0;">
            <li>0 * * * * * - {{ t('message.everyMinute') }}</li>
            <li>*/20 * * * * * - {{ t('message.every20Seconds') }}</li>
            <li>0 30 21 * * * - {{ t('message.everyDay21_30') }}</li>
            <li>0 0 23 * * 6 - {{ t('message.everySaturday23') }}</li>
          </ul>
          <h4>{{ t('task.cronShortcut') }}</h4>
          <ul style="padding-left: 20px; margin: 10px 0;">
            <li>@reboot - {{ t('message.reboot') }}</li>
            <li>@yearly - {{ t('message.yearly') }}</li>
            <li>@monthly - {{ t('message.monthly') }}</li>
            <li>@weekly - {{ t('message.weekly') }}</li>
            <li>@daily - {{ t('message.daily') }}</li>
            <li>@hourly - {{ t('message.hourly') }}</li>
            <li>@every 30s - {{ t('message.every30s') }}</li>
            <li>@every 1m20s - {{ t('message.every1m20s') }}</li>
          </ul>
        </div>
      </el-popover>
    </template>
  </el-input>
</template>

<script>
import { useI18n } from 'vue-i18n'

export default {
  name: 'CronInput',
  props: {
    modelValue: {
      type: String,
      default: ''
    }
  },
  emits: ['update:modelValue'],
  setup() {
    const { t } = useI18n()
    return { t }
  },
  computed: {
    innerValue: {
      get() {
        return this.modelValue
      },
      set(val) {
        this.$emit('update:modelValue', val)
      }
    }
  },
  methods: {
    onInput(val) {
      this.$emit('update:modelValue', val)
    }
  }
}
</script>
