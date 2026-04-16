import httpClient from '../utils/httpClient'

export default {
  list(query, callback) {
    httpClient.batchGet([{ uri: '/task', params: query }, { uri: '/host/all' }], callback)
  },

  detail(id, callback) {
    if (!id) {
      httpClient.get('/host/all', {}, hosts => {
        callback(null, hosts)
      })
      return
    }
    httpClient.batchGet([{ uri: `/task/${id}` }, { uri: '/host/all' }], callback)
  },

  update(data, callback) {
    httpClient.post('/task/store', data, callback)
  },

  remove(id, callback) {
    httpClient.post(`/task/remove/${id}`, {}, callback)
  },

  enable(id, callback) {
    httpClient.post(`/task/enable/${id}`, {}, callback)
  },

  disable(id, callback) {
    httpClient.post(`/task/disable/${id}`, {}, callback)
  },

  run(id, callback) {
    httpClient.get(`/task/run/${id}`, { _t: Date.now() }, callback)
  },

  allTags(callback) {
    httpClient.get('/task/tags', {}, callback)
  },

  batchEnable(ids, callback) {
    httpClient.postJson('/task/batch-enable', { ids }, callback)
  },

  batchDisable(ids, callback) {
    httpClient.postJson('/task/batch-disable', { ids }, callback)
  },

  batchRemove(ids, callback) {
    httpClient.postJson('/task/batch-remove', { ids }, callback)
  },

  versions(taskId, params, callback) {
    httpClient.get(`/task/versions/${taskId}`, params, callback)
  },

  versionDetail(taskId, versionId, callback) {
    httpClient.get(`/task/versions/${taskId}/${versionId}`, {}, callback)
  },

  versionRollback(taskId, versionId, callback) {
    httpClient.post(`/task/versions/${taskId}/${versionId}/rollback`, {}, callback)
  }
}
