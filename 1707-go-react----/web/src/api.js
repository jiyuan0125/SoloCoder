const API_BASE = '/api'

export async function apiRequest(url, options = {}) {
  const response = await fetch(`${API_BASE}${url}`, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    throw new Error(errorData.error || `HTTP ${response.status}`)
  }

  if (response.status === 204) {
    return null
  }

  return response.json()
}

export const devices = {
  getAll: (params = {}) => {
    const queryString = new URLSearchParams(params).toString()
    return apiRequest(`/devices${queryString ? `?${queryString}` : ''}`)
  },
  get: (id) => apiRequest(`/devices/${id}`),
  create: (data) => apiRequest('/devices', {
    method: 'POST',
    body: JSON.stringify(data),
  }),
  updateStatus: (id, status) => apiRequest(`/devices/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  }),
  exportCSV: () => fetch(`${API_BASE}/devices/export`),
}

export const workorders = {
  getAll: (params = {}) => {
    const queryString = new URLSearchParams(params).toString()
    return apiRequest(`/workorders${queryString ? `?${queryString}` : ''}`)
  },
  get: (id) => apiRequest(`/workorders/${id}`),
  create: (data) => apiRequest('/workorders', {
    method: 'POST',
    body: JSON.stringify(data),
  }),
  updateStatus: (id, data) => apiRequest(`/workorders/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify(data),
  }),
}

export const maintenance = {
  getAll: (params = {}) => {
    const queryString = new URLSearchParams(params).toString()
    return apiRequest(`/maintenance-plans${queryString ? `?${queryString}` : ''}`)
  },
  create: (data) => apiRequest('/maintenance-plans', {
    method: 'POST',
    body: JSON.stringify(data),
  }),
  complete: (id, data) => apiRequest(`/maintenance-plans/${id}/complete`, {
    method: 'PUT',
    body: JSON.stringify(data),
  }),
  cancel: (id) => apiRequest(`/maintenance-plans/${id}/cancel`, {
    method: 'PUT',
  }),
}

export const calibration = {
  getAgencies: () => apiRequest('/calibration-agencies'),
  createAgency: (data) => apiRequest('/calibration-agencies', {
    method: 'POST',
    body: JSON.stringify(data),
  }),
  getRecords: (params = {}) => {
    const queryString = new URLSearchParams(params).toString()
    return apiRequest(`/calibration-records${queryString ? `?${queryString}` : ''}`)
  },
  getReminders: () => apiRequest('/calibration-records/reminders'),
  createRecord: (data) => apiRequest('/calibration-records', {
    method: 'POST',
    body: JSON.stringify(data),
  }),
  updateStatus: (id, data) => apiRequest(`/calibration-records/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify(data),
  }),
}
