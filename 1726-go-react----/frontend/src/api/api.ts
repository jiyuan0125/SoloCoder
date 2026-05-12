import axios from 'axios';
import type { Paper, User, PublicationInfo } from '../types';

const API_BASE = '/api/r';

const getHeaders = (userId: string) => ({
  headers: {
    'X-User-ID': userId,
    'Content-Type': 'application/json',
  },
});

export const api = {
  async getUsers(): Promise<User[]> {
    const res = await axios.get<User[]>(`${API_BASE}/users`);
    return res.data;
  },

  async createPaper(userId: string, data: {
    title: string;
    abstract: string;
    keywords: string;
    authors: any[];
    subject_category: string;
    target_journal: string;
  }): Promise<Paper> {
    const res = await axios.post<Paper>(`${API_BASE}`, data, getHeaders(userId));
    return res.data;
  },

  async getPapers(userId: string, params?: { status?: string; keyword?: string }): Promise<Paper[]> {
    const res = await axios.get<Paper[]>(`${API_BASE}`, {
      ...getHeaders(userId),
      params,
    });
    return res.data;
  },

  async getPaper(userId: string, id: string): Promise<Paper> {
    const res = await axios.get<Paper>(`${API_BASE}/${id}`, getHeaders(userId));
    return res.data;
  },

  async updatePaper(userId: string, id: string, data: {
    title?: string;
    abstract?: string;
    keywords?: string;
    authors?: any[];
    subject_category?: string;
    target_journal?: string;
  }): Promise<Paper> {
    const res = await axios.put<Paper>(`${API_BASE}/${id}`, data, getHeaders(userId));
    return res.data;
  },

  async submitForReview(userId: string, id: string): Promise<Paper> {
    const res = await axios.post<Paper>(`${API_BASE}/${id}/submit`, {}, getHeaders(userId));
    return res.data;
  },

  async assignReviewers(userId: string, id: string, reviewers: any[]): Promise<Paper> {
    const res = await axios.post<Paper>(`${API_BASE}/${id}/assign-reviewers`, { reviewers }, getHeaders(userId));
    return res.data;
  },

  async submitReview(userId: string, id: string, data: {
    reviewer_id: string;
    decision: string;
    score: number;
    comments: string;
  }): Promise<Paper> {
    const res = await axios.post<Paper>(`${API_BASE}/${id}/submit-review`, data, getHeaders(userId));
    return res.data;
  },

  async publishPaper(userId: string, id: string, info: PublicationInfo): Promise<Paper> {
    const res = await axios.post<Paper>(`${API_BASE}/${id}/publish`, { publication_info: info }, getHeaders(userId));
    return res.data;
  },

  async approveAction(userId: string, id: string): Promise<Paper> {
    const res = await axios.post<Paper>(`${API_BASE}/${id}/actions/approve`, {}, getHeaders(userId));
    return res.data;
  },

  async rejectAction(userId: string, id: string): Promise<Paper> {
    const res = await axios.post<Paper>(`${API_BASE}/${id}/actions/reject`, {}, getHeaders(userId));
    return res.data;
  },

  async cancelAction(userId: string, id: string): Promise<Paper> {
    const res = await axios.post<Paper>(`${API_BASE}/${id}/actions/cancel`, {}, getHeaders(userId));
    return res.data;
  },
};
