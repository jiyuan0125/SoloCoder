import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const callService = {
  receiveCall: (data) => api.post('/calls', data),
  listCalls: () => api.get('/calls'),
  getPendingCalls: () => api.get('/calls/pending'),
  acceptCall: (id) => api.post(`/calls/${id}/accept`),
  processCall: (id) => api.post(`/calls/${id}/process`),
  submitForReview: (id) => api.post(`/calls/${id}/submit-review`),
  completeCall: (id) => api.post(`/calls/${id}/complete`),
  rejectCall: (id) => api.post(`/calls/${id}/reject`),
};

export const vehicleService = {
  listVehicles: () => api.get('/vehicles'),
  createVehicle: (data) => api.post('/vehicles', data),
  updateStatus: (id, status) => api.put(`/vehicles/${id}/status`, { status }),
  setMaintenance: (id) => api.post(`/vehicles/${id}/maintenance`),
};

export const dispatchService = {
  recommendVehicle: (callId) => api.get(`/dispatch/recommend/${callId}`),
  dispatchVehicle: (data) => api.post('/dispatch', data),
  listRecords: () => api.get('/dispatch/records'),
};

export const triageService = {
  createTriage: (data) => api.post('/triage', data),
  listRecords: () => api.get('/triage/records'),
};

export const statsService = {
  getStats: () => api.get('/stats'),
};

export default api;
