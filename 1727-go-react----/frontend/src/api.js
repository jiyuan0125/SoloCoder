const API_BASE = 'http://localhost:8090/api';

export const fenToYuan = (fen) => {
  if (fen === null || fen === undefined) return '0.00';
  const sign = fen < 0 ? '-' : '';
  const abs = Math.abs(fen);
  const yuan = Math.floor(abs / 100);
  const cents = abs % 100;
  return `${sign}${yuan}.${cents.toString().padStart(2, '0')}`;
};

export const yuanToFen = (yuan) => {
  if (!yuan) return 0;
  const str = String(yuan).trim();
  if (str.includes('.')) {
    const parts = str.split('.');
    if (parts.length > 2) return null;
    const intPart = parts[0] || '0';
    let decPart = parts[1];
    if (decPart.length > 2) return null;
    while (decPart.length < 2) decPart += '0';
    const intVal = parseInt(intPart, 10);
    const decVal = parseInt(decPart, 10);
    if (isNaN(intVal) || isNaN(decVal)) return null;
    return intVal * 100 + decVal;
  }
  const val = parseInt(str, 10);
  return isNaN(val) ? null : val * 100;
};

const request = async (url, options = {}) => {
  const res = await fetch(`${API_BASE}${url}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || '请求失败');
  }
  return data;
};

export const projectAPI = {
  create: (data) => request('/projects', { method: 'POST', body: JSON.stringify(data) }),
  list: () => request('/projects'),
  get: (id) => request(`/projects/${id}`),
  close: (id) => request(`/projects/${id}/close`, { method: 'POST' }),
  finalReport: (id) => request(`/projects/${id}/final-report`),
};

export const reimbursementAPI = {
  create: (data) => request('/reimbursements', { method: 'POST', body: JSON.stringify(data) }),
  list: (projectId) => request(`/reimbursements${projectId ? `?project_id=${projectId}` : ''}`),
  pending: () => request('/reimbursements/pending'),
  approve: (id) => request(`/reimbursements/${id}/approve`, { method: 'POST' }),
  reject: (id, reason) => request(`/reimbursements/${id}/reject`, { method: 'POST', body: JSON.stringify({ reason }) }),
  resubmit: (id, data) => request(`/reimbursements/${id}/resubmit`, { method: 'POST', body: JSON.stringify(data) }),
};

export const CATEGORIES = [
  { key: 'equipment', name: '设备费' },
  { key: 'material', name: '材料费' },
  { key: 'travel', name: '差旅费' },
  { key: 'labor', name: '劳务费' },
  { key: 'expert', name: '专家咨询费' },
  { key: 'other', name: '其他费用' },
];

export const CATEGORY_COLORS = {
  equipment: '#FF6B6B',
  material: '#4ECDC4',
  travel: '#45B7D1',
  labor: '#96CEB4',
  expert: '#FFEAA7',
  other: '#DDA0DD',
};
