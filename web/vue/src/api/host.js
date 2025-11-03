import httpClient from '../utils/httpClient'

export default {
  // 任务列表
  list (query, callback) {
    httpClient.get('/host', query, callback)
  },

  all (query, callback) {
    httpClient.get('/host/all', {}, callback)
  },

  detail (id, callback) {
    httpClient.get(`/host/${id}`, {}, callback)
  },

  update (data, callback) {
    httpClient.post('/host/store', data, callback)
  },

  remove (id, callback) {
    httpClient.post(`/host/remove/${id}`, {}, callback)
  },

  ping (id, callback) {
    httpClient.get(`/host/ping/${id}`, {}, callback)
  },

  getRegisterToken (callback) {
    httpClient.get('/host/register-token', {}, callback)
  },

  generateRegisterToken (callback) {
    httpClient.post('/host/register-token/generate', {}, callback)
  }
}
