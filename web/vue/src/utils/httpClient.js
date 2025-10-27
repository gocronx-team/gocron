import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router/index'
import { useUserStore } from '../stores/user'
import qs from 'qs'

const errorMessage = '加载失败, 请稍后再试'
// 成功状态码
const SUCCESS_CODE = 0
// 认证失败
const AUTH_ERROR_CODE = 401
// 应用未安装
const APP_NOT_INSTALL_CODE = 801

axios.defaults.baseURL = '/api'
axios.defaults.timeout = 10000
axios.defaults.responseType = 'json'
axios.interceptors.request.use(config => {
  const userStore = useUserStore()
  config.headers['Auth-Token'] = userStore.token
  return config
}, error => {
  ElMessage.error({
    message: errorMessage
  })

  return Promise.reject(error)
})

axios.interceptors.response.use(data => {
  return data
}, error => {
  ElMessage.error({
    message: errorMessage
  })

  return Promise.reject(error)
})

function handle (promise, next) {
  promise.then((res) => successCallback(res, next))
    .catch((error) => failureCallback(error))
}

function checkResponseCode (code, msg) {
  switch (code) {
    // 应用未安装
    case APP_NOT_INSTALL_CODE:
      router.push('/install')
      return false
    // 认证失败
    case AUTH_ERROR_CODE:
      router.push('/user/login')
      return false
  }
  if (code !== SUCCESS_CODE) {
    ElMessage.error({
      message: msg
    })
    return false
  }

  return true
}

function successCallback (res, next) {
  if (!checkResponseCode(res.data.code, res.data.message)) {
    return
  }
  if (!next) {
    return
  }
  next(res.data.data, res.data.code, res.data.message)
}

function failureCallback (error) {
  ElMessage.error({
    message: '请求失败 - ' + error
  })
}

export default {
  get (uri, params, next) {
    const promise = axios.get(uri, {params})
    handle(promise, next)
  },

  batchGet (uriGroup, next) {
    const requests = []
    for (let item of uriGroup) {
      let params = {}
      if (item.params !== undefined) {
        params = item.params
      }
      requests.push(axios.get(item.uri, {params}))
    }

    Promise.all(requests).then(function (res) {
      const result = []
      for (let item of res) {
        if (!checkResponseCode(item.data.code, item.data.message)) {
          return
        }
        result.push(item.data.data)
      }
      next(...result)
    }).catch((error) => failureCallback(error))
  },

  post (uri, data, next) {
    const promise = axios.post(uri, qs.stringify(data), {
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded'
      }
    })
    handle(promise, next)
  }
}
