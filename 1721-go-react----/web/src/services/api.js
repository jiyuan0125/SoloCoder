import axios from 'axios'

const API_BASE = '/api'

export const courseAPI = {
  getAll: (params = {}) => axios.get(`${API_BASE}/courses`, { params }),
  getById: (id) => axios.get(`${API_BASE}/courses/${id}`),
  create: (data) => axios.post(`${API_BASE}/courses`, data),
  updateStatus: (id, status) => axios.put(`${API_BASE}/courses/${id}/status`, { status }),
  getStats: (id) => axios.get(`${API_BASE}/courses/${id}/stats`),
  getReviews: (id) => axios.get(`${API_BASE}/courses/${id}/reviews`),
  createVote: (id, data) => axios.post(`${API_BASE}/courses/${id}/vote`, data),
  getVotes: (id) => axios.get(`${API_BASE}/courses/${id}/votes`),
}

export const studentAPI = {
  getAll: () => axios.get(`${API_BASE}/students`),
  getById: (id) => axios.get(`${API_BASE}/students/${id}`),
  register: (data) => axios.post(`${API_BASE}/students`, data),
  getEnrollments: (id) => axios.get(`${API_BASE}/students/${id}/enrollments`),
}

export const progressAPI = {
  getByStudent: (studentId) => axios.get(`${API_BASE}/progress/${studentId}`),
  update: (studentId, courseId, watchedSeconds) => 
    axios.post(`${API_BASE}/progress/${studentId}/${courseId}`, { watched_seconds: watchedSeconds }),
}

export const reviewAPI = {
  create: (data) => axios.post(`${API_BASE}/reviews`, data),
}

export const paymentAPI = {
  recharge: (studentId, cardNumber, password) => 
    axios.post(`${API_BASE}/recharge`, { student_id: studentId, card_number: cardNumber, password }),
  enroll: (studentId, courseId) => 
    axios.post(`${API_BASE}/enroll`, { student_id: studentId, course_id: courseId }),
}

export const statsAPI = {
  getSubjects: () => axios.get(`${API_BASE}/stats/subjects`),
  getInstructors: () => axios.get(`${API_BASE}/stats/instructors`),
  getLiveStats: () => axios.get(`${API_BASE}/stats/live`),
}

export const subjects = [
  { id: 'chinese', name: '语文' },
  { id: 'math', name: '数学' },
  { id: 'english', name: '英语' },
  { id: 'physics', name: '物理' },
  { id: 'chemistry', name: '化学' },
  { id: 'biology', name: '生物' },
  { id: 'history', name: '历史' },
  { id: 'geography', name: '地理' },
  { id: 'politics', name: '政治' },
]

export const courseTypes = [
  { id: 'live', name: '直播课' },
  { id: 'record', name: '录播课' },
]
