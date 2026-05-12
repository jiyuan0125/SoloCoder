const API_BASE = 'http://localhost:8081/api';

async function request(endpoint, options = {}) {
  const response = await fetch(API_BASE + endpoint, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  });

  if (response.status === 204 || response.status === 200 && response.headers.get('content-length') === '0') {
    return null;
  }

  const text = await response.text();
  if (!text) {
    return null;
  }

  const data = JSON.parse(text);

  if (!response.ok) {
    throw new Error(data.error || '请求失败');
  }

  return data;
}

export const api = {
  getDoctors: () => request('/doctors'),
  getPatients: () => request('/patients'),
  getTodayPatients: () => request('/today-patients'),
  
  getAppointments: () => request('/appointments'),
  bookAppointment: (data) => request('/appointments', {
    method: 'POST',
    body: JSON.stringify(data),
  }),
  
  getTreatmentPlans: () => request('/treatment-plans'),
  createTreatmentPlan: (data) => request('/treatment-plans', {
    method: 'POST',
    body: JSON.stringify(data),
  }),
  updateStepStatus: (planId, stepId, status) => request(`/treatment-plans/${planId}/steps`, {
    method: 'PUT',
    body: JSON.stringify({ step_id: stepId, status }),
  }),
  
  getFollowUps: () => request('/follow-ups'),
  completeFollowUp: (id, method, feedback, notes) => request(`/follow-ups/${id}/complete`, {
    method: 'POST',
    body: JSON.stringify({ method, feedback, notes }),
  }),
  batchCompleteFollowUps: (ids, method, feedback, notes) => request('/follow-ups/batch-complete', {
    method: 'POST',
    body: JSON.stringify({ ids, method, feedback, notes }),
  }),
  
  getTodos: () => request('/todos'),
};
