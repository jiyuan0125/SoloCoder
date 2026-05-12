import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

export const unitAPI = {
  list: (params = {}) => api.get('/units', { params }),
  get: (id) => api.get(`/units/${id}`),
  create: (data) => api.post('/units', data),
  update: (id, data) => api.put(`/units/${id}`, data),
  delete: (id) => api.delete(`/units/${id}`),
  getWarnings: () => api.get('/units/warnings'),
};

export const inspectionAPI = {
  list: (params = {}) => api.get('/inspections', { params }),
  get: (id) => api.get(`/inspections/${id}`),
  create: (data) => api.post('/inspections', data),
  updateItems: (id, data) => api.put(`/inspections/${id}/items`, data),
  delete: (id) => api.delete(`/inspections/${id}`),
  getTemplates: (unitType) => api.get('/inspections/templates', { params: { unit_type: unitType } }),
};

export const opinionAPI = {
  list: (params = {}) => api.get('/opinions', { params }),
  get: (id) => api.get(`/opinions/${id}`),
  create: (data) => api.post('/opinions', data),
  review: (id, data) => api.post(`/opinions/${id}/review`, data),
};

export const penaltyAPI = {
  list: (params = {}) => api.get('/penalties', { params }),
  get: (id) => api.get(`/penalties/${id}`),
  create: (data) => api.post('/penalties', data),
  getFineRanges: () => api.get('/penalties/fine-ranges'),
};

export const noticeAPI = {
  list: (params = {}) => api.get('/notices', { params }),
  get: (id) => api.get(`/notices/${id}`),
  create: (data) => api.post('/notices', data),
  updateExpiry: (id, data) => api.put(`/notices/${id}/expiry`, data),
  delete: (id) => api.delete(`/notices/${id}`),
  checkExpiry: () => api.post('/notices/check-expiry'),
};

export const auditAPI = {
  list: (params = {}) => api.get('/audit', { params }),
  get: (id) => api.get(`/audit/${id}`),
};

export default api;
