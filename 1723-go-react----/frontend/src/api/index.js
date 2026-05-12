import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
})

export const courseApi = {
  getAll: () => api.get('/courses'),
  get: (id) => api.get(`/courses/${id}`),
  getGraph: () => api.get('/courses/graph'),
  create: (data) => api.post('/courses', data),
  addPrerequisite: (data) => api.post('/courses/prerequisites', data),
}

export const studentApi = {
  create: (data) => api.post('/students', data),
  startLearning: (data) => api.post('/students/start-learning', data),
  completeUnit: (data) => api.post('/students/complete-unit', data),
  getProgress: (studentId, courseId) => api.get('/students/progress', { params: { student_id: studentId, course_id: courseId } }),
  generatePath: (data) => api.post('/students/learning-path', data),
  getAnalytics: (studentId) => api.get('/students/analytics', { params: { student_id: studentId } }),
}

export const quizApi = {
  get: (studentId, courseId) => api.get('/quiz', { params: { student_id: studentId, course_id: courseId } }),
  submit: (data) => api.post('/quiz/submit', data),
}

export const productApi = {
  getAll: (page = 1, size = 20) => api.get('/r', { params: { page, size } }),
  get: (id) => api.get(`/r/${id}`),
  getSub: (id) => api.get(`/r/${id}/sub`),
  create: (data) => api.post('/r', data),
  updatePrice: (id, data) => api.put(`/r/${id}/price`, data),
}

export const orderApi = {
  getAll: (page = 1, size = 20) => api.get('/orders', { params: { page, size } }),
  get: (id) => api.get(`/orders/${id}`),
  create: (data) => api.post('/orders', data),
  confirm: (id) => api.post(`/orders/${id}/confirm`),
  getPurchaseRequests: () => api.get('/purchase-requests'),
}

export default api
