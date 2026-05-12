import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const occupationApi = {
  getAll: () => api.get('/occupations'),
  getById: (id) => api.get(`/occupations/${id}`),
  create: (data) => api.post('/occupations', data),
  update: (id, data) => api.put(`/occupations/${id}`, data),
  delete: (id) => api.delete(`/occupations/${id}`),
};

export const batchApi = {
  getAll: () => api.get('/batches'),
  getById: (id) => api.get(`/batches/${id}`),
  create: (data) => api.post('/batches', data),
  delete: (id) => api.delete(`/batches/${id}`),
  nextStatus: (id) => api.post(`/batches/${id}/status`, { action: 'next' }),
  registerCandidate: (id, data) => api.post(`/batches/${id}/register`, data),
  enterScores: (id, data) => api.post(`/batches/${id}/scores`, data),
};

export const certificateApi = {
  getAll: () => api.get('/certificates'),
  getById: (id) => api.get(`/certificates/${id}`),
  issue: (data) => api.post('/certificates', data),
  updateStatus: (id, status) => api.put(`/certificates/${id}/status`, { status }),
  getExpiringSoon: () => api.get('/certificates/expiring-soon'),
};

export const statsApi = {
  getPassRate: () => api.get('/stats/pass-rate'),
};

export default api;
