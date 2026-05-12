import axios from 'axios';

const API_BASE = '/api';
const USER_ID = 'user-1';

const api = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json',
    'X-User-ID': USER_ID,
  },
});

export const projectsAPI = {
  getAll: () => api.get('/projects'),
  get: (id) => api.get(`/projects/${id}`),
  create: (data) => api.post('/projects', data),
  update: (id, data) => api.put(`/projects/${id}`, data),
  updateStatus: (id, status) => api.put(`/projects/${id}/status`, { status }),
  delete: (id) => api.delete(`/projects/${id}`),
};

export const tasksAPI = {
  getAll: () => api.get('/tasks'),
  getByProject: (projectId) => api.get(`/tasks/project/${projectId}`),
  get: (id) => api.get(`/tasks/${id}`),
  create: (data) => api.post('/tasks', data),
  update: (id, data) => api.put(`/tasks/${id}`, data),
  updateStatus: (id, status) => api.put(`/tasks/${id}/status`, { status }),
};

export const dataAPI = {
  getAll: (params = {}) => api.get('/data', { params }),
  get: (id) => api.get(`/data/${id}`),
  upload: (data) => api.post('/data', data),
  download: (id) => api.post(`/data/${id}/download`),
  getLogs: (id) => api.get(`/data/${id}/logs`),
};

export const achievementsAPI = {
  getAll: () => api.get('/achievements'),
  get: (id) => api.get(`/achievements/${id}`),
  create: (data) => api.post('/achievements', data),
  update: (id, data) => api.put(`/achievements/${id}`, data),
  export: () => api.get('/achievements/export'),
};

export const usersAPI = {
  getAll: () => api.get('/users'),
};

export const departmentsAPI = {
  getAll: () => api.get('/departments'),
};

export const approvalsAPI = {
  get: (id) => api.get(`/approvals/${id}`),
  create: (data) => api.post('/approvals', data),
  process: (id, action, data) => api.post(`/approvals/${id}/process`, { action, ...data }),
};

export const todosAPI = {
  getMy: () => api.get('/todos'),
};

export default api;
