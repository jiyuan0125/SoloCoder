import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

export const outbreakApi = {
  createReport: (data) => api.post('/outbreak/reports', data),
  listReports: () => api.get('/outbreak/reports'),
  getAlerts: () => api.get('/outbreak/alerts'),
}

export const investigationApi = {
  create: (data) => api.post('/investigations', data),
  list: () => api.get('/investigations'),
}

export const contactApi = {
  create: (data) => api.post('/contacts', data),
  list: (investigationId) => api.get('/contacts', { params: { investigationId } }),
  updateStatus: (id, status) => api.put(`/contacts/${id}/status`, { status }),
}

export const vaccineApi = {
  listVaccines: () => api.get('/vaccines'),
  createVaccine: (data) => api.post('/vaccines', data),
  recordVaccination: (data) => api.post('/vaccinations', data),
  listRecords: (recipientId) => api.get('/vaccinations', { params: { recipientId } }),
  listAllRecords: () => api.get('/vaccinations'),
}

export const todoApi = {
  list: () => api.get('/todos'),
  updateStatus: (id, status) => api.put(`/todos/${id}/status`, { status }),
}

export const statsApi = {
  getStatistics: () => api.get('/statistics'),
  listWeeklyReports: () => api.get('/weekly-reports'),
  generateWeeklyReport: () => api.post('/weekly-reports/generate'),
  listDiseases: () => api.get('/diseases'),
}

export default api
