import axios from 'axios';
import type {
  Teacher,
  Training,
  Registration,
  TeachingEvaluation,
  TitleApplication,
  StatsResponse,
  AttendanceItem,
  Score,
  Title,
  ReviewStage,
  ReviewStatus,
} from './types';

const api = axios.create({
  baseURL: 'http://localhost:8080/api',
});

export const teacherApi = {
  getAll: () => api.get<Teacher[]>('/teachers'),
  create: (data: Partial<Teacher>) => api.post<Teacher>('/teachers', data),
  getById: (id: string) => api.get<Teacher>(`/teachers/detail/${id}`),
};

export const trainingApi = {
  getAll: () => api.get<Training[]>('/trainings'),
  create: (data: {
    name: string;
    type: Training['type'];
    form: Training['form'];
    date: string;
    hours: number;
    lecturer: string;
    capacity: number;
  }) => api.post<Training>('/trainings', data),
  register: (id: string, teacherId: string) =>
    api.post<Registration>(`/trainings/${id}/register`, { teacher_id: teacherId }),
  getRegistrations: (id: string) =>
    api.get<Registration[]>(`/trainings/${id}/registrations`),
};

export const registrationApi = {
  getAll: (teacherId?: string) =>
    api.get<Registration[]>('/registrations', {
      params: teacherId ? { teacher_id: teacherId } : {},
    }),
  update: (
    id: string,
    data: { attendance?: AttendanceItem[]; score?: number; study_report?: boolean }
  ) => api.put<Registration>(`/registrations/${id}`, data),
};

export const evaluationApi = {
  getAll: (teacherId?: string) =>
    api.get<TeachingEvaluation[]>('/evaluations', {
      params: teacherId ? { teacher_id: teacherId } : {},
    }),
  create: (data: {
    teacher_id: string;
    semester: string;
    attitude?: Score;
    content?: Score;
    method?: Score;
    effect?: Score;
  }) => api.post<TeachingEvaluation>('/evaluations', data),
};

export const applicationApi = {
  getAll: (teacherId?: string) =>
    api.get<TitleApplication[]>('/applications', {
      params: teacherId ? { teacher_id: teacherId } : {},
    }),
  create: (data: {
    teacher_id: string;
    apply_title: Title;
    materials: string;
    achievement_summary: string;
  }) => api.post<TitleApplication>('/applications', data),
  submitReview: (
    id: string,
    data: { stage: ReviewStage; status: ReviewStatus; comment?: string }
  ) => api.post<TitleApplication>(`/applications/${id}/review`, data),
};

export const statsApi = {
  getAll: () => api.get<StatsResponse>('/stats'),
};
