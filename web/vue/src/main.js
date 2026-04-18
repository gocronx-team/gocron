import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import { ElMessageBox, ElMessage } from 'element-plus'
import 'element-plus/dist/index.css'
import './assets/tailwind.css'
import 'nprogress/nprogress.css'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import timezone from 'dayjs/plugin/timezone'
import App from './App.vue'
import router from './router'
import i18n from './locales'

dayjs.extend(utc)
dayjs.extend(timezone)

const app = createApp(App)
const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)

app.use(pinia)
app.use(router)
app.use(i18n)

app.directive('focus', {
  mounted(el) {
    el.focus()
  }
})

app.config.globalProperties.$appConfirm = function (callback) {
  ElMessageBox.confirm(i18n.global.t('common.confirmOperation'), i18n.global.t('common.tip'), {
    confirmButtonText: i18n.global.t('common.confirm'),
    cancelButtonText: i18n.global.t('common.cancel'),
    type: 'warning',
    center: true,
    customClass: 'custom-message-box'
  })
    .then(() => {
      callback()
    })
    .catch(() => {})
}

app.config.globalProperties.$message = ElMessage

app.config.globalProperties.$filters = {
  formatTime(time) {
    if (!time) return ''
    return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
  }
}

// 全局错误处理
app.config.errorHandler = (err, instance, info) => {
  if (import.meta.env.DEV) {
    console.error('[Global Error]', err, info)
  }
  ElMessage.error('系统错误，请刷新页面重试')
}

app.config.warnHandler = (msg, instance, trace) => {
  if (import.meta.env.DEV) {
    console.warn('[Vue Warn]', msg, trace)
  }
}

// 开发环境性能监控
if (import.meta.env.DEV) {
  app.config.performance = true
}

// 生产环境禁用 devtools
if (import.meta.env.PROD) {
  app.config.devtools = false
}

app.mount('#app')
