import axios from 'axios'
import { useNotify } from '../composables/useNotify'
import router from '../router/index'
import { useUserStore } from '../stores/user'
import qs from 'qs'
import NProgress from '@/utils/progress'

const getLocale = () => localStorage.getItem('locale') || 'zh-CN'

const messages = {
  'zh-CN': {
    loadFailed: '加载失败, 请稍后再试',
    requestTimeout: '请求超时，请稍后重试',
    authExpired: '登录已过期，请重新登录',
    requestFailed: '请求失败'
  },
  'en-US': {
    loadFailed: 'Load failed, please try again later',
    requestTimeout: 'Request timeout, please try again later',
    authExpired: 'Login expired, please login again',
    requestFailed: 'Request failed'
  }
}

const t = key => {
  const locale = getLocale()
  return messages[locale]?.[key] || messages['zh-CN'][key]
}
// 成功状态码
const SUCCESS_CODE = 0
// 认证失败
const AUTH_ERROR_CODE = 401
// 应用未安装
const APP_NOT_INSTALL_CODE = 801

axios.defaults.baseURL = '/api'
axios.defaults.timeout = 30000
axios.defaults.responseType = 'json'
axios.interceptors.request.use(
  config => {
    NProgress.start()
    const userStore = useUserStore()
    config.headers['Auth-Token'] = userStore.token
    config.headers['Accept-Language'] = localStorage.getItem('locale') || 'zh-CN'
    return config
  },
  error => {
    NProgress.done()
    const notify = useNotify()
    notify.error(t('loadFailed'))

    return Promise.reject(error)
  }
)

axios.interceptors.response.use(
  data => {
    NProgress.done()
    // 检查是否有新的 token
    const newToken = data.headers['new-auth-token']
    if (newToken) {
      const userStore = useUserStore()
      userStore.token = newToken
    }
    return data
  },
  error => {
    NProgress.done()
    const notify = useNotify()
    // 处理超时
    if (error.code === 'ECONNABORTED' && error.message.includes('timeout')) {
      notify.error(t('requestTimeout'))
      return Promise.reject(error)
    }

    // 处理认证失败
    if (error.response && error.response.status === 401) {
      const userStore = useUserStore()
      userStore.token = ''
      notify.warning(t('authExpired'))
      setTimeout(() => {
        window.location.href = '/'
      }, 500)
      return Promise.reject(error)
    }

    notify.error(t('loadFailed'))

    return Promise.reject(error)
  }
)

function handle(promise, next, errorCallback) {
  promise
    .then(res => successCallback(res, next, errorCallback))
    .catch(error => failureCallback(error))
}

function checkResponseCode(code, msg) {
  const notify = useNotify()
  switch (code) {
    // 应用未安装
    case APP_NOT_INSTALL_CODE:
      router.push('/install')
      return false
    // 认证失败
    case AUTH_ERROR_CODE: {
      const userStore = useUserStore()
      userStore.token = ''
      notify.warning(t('authExpired'))
      setTimeout(() => {
        window.location.href = '/'
      }, 500)
      return false
    }
  }
  if (code !== SUCCESS_CODE) {
    notify.error(msg)
    return false
  }

  return true
}

function successCallback(res, next, errorCallback) {
  if (res.data.code !== SUCCESS_CODE) {
    if (errorCallback) {
      errorCallback(res.data.code, res.data.message)
      return
    }
    if (!checkResponseCode(res.data.code, res.data.message)) {
      return
    }
  }
  if (!next) {
    return
  }
  next(res.data.data, res.data.code, res.data.message)
}

function failureCallback(error) {
  // 避免重复提示（已在 interceptor 中处理）
  if (error.response && error.response.status === 401) {
    return
  }
  if (error.code === 'ECONNABORTED') {
    return
  }
  const notify = useNotify()
  notify.error(t('requestFailed') + ' - ' + error.message)
}

export default {
  get(uri, params, next) {
    const promise = axios.get(uri, { params })
    handle(promise, next)
  },

  batchGet(uriGroup, next) {
    const requests = []
    for (let item of uriGroup) {
      let params = {}
      if (item.params !== undefined) {
        params = item.params
      }
      requests.push(axios.get(item.uri, { params }))
    }

    Promise.all(requests)
      .then(function (res) {
        const result = []
        for (let item of res) {
          if (!checkResponseCode(item.data.code, item.data.message)) {
            return
          }
          result.push(item.data.data)
        }
        next(...result)
      })
      .catch(error => failureCallback(error))
  },

  post(uri, data, next, errorCallback) {
    const promise = axios.post(uri, qs.stringify(data), {
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded'
      }
    })
    handle(promise, next, errorCallback)
  },

  postJson(uri, data, next, errorCallback) {
    const promise = axios.post(uri, data, {
      headers: {
        'Content-Type': 'application/json'
      }
    })
    handle(promise, next, errorCallback)
  }
}
