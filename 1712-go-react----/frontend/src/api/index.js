const API_BASE = 'http://localhost:8080/api';

const request = async (url, options = {}) => {
  const response = await fetch(`${API_BASE}${url}`, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  });
  
  if (!response.ok) {
    const error = await response.text();
    throw new Error(error || 'Request failed');
  }
  
  return response;
};

export const residents = {
  search: (name = '', idCard = '', community = '') => 
    request(`/residents?name=${encodeURIComponent(name)}&idCard=${encodeURIComponent(idCard)}&community=${encodeURIComponent(community)}`).then(r => r.json()),
  create: (data) => 
    request('/residents', { method: 'POST', body: JSON.stringify(data) }).then(r => r.json()),
  get: (id) => 
    request(`/residents/${id}`).then(r => r.json()),
};

export const checkups = {
  get: (residentID) => 
    request(`/checkups?residentID=${residentID}`).then(r => r.json()),
  create: (data) => 
    request('/checkups', { method: 'POST', body: JSON.stringify(data) }).then(r => r.json()),
};

export const followups = {
  get: (residentID = '', disease = '') => 
    request(`/followups?residentID=${residentID}&disease=${disease}`).then(r => r.json()),
  create: (data) => 
    request('/followups', { method: 'POST', body: JSON.stringify(data) }).then(r => r.json()),
};

export const chronicPatients = {
  list: () => request('/chronic-patients').then(r => r.json()),
};

export const families = {
  list: () => request('/families').then(r => r.json()),
  get: (id) => request(`/families/${id}`).then(r => r.json()),
  create: (data) => 
    request('/families', { method: 'POST', body: JSON.stringify(data) }).then(r => r.json()),
};

export const exports = {
  communityStats: (community) => 
    `${API_BASE}/export/community-stats?community=${encodeURIComponent(community)}`,
  chronicPatients: () => 
    `${API_BASE}/export/chronic-patients`,
  completeArchive: (residentID) => 
    `${API_BASE}/export/complete-archive?residentID=${residentID}`,
};
