import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import { ElMessageBox, ElMessage } from 'element-plus'
import App from './App.vue'
import router from './router'

const app = createApp(App)
const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)

app.use(pinia)
app.use(router)

app.directive('focus', {
  mounted(el) {
    el.focus()
  }
})

app.config.globalProperties.$appConfirm = function (callback) {
  ElMessageBox.confirm('确定执行此操作?', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
    center: true,
    customClass: 'custom-message-box'
  }).then(() => {
    callback()
  }).catch(() => {})
}

app.config.globalProperties.$message = ElMessage

app.config.globalProperties.$filters = {
  formatTime(time) {
    if (!time) return ''
    const fillZero = (num) => (num >= 10 ? num : '0' + num)
    const date = new Date(time)
    const result = `${date.getFullYear()}-${fillZero(date.getMonth() + 1)}-${fillZero(date.getDate())} ${fillZero(date.getHours())}:${fillZero(date.getMinutes())}:${fillZero(date.getSeconds())}`
    return result.indexOf('20') === 0 ? result : ''
  }
}

app.mount('#app')
