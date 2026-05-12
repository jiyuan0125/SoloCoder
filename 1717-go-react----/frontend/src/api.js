import axios from 'axios';

const API_BASE = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const infectionAPI = {
  getAll: (params) => api.get('/infections', { params }),
  get: (id) => api.get(`/infections/${id}`),
  create: (data) => api.post('/infections', data),
  update: (id, data) => api.put(`/infections/${id}`, data),
  delete: (id) => api.delete(`/infections/${id}`),
};

export const measureAPI = {
  getAll: (params) => api.get('/measures', { params }),
  get: (id) => api.get(`/measures/${id}`),
  create: (data) => api.post('/measures', data),
  delete: (id) => api.delete(`/measures/${id}`),
  getRecommended: (site) => api.get(`/measures/recommended?site=${site}`),
};

export const monitoringAPI = {
  getAll: (params) => api.get('/monitoring', { params }),
  get: (id) => api.get(`/monitoring/${id}`),
  create: (data) => api.post('/monitoring', data),
  delete: (id) => api.delete(`/monitoring/${id}`),
  getTargetDepartments: () => api.get('/monitoring/departments'),
};

export const statisticsAPI = {
  getRates: (month) => api.get(`/statistics/rates${month ? `?month=${month}` : ''}`),
  getTrend: () => api.get('/statistics/trend'),
  getAlerts: (params) => api.get('/statistics/alerts', { params }),
};

export const reportAPI = {
  getAll: (params) => api.get('/reports', { params }),
  get: (id) => api.get(`/reports/${id}`),
  generate: (data) => api.post('/reports/generate', data),
  approve: (id, action, data) => api.post(`/reports/${id}/approve/${action}`, data),
  getHistory: (id) => api.get(`/reports/${id}/history`),
  delete: (id) => api.delete(`/reports/${id}`),
};

export const departmentAPI = {
  getAll: () => api.get('/departments'),
};

export default api;
