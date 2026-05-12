const API_BASE = '/api'

const request = async (url, options = {}) => {
  const response = await fetch(API_BASE + url, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers
    },
    ...options
  })
  
  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: '请求失败' }))
    throw new Error(error.error || '请求失败')
  }
  
  return response.json()
}

export const donorsAPI = {
  list: (page = 1, size = 20) => request(`/donors?page=${page}&size=${size}`),
  get: (id) => request(`/donors/${id}`),
  create: (data) => request('/donors', { method: 'POST', body: JSON.stringify(data) })
}

export const recipientsAPI = {
  list: (page = 1, size = 20) => request(`/recipients?page=${page}&size=${size}`),
  get: (id) => request(`/recipients/${id}`),
  create: (data) => request('/recipients', { method: 'POST', body: JSON.stringify(data) })
}

export const organsAPI = {
  assess: (id, data) => request(`/organs/${id}/assess`, { method: 'POST', body: JSON.stringify(data) }),
  startMatching: (id) => request(`/organs/${id}/match`, { method: 'POST' }),
  confirmMatch: (id, data) => request(`/organs/${id}/confirm`, { method: 'POST', body: JSON.stringify(data) }),
  acquire: (id, data) => request(`/organs/${id}/acquire`, { method: 'POST', body: JSON.stringify(data) })
}

export const transplantsAPI = {
  list: (page = 1, size = 20) => request(`/transplants?page=${page}&size=${size}`),
  get: (id) => request(`/transplants/${id}`),
  create: (data) => request('/transplants', { method: 'POST', body: JSON.stringify(data) }),
  updateStatus: (id, status) => request(`/transplants/${id}/status`, { method: 'PUT', body: JSON.stringify({ status }) }),
  addFollowUp: (id, data) => request(`/transplants/${id}/followup`, { method: 'POST', body: JSON.stringify(data) })
}

export const todosAPI = {
  list: (page = 1, size = 50) => request(`/todos?page=${page}&size=${size}`),
  updateStatus: (id, status) => request(`/todos/${id}/status`, { method: 'PUT', body: JSON.stringify({ status }) })
}

export const locationsAPI = {
  list: () => request('/locations'),
  create: (data) => request('/locations', { method: 'POST', body: JSON.stringify(data) })
}

export const personnelAPI = {
  list: () => request('/personnel'),
  create: (data) => request('/personnel', { method: 'POST', body: JSON.stringify(data) })
}

export const reportsAPI = {
  list: () => request('/reports'),
  generate: () => request('/reports/generate', { method: 'POST' })
}

export const adminAPI = {
  upgradeUrgency: () => request('/upgrade-urgency', { method: 'POST' })
}
