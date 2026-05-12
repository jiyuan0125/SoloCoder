import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

export const meetingAPI = {
  getAll: () => api.get('/meetings'),
  getById: (id) => api.get(`/meetings/${id}`),
  create: (data) => api.post('/meetings', data),
  update: (id, data) => api.put(`/meetings/${id}`, data),
  delete: (id) => api.delete(`/meetings/${id}`),
  getPapers: (id) => api.get(`/meetings/${id}/papers`),
  exportPapers: (id) => api.get(`/meetings/${id}/papers/export`),
  getSessions: (id) => api.get(`/meetings/${id}/sessions`),
  createSession: (id, data) => api.post(`/meetings/${id}/sessions`, data),
  getSchedule: (id) => api.get(`/meetings/${id}/schedule`),
};

export const paperAPI = {
  getById: (id) => api.get(`/papers/${id}`),
  create: (data) => api.post('/papers', data),
  submit: (id) => api.post(`/papers/${id}/submit`),
  updateStatus: (id, status) => api.put(`/papers/${id}/status`, { status }),
};

export const reviewAPI = {
  assign: (data) => api.post('/reviews/assign', data),
  submit: (id, data) => api.put(`/reviews/${id}`, data),
  getByReviewer: (id) => api.get(`/reviews/reviewer/${id}`),
};

export const scheduleAPI = {
  create: (data) => api.post('/schedules', data),
  delete: (id) => api.delete(`/schedules/${id}`),
};

export const userAPI = {
  getAll: () => api.get('/users'),
  getReviewers: () => api.get('/users/reviewers'),
  create: (data) => api.post('/users', data),
  getReminders: (id) => api.get(`/users/${id}/reminders`),
  markReminderRead: (id) => api.put(`/users/reminders/${id}/read`),
};

export default api;
