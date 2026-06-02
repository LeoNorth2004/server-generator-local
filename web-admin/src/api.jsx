import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 60000,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
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
    } else if (error.response?.status === 403) {
      console.warn('Access denied: insufficient permissions');
    } else if (!error.response && error.message === 'Network Error') {
      console.error('Network error - backend may be unavailable');
    }
    return Promise.reject(error);
  }
);

export const authAPI = {
  login: (data) => api.post('/auth/login', data),
  register: (data) => api.post('/auth/register', data),
  getMe: () => api.get('/auth/me'),
};

export const projectAPI = {
  list: () => api.get('/projects'),
  get: (id) => api.get(`/projects/${id}`),
  create: (data) => api.post('/projects', data),
  update: (id, data) => api.put(`/projects/${id}`, data),
  delete: (id) => api.delete(`/projects/${id}`),
  regenerate: (id) => api.post(`/generator/generate/${id}`),
};

export const generatorAPI = {
  generate: (data) => api.post('/generator/generate', data),
  generateFromProject: (id) => api.post(`/generator/generate/${id}`),
  download: (projectId) => api.get(`/generator/download/${projectId}`, { responseType: 'blob' }),
  preview: (projectId) => api.get(`/generator/preview/${projectId}`),
};

export const userAPI = {
  list: () => api.get('/users'),
  get: (id) => api.get(`/users/${id}`),
  create: (data) => api.post('/users', data),
  update: (id, data) => api.put(`/users/${id}`, data),
  delete: (id) => api.delete(`/users/${id}`),
};

export const operationsAPI = {
  health: () => api.get('/operations/health'),
  stats: () => api.get('/operations/stats'),
  getMetrics: () => api.get('/operations/metrics'),
  getServices: () => api.get('/operations/services'),
  getEvents: (lang) => api.get(`/operations/events${lang ? '?lang=' + lang : ''}`),
  getOperationLogs: (params = {}) => api.get('/operations/operation-logs', { params }),
  recordOperationLog: (data) => api.post('/operations/operation-logs/record', data),
};

export const metadataAPI = {
  register: (data) => api.post('/generator/metadata/register', data),
  get: (modelName) => api.get(`/generator/metadata/${modelName}`),
  list: () => api.get('/generator/metadata'),
  getSchema: (modelName) => api.get(`/generator/metadata/${modelName}/schema`),
  getHandlers: (modelName) => api.get(`/generator/metadata/${modelName}/handlers`),
  export: (modelName) => api.get(`/generator/metadata/${modelName}/export`),
  import: (data) => api.post('/generator/metadata/import', data),
};

export const engineAPI = {
  getStats: () => api.post('/generator/engine/stats'),
  healthCheck: () => api.get('/health'),
  healthDetails: () => api.get('/health/details'),
  healthMetrics: () => api.get('/health/metrics'),
  healthHistory: () => api.get('/health/history'),
};

export default api;