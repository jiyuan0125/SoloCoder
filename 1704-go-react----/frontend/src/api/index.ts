import axios from 'axios';
import type {
  Patient, Plan, Task, TrainingRecord, Assessment, PatientSummary
} from '../types';

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

export const patientApi = {
  list: () => api.get<Patient[]>('/patients'),
  get: (id: number) => api.get<Patient>(`/patients/${id}`),
  create: (data: Partial<Patient>) => api.post<Patient>('/patients', data),
  update: (id: number, data: Partial<Patient>) => api.put<Patient>(`/patients/${id}`, data),
  delete: (id: number) => api.delete(`/patients/${id}`),
};

export const planApi = {
  list: (patientId?: number) => api.get<Plan[]>('/plans', { params: { patientId } }),
  get: (id: number) => api.get<Plan>(`/plans/${id}`),
  create: (data: Partial<Plan>) => api.post<Plan>('/plans', data),
  getTasks: (id: number) => api.get<Task[]>(`/plans/${id}/tasks`),
};

export const trainingApi = {
  getPatientTasks: (patientId: number, date?: string) => 
    api.get<Task[]>(`/training/patients/${patientId}/tasks`, { params: { date } }),
  getTask: (id: number) => api.get<Task>(`/training/tasks/${id}`),
  createRecord: (data: Partial<TrainingRecord>) => 
    api.post<TrainingRecord>('/training/records', data),
};

export const assessmentApi = {
  list: (patientId?: number, planId?: number) => 
    api.get<Assessment[]>('/assessments', { params: { patientId, planId } }),
  get: (id: number) => api.get<Assessment>(`/assessments/${id}`),
  record: (id: number, data: Partial<Assessment>) => 
    api.post(`/assessments/${id}/record`, data),
};

export const summaryApi = {
  getPatientSummary: (patientId: number, startDate?: string, endDate?: string) =>
    api.get<PatientSummary>(`/summary/patients/${patientId}`, { params: { startDate, endDate } }),
};
