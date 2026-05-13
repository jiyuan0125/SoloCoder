const API_BASE = 'http://localhost:8430/api';

async function request(path, options = {}) {
  const res = await fetch(API_BASE + path, {
    headers: { 'Content-Type': 'application/json', ...options.headers },
    ...options,
  });
  if (!res.ok) {
    const text = await res.text();
    let errMsg = `HTTP ${res.status}`;
    try {
      const data = JSON.parse(text);
      errMsg = data.error || errMsg;
    } catch (e) {}
    throw new Error(errMsg);
  }
  const text = await res.text();
  if (!text) return null;
  return JSON.parse(text);
}

export const api = {
  enterprises: {
    list: (page = 1, size = 20) => request(`/enterprises?page=${page}&size=${size}`),
    get: (id) => request(`/enterprises/${id}`),
    create: (data) => request('/enterprises', { method: 'POST', body: JSON.stringify(data) }),
    update: (id, data) => request(`/enterprises/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    remove: (id) => request(`/enterprises/${id}`, { method: 'DELETE' }),
    getSub: (id, sub) => request(`/enterprises/${id}/sub/${sub}`),
  },
  hazardFactors: {
    list: (params = {}) => {
      const qs = new URLSearchParams(params).toString();
      return request(`/hazard-factors${qs ? '?' + qs : ''}`);
    },
    create: (data) => request('/hazard-factors', { method: 'POST', body: JSON.stringify(data) }),
    update: (id, data) => request(`/hazard-factors/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    remove: (id) => request(`/hazard-factors/${id}`, { method: 'DELETE' }),
    updateMonitor: (id, data) => request(`/hazard-factors/${id}/monitor`, { method: 'PUT', body: JSON.stringify(data) }),
  },
  workers: {
    list: (params = {}) => {
      const qs = new URLSearchParams(params).toString();
      return request(`/workers${qs ? '?' + qs : ''}`);
    },
    create: (data) => request('/workers', { method: 'POST', body: JSON.stringify(data) }),
    update: (id, data) => request(`/workers/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    remove: (id) => request(`/workers/${id}`, { method: 'DELETE' }),
    setExposedFactors: (id, factors) => request(`/workers/${id}/exposed-factors`, { method: 'PUT', body: JSON.stringify(factors) }),
  },
  examinations: {
    list: (params = {}) => {
      const qs = new URLSearchParams(params).toString();
      return request(`/examinations${qs ? '?' + qs : ''}`);
    },
    get: (id) => request(`/examinations/${id}`),
    schedule: (data) => request('/examinations', { method: 'POST', body: JSON.stringify(data) }),
    recordResult: (id, data) => request(`/examinations/${id}/results`, { method: 'PUT', body: JSON.stringify(data) }),
    updateFlow: (id, flowStatus) => request(`/examinations/${id}/flow`, { method: 'PUT', body: JSON.stringify({ flow_status: flowStatus }) }),
  },
  reports: {
    list: (params = {}) => {
      const qs = new URLSearchParams(params).toString();
      return request(`/reports${qs ? '?' + qs : ''}`);
    },
    generate: (data) => request('/reports', { method: 'POST', body: JSON.stringify(data) }),
    get: (id) => request(`/reports/${id}`),
    exportUrl: (id) => `${API_BASE}/reports/${id}/export`,
  },
  stats: {
    regulatory: (params = {}) => {
      const qs = new URLSearchParams(params).toString();
      return request(`/stats/regulatory${qs ? '?' + qs : ''}`);
    },
    metrics: () => request('/stats/metrics'),
  },
  todos: {
    list: (params = {}) => {
      const qs = new URLSearchParams(params).toString();
      return request(`/todos${qs ? '?' + qs : ''}`);
    },
    update: (id, data) => request(`/todos/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  },
};
