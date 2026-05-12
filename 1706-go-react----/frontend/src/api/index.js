import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

export const drugApi = {
  list: (search = '') => api.get('/drugs', { params: { search } }),
  get: (id) => api.get(`/drugs/${id}`),
  create: (data) => api.post('/drugs', data),
  update: (id, data) => api.put(`/drugs/${id}`, data),
  remove: (id) => api.delete(`/drugs/${id}`),
}

export const categoryApi = {
  list: () => api.get('/categories'),
  create: (data) => api.post('/categories', data),
}

export const stockApi = {
  in: (data) => api.post('/stock/in', data),
  out: (data) => api.post('/stock/out', data),
  items: (drugId) => api.get(`/stock/items/${drugId}`),
  alerts: () => api.get('/stock/alerts'),
  markAlertRead: (id) => api.post(`/stock/alerts/${id}/read`),
  nearExpiry: () => api.get('/stock/near-expiry'),
  value: () => api.get('/stock/value'),
  transactions: (start, end) => api.get('/stock/transactions', { params: { start, end } }),
}

export const prescriptionApi = {
  list: (status) => api.get('/prescriptions', { params: { status } }),
  get: (id) => api.get(`/prescriptions/${id}`),
  create: (data) => api.post('/prescriptions', data),
  approve: (id, data) => api.post(`/prescriptions/${id}/approve`, data),
  reject: (id, data) => api.post(`/prescriptions/${id}/reject`, data),
  cancel: (id) => api.post(`/prescriptions/${id}/cancel`),
}

export default api
