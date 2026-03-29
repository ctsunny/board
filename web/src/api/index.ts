import axios, { type AxiosResponse } from 'axios'
import { getBoardBasePath } from '@/utils/runtime'

const baseURL = `${getBoardBasePath()}/api/v1`

export const http = axios.create({
  baseURL,
  timeout: 15000,
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('board_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('board_token')
      // avoid circular import by using window.location
      window.location.href = '#/login'
    }
    return Promise.reject(error)
  }
)

// ── Auth ──────────────────────────────────────────────────────────────────────
export const authApi = {
  login: (data: { username: string; password: string }) =>
    http.post<{ token: string }>('/auth/login', data),
}

// ── Dashboard ─────────────────────────────────────────────────────────────────
export const dashboardApi = {
  getDashboard: () => http.get('/dashboard'),
}

// ── Customers ─────────────────────────────────────────────────────────────────
export const customersApi = {
  list: (params?: Record<string, unknown>) => http.get('/customers', { params }),
  create: (data: Record<string, unknown>) => http.post('/customers', data),
  update: (id: number, data: Record<string, unknown>) => http.put(`/customers/${id}`, data),
  delete: (id: number) => http.delete(`/customers/${id}`),
  batchDelete: (ids: number[]) => http.post('/customers/batch-delete', { ids }),
  batchRenew: (ids: number[], days: number) => http.post('/customers/batch-renew', { ids, days }),
  exportCSV: (params?: Record<string, unknown>) =>
    http.get('/customers/export', { params, responseType: 'blob' }),
}

// ── Regions ───────────────────────────────────────────────────────────────────
export const regionsApi = {
  list: (params?: Record<string, unknown>) => http.get('/regions', { params }),
  create: (data: Record<string, unknown>) => http.post('/regions', data),
  update: (id: number, data: Record<string, unknown>) => http.put(`/regions/${id}`, data),
  delete: (id: number) => http.delete(`/regions/${id}`),
}

// ── Servers ───────────────────────────────────────────────────────────────────
export const serversApi = {
  list: (params?: Record<string, unknown>) => http.get('/servers', { params }),
  create: (data: Record<string, unknown>) => http.post('/servers', data),
  update: (id: number, data: Record<string, unknown>) => http.put(`/servers/${id}`, data),
  delete: (id: number) => http.delete(`/servers/${id}`),
  ping: (id: number) => http.post(`/servers/${id}/ping`),
}

// ── Routes ────────────────────────────────────────────────────────────────────
export const routesApi = {
  list: (params?: Record<string, unknown>) => http.get('/routes', { params }),
  create: (data: Record<string, unknown>) => http.post('/routes', data),
  update: (id: number, data: Record<string, unknown>) => http.put(`/routes/${id}`, data),
  delete: (id: number) => http.delete(`/routes/${id}`),
}

// ── Nodes ─────────────────────────────────────────────────────────────────────
export const nodesApi = {
  list: (params?: Record<string, unknown>) => http.get('/nodes', { params }),
  create: (data: Record<string, unknown>) => http.post('/nodes', data),
  update: (id: number, data: Record<string, unknown>) => http.put(`/nodes/${id}`, data),
  delete: (id: number) => http.delete(`/nodes/${id}`),
}

// ── Audit Logs ────────────────────────────────────────────────────────────────
export const auditLogsApi = {
  list: (params?: Record<string, unknown>) => http.get('/audit-logs', { params }),
}

// ── API Tokens ────────────────────────────────────────────────────────────────
export const tokensApi = {
  list: () => http.get('/tokens'),
  create: (data: { name: string }) => http.post('/tokens', data),
  delete: (id: number) => http.delete(`/tokens/${id}`),
}

// ── Settings ──────────────────────────────────────────────────────────────────
export const settingsApi = {
  get: () => http.get('/settings'),
  update: (data: Record<string, unknown>) => http.put('/settings', data),
  getGmailAuthUrl: () => http.get('/settings/gmail/auth-url'),
  submitGmailCallback: (code: string) => http.post('/settings/gmail/callback', { code }),
}

// ── System ────────────────────────────────────────────────────────────────────
export const systemApi = {
  getVersion: () => http.get('/system/version'),
  triggerUpdate: () => http.post('/system/update'),
}
