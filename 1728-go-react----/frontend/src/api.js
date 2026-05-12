import axios from 'axios';

const API_BASE = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE,
  headers: {
    'X-User-ID': 'user-001',
    'X-User-Name': '测试用户',
    'Content-Type': 'application/json',
  },
});

export default api;
