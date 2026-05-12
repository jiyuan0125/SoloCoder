import axios from 'axios';
import { 
  LoginRequest, 
  LoginResponse, 
  User, 
  Diploma, 
  CreateDiplomaRequest, 
  UpdateDiplomaRequest, 
  VerifyRequest, 
  VerifyResponse, 
  CreateUserRequest, 
  UpdateUserRequest, 
  OperationLog, 
  PagedResponse 
} from '../types';

const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080/api';

const getToken = (): string | null => {
  return localStorage.getItem('token');
};

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export const authAPI = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const response = await axios.post(`${API_BASE_URL}/login`, data);
    return response.data;
  },
  
  getCurrentUser: async (): Promise<User> => {
    const response = await api.get('/me');
    return response.data;
  },
};

export const diplomaAPI = {
  list: async (): Promise<Diploma[]> => {
    const response = await api.get('/diplomas');
    return response.data;
  },
  
  get: async (id: string): Promise<Diploma> => {
    const response = await api.get(`/diplomas/${id}`);
    return response.data;
  },
  
  create: async (data: CreateDiplomaRequest): Promise<Diploma> => {
    const response = await api.post('/diplomas', data);
    return response.data;
  },
  
  update: async (id: string, data: UpdateDiplomaRequest): Promise<Diploma> => {
    const response = await api.put(`/diplomas/${id}`, data);
    return response.data;
  },
  
  delete: async (id: string): Promise<void> => {
    await api.delete(`/diplomas/${id}`);
  },
};

export const verifyAPI = {
  verify: async (data: VerifyRequest): Promise<VerifyResponse> => {
    const response = await api.post('/verify', data);
    return response.data;
  },
};

export const userAPI = {
  list: async (): Promise<User[]> => {
    const response = await api.get('/users');
    return response.data;
  },
  
  create: async (data: CreateUserRequest): Promise<User> => {
    const response = await api.post('/users', data);
    return response.data;
  },
  
  update: async (id: string, data: UpdateUserRequest): Promise<User> => {
    const response = await api.put(`/users/${id}`, data);
    return response.data;
  },
  
  delete: async (id: string): Promise<void> => {
    await api.delete(`/users/${id}`);
  },
};

export const logAPI = {
  list: async (params: {
    page?: number;
    page_size?: number;
    operation_type?: string;
    start_date?: string;
    end_date?: string;
  }): Promise<PagedResponse<OperationLog>> => {
    const response = await api.get('/logs', { params });
    return response.data;
  },
};
