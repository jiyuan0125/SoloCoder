import axios from 'axios';
import { Occupation, ExamBatch, Certificate, PassRateStats, ExamCandidate } from '../types';

const API_BASE_URL = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const occupationApi = {
  getAll: () => api.get<Occupation[]>('/occupations'),
  getById: (id: string) => api.get<Occupation>(`/occupations/${id}`),
  create: (data: Partial<Occupation>) => api.post<Occupation>('/occupations', data),
  update: (id: string, data: Partial<Occupation>) => api.put<Occupation>(`/occupations/${id}`, data),
  delete: (id: string) => api.delete(`/occupations/${id}`),
};

export const batchApi = {
  getAll: () => api.get<ExamBatch[]>('/batches'),
  getById: (id: string) => api.get<ExamBatch>(`/batches/${id}`),
  create: (data: any) => api.post<ExamBatch>('/batches', data),
  delete: (id: string) => api.delete(`/batches/${id}`),
  nextStatus: (id: string) => api.post<ExamBatch>(`/batches/${id}/status`, { action: 'next' }),
  registerCandidate: (id: string, data: any) => api.post<ExamCandidate>(`/batches/${id}/register`, data),
  enterScores: (id: string, data: any) => api.post<ExamCandidate>(`/batches/${id}/scores`, data),
};

export const certificateApi = {
  getAll: () => api.get<Certificate[]>('/certificates'),
  getById: (id: string) => api.get<Certificate>(`/certificates/${id}`),
  issue: (data: any) => api.post<Certificate>('/certificates', data),
  updateStatus: (id: string, status: string) => api.put<Certificate>(`/certificates/${id}/status`, { status }),
  getExpiringSoon: () => api.get<Certificate[]>('/certificates/expiring-soon'),
};

export const statsApi = {
  getPassRate: () => api.get<PassRateStats[]>('/stats/pass-rate'),
};

export default api;
