import axios from 'axios'
import { useNotify } from '../composables/useNotify'
import router from '../router'
import { useUserStore } from '../stores/user'

const SUCCESS_CODE = 0
const AUTH_ERROR_CODE = 401
const APP_NOT_INSTALL_CODE = 801

// 请求取消管理
const pendingRequests = new Map()

const request = axios.create({
  baseURL: '/api',
  timeout: 10000,
  withCredentials: false,
  headers: {
    'X-Requested-With': 'XMLHttpRequest'
  }
})

request.interceptors.request.use(
  config => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers['Auth-Token'] = userStore.token
    }

    // 取消重复请求
    const requestKey = `${config.method}_${config.url}`
    if (pendingRequests.has(requestKey)) {
      const controller = pendingRequests.get(requestKey)
      controller.abort()
    }
    const controller = new AbortController()
    config.signal = controller.signal
    pendingRequests.set(requestKey, controller)

    return config
  },
  error => {
    const notify = useNotify()
    notify.error('请求失败')
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  response => {
    // 清除已完成的请求
    const requestKey = `${response.config.method}_${response.config.url}`
    pendingRequests.delete(requestKey)

    const { code, message, data } = response.data

    if (code === APP_NOT_INSTALL_CODE) {
      router.push('/install')
      return Promise.reject(new Error(message))
    }

    if (code === AUTH_ERROR_CODE) {
      const userStore = useUserStore()
      userStore.logout()
      router.push('/user/login')
      return Promise.reject(new Error(message))
    }

    if (code !== SUCCESS_CODE) {
      const notify = useNotify()
      notify.error(message || '请求失败')
      return Promise.reject(new Error(message))
    }

    return data
  },
  error => {
    // 清除失败的请求
    if (error.config) {
      const requestKey = `${error.config.method}_${error.config.url}`
      pendingRequests.delete(requestKey)
    }

    // 忽略取消的请求
    if (axios.isCancel(error)) {
      return Promise.reject(error)
    }

    const notify = useNotify()
    // 网络错误或超时
    if (error.code === 'ECONNABORTED') {
      notify.error('请求超时，请稍后重试')
    } else if (!error.response) {
      notify.error('网络连接失败，请检查网络')
    } else {
      notify.error(error.message || '请求失败')
    }
    return Promise.reject(error)
  }
)

export default request
