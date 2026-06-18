/**
 * API 客户端封装
 * 1. 统一 axios 实例与拦截器
 * 2. 通过工厂函数生成各业务模块的 API 端点集合
 */
import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';
const TOKEN_KEY = 'token';
const USER_KEY = 'user';
const LOGIN_PATH = '/login';
const REQUEST_TIMEOUT = 60000;
const DEFAULT_HEADER = 'Content-Type';
const DEFAULT_CONTENT_TYPE = 'application/json';
const AUTH_HEADER = 'Authorization';
const AUTH_SCHEME = 'Bearer';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: { [DEFAULT_HEADER]: DEFAULT_CONTENT_TYPE },
  timeout: REQUEST_TIMEOUT,
});

/**
 * 请求拦截器：自动注入 token
 */
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY);
  if (token) {
    config.headers[AUTH_HEADER] = `${AUTH_SCHEME} ${token}`;
  }
  return config;
});

/**
 * 响应拦截器：处理 401/403/网络错误
 */
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status;
    if (status === 401) {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
      window.location.href = LOGIN_PATH;
    } else if (status === 403) {
      console.warn('Access denied: insufficient permissions');
    } else if (!error.response && error.message === 'Network Error') {
      console.error('Network error - backend may be unavailable');
    }
    return Promise.reject(error);
  }
);

/**
 * 构造 RESTful 资源 API
 * 提供 list/get/create/update/delete 通用方法
 */
const createResourceApi = (basePath) => ({
  list: (params) => apiClient.get(basePath, { params }),
  get: (id) => apiClient.get(`${basePath}/${id}`),
  create: (payload) => apiClient.post(basePath, payload),
  update: (id, payload) => apiClient.put(`${basePath}/${id}`, payload),
  delete: (id) => apiClient.delete(`${basePath}/${id}`),
});

const authAPI = {
  login: (data) => apiClient.post('/auth/login', data),
  register: (data) => apiClient.post('/auth/register', data),
  getMe: () => apiClient.get('/auth/me'),
};

const projectAPI = {
  ...createResourceApi('/projects'),
  regenerate: (id) => apiClient.post(`/generator/generate/${id}`),
};

const generatorAPI = {
  generate: (data) => apiClient.post('/generator/generate', data),
  generateFromProject: (id) => apiClient.post(`/generator/generate/${id}`),
  download: (projectId) =>
    apiClient.get(`/generator/download/${projectId}`, { responseType: 'blob' }),
  preview: (projectId) => apiClient.get(`/generator/preview/${projectId}`),
};

const userAPI = createResourceApi('/users');

const operationsAPI = {
  health: () => apiClient.get('/operations/health'),
  stats: () => apiClient.get('/operations/stats'),
  getMetrics: () => apiClient.get('/operations/metrics'),
  getServices: () => apiClient.get('/operations/services'),
  getEvents: (lang) =>
    apiClient.get(`/operations/events${lang ? `?lang=${lang}` : ''}`),
  getOperationLogs: (params = {}) =>
    apiClient.get('/operations/operation-logs', { params }),
  recordOperationLog: (data) =>
    apiClient.post('/operations/operation-logs/record', data),
};

const metadataAPI = {
  register: (data) => apiClient.post('/generator/metadata/register', data),
  get: (modelName) => apiClient.get(`/generator/metadata/${modelName}`),
  list: () => apiClient.get('/generator/metadata'),
  getSchema: (modelName) =>
    apiClient.get(`/generator/metadata/${modelName}/schema`),
  getHandlers: (modelName) =>
    apiClient.get(`/generator/metadata/${modelName}/handlers`),
  export: (modelName) =>
    apiClient.get(`/generator/metadata/${modelName}/export`),
  import: (data) => apiClient.post('/generator/metadata/import', data),
};

const engineAPI = {
  getStats: () => apiClient.post('/generator/engine/stats'),
  healthCheck: () => apiClient.get('/health'),
  healthDetails: () => apiClient.get('/health/details'),
  healthMetrics: () => apiClient.get('/health/metrics'),
  healthHistory: () => apiClient.get('/health/history'),
};

export {
  apiClient,
  authAPI,
  projectAPI,
  generatorAPI,
  userAPI,
  operationsAPI,
  metadataAPI,
  engineAPI,
};

export default apiClient;
