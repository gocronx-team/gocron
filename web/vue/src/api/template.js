import httpClient from '../utils/httpClient'

export default {
  list(query, callback) {
    httpClient.get('/template', query, callback)
  },

  categories(callback) {
    httpClient.get('/template/categories', {}, callback)
  },

  detail(id, callback) {
    httpClient.get(`/template/${id}`, {}, callback)
  },

  store(data, callback) {
    httpClient.post('/template/store', data, callback)
  },

  remove(id, callback) {
    httpClient.post(`/template/remove/${id}`, {}, callback)
  },

  apply(id, callback) {
    httpClient.post(`/template/apply/${id}`, {}, callback)
  },

  saveFromTask(data, callback) {
    httpClient.post('/template/save-from-task', data, callback)
  },
}
