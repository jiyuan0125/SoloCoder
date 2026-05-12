import axios from 'axios';

const API_BASE_URL = 'http://localhost:8300/api';

const api = axios.create({
  baseURL: API_BASE_URL,
});

export const protocolAPI = {
  list: () => api.get('/protocols'),
  get: (id) => api.get(`/protocols/${id}`),
  create: (data) => api.post('/protocols', data),
  update: (id, data) => api.put(`/protocols/${id}`, data),
  addSite: (id, data) => api.post(`/protocols/${id}/sites`, data),
  addVisit: (id, data) => api.post(`/protocols/${id}/visits`, data),
};

export const subjectAPI = {
  list: (params) => api.get('/subjects', { params }),
  get: (id) => api.get(`/subjects/${id}`),
  enroll: (data) => api.post('/subjects', data),
  updateStatus: (id, status) => api.put(`/subjects/${id}/status`, { status }),
  withdraw: (id) => api.post(`/subjects/${id}/withdraw`),
};

export const visitAPI = {
  record: (data) => api.post('/visits/records', data),
  list: (params) => api.get('/visits/records', { params }),
};

export const aeAPI = {
  list: (params) => api.get('/aes', { params }),
  get: (id) => api.get(`/aes/${id}`),
  create: (data) => api.post('/aes', data),
  update: (id, data) => api.put(`/aes/${id}`, data),
  submitReport: (id) => api.post(`/aes/${id}/submit-report`),
};

export const todoAPI = {
  list: () => api.get('/todos'),
  updateStatus: (id, status) => api.put(`/todos/${id}/status`, { status }),
};

export const exportAPI = {
  subjects: (protocolId, siteId) => 
    `${API_BASE_URL}/export/data?protocol_id=${protocolId}${siteId ? `&site_id=${siteId}` : ''}`,
  visits: (protocolId, siteId) => 
    `${API_BASE_URL}/export/visits?protocol_id=${protocolId}${siteId ? `&site_id=${siteId}` : ''}`,
  aes: (protocolId, siteId) => 
    `${API_BASE_URL}/export/data?protocol_id=${protocolId}${siteId ? `&site_id=${siteId}` : ''}&type=ae`,
  sae: (protocolId, siteId) => 
    `${API_BASE_URL}/export/data?protocol_id=${protocolId}${siteId ? `&site_id=${siteId}` : ''}&type=sae`,
};

export default api;
