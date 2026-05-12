import axios from 'axios';

const API_BASE = 'http://localhost:9090/api';

const api = axios.create({
  baseURL: API_BASE,
  timeout: 10000,
});

export const pathsApi = {
  list: () => api.get('/r'),
  get: (id) => api.get(`/r/${id}`),
  create: (data) => api.post('/r', data),
  update: (id, data) => api.put(`/r/${id}`, data),
  listStages: (id) => api.get(`/r/${id}/stages`),
  createStage: (id, data) => api.post(`/r/${id}/stages`, data),
  listItems: (id) => api.get(`/r/${id}/items`),
  createItem: (id, data) => api.post(`/r/${id}/items`, data),
};

export const patientsApi = {
  list: () => api.get('/patients'),
  create: (data) => api.post('/patients', data),
};

export const enrollmentsApi = {
  list: () => api.get('/enrollments'),
  get: (id) => api.get(`/enrollments/${id}`),
  create: (data) => api.post('/enrollments', data),
  listOrders: (id) => api.get(`/enrollments/${id}/orders`),
  listVariations: (id) => api.get(`/enrollments/${id}/variations`),
  createVariation: (id, data) => api.post(`/enrollments/${id}/variations`, data),
  exit: (id, data) => api.post(`/enrollments/${id}/exit`, data),
  complete: (id, data) => api.post(`/enrollments/${id}/complete`, data),
};

export const ordersApi = {
  execute: (id, executed) => api.put(`/orders/${id}`, { executed }),
};

export const qualityApi = {
  getMetrics: (year, month) => api.get('/quality', { params: { year, month } }),
};

export default api;
