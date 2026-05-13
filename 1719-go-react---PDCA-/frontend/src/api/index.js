import axios from 'axios'

const api = axios.create({
  baseURL: 'http://localhost:8501/api',
  timeout: 10000,
})

export const indicators = {
  list: () => api.get('/indicators'),
  get: (id) => api.get(`/indicators/${id}`),
  create: (data) => api.post('/indicators', data),
  update: (id, data) => api.put(`/indicators/${id}`, data),
  updateTarget: (id, data) => api.put(`/indicators/${id}/target`, data),
  delete: (id) => api.delete(`/indicators/${id}`),
  listData: (id) => api.get(`/indicators/${id}/data`),
  createData: (id, data) => api.post(`/indicators/${id}/data`, data),
  getTrend: (id, months = 12) => api.get(`/indicators/${id}/trend?months=${months}`),
}

export const pdca = {
  list: () => api.get('/pdca'),
  get: (id) => api.get(`/pdca/${id}`),
  create: (data) => api.post('/pdca', data),
  update: (id, data) => api.put(`/pdca/${id}`, data),
  delete: (id) => api.delete(`/pdca/${id}`),
  createPhase: (id, data) => api.post(`/pdca/${id}/phases`, data),
  updatePhase: (id, phase, data) => api.put(`/pdca/${id}/phases/${phase}`, data),
  nextPhase: (id) => api.post(`/pdca/${id}/next-phase`),
  startNextCycle: (id) => api.post(`/pdca/${id}/start-next-cycle`),
}

export const todos = {
  list: (status, type) => {
    const params = new URLSearchParams()
    if (status) params.append('status', status)
    if (type) params.append('type', type)
    return api.get(`/todos?${params.toString()}`)
  },
  get: (id) => api.get(`/todos/${id}`),
  update: (id, data) => api.put(`/todos/${id}`, data),
  updateStatus: (id, status) => api.put(`/todos/${id}/status`, { status }),
  delete: (id) => api.delete(`/todos/${id}`),
}

export const meetings = {
  list: () => api.get('/meetings'),
  get: (id) => api.get(`/meetings/${id}`),
  create: (data) => api.post('/meetings', data),
  update: (id, data) => api.put(`/meetings/${id}`, data),
  delete: (id) => api.delete(`/meetings/${id}`),
}

export const aggregations = {
  byDepartment: (month) => {
    const params = month ? `?month=${month}` : ''
    return api.get(`/aggregations/department${params}`)
  },
  byCategory: (from, to) => {
    const params = new URLSearchParams()
    if (from) params.append('from', from)
    if (to) params.append('to', to)
    return api.get(`/aggregations/category?${params.toString()}`)
  },
  byTime: (period, from, to) => {
    const params = new URLSearchParams()
    if (period) params.append('period', period)
    if (from) params.append('from', from)
    if (to) params.append('to', to)
    return api.get(`/aggregations/time?${params.toString()}`)
  },
}

export default api
