import axios from 'axios'

// Python 数据服务 API 封装
const dataService = axios.create({
  baseURL: '/data/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 友好错误消息映射
const friendlyMessages: Record<number, string> = {
  400: '请求参数有误',
  401: '认证失败',
  404: '数据不存在',
  500: '服务器错误',
  502: '服务维护中',
  503: '服务暂不可用',
}

function getFriendlyMessage(status?: number): string {
  return friendlyMessages[status ?? 0] || '请求失败，请稍后重试'
}

// 请求拦截器：自动附加 JWT
dataService.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('zq_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error),
)

// 响应拦截器：统一错误处理
dataService.interceptors.response.use(
  (response) => {
    const { code, message, data } = response.data
    if (code === 0) {
      return data
    }
    return Promise.reject(new Error(message || getFriendlyMessage(response.status)))
  },
  async (error) => {
    const config = error.config
    const status = error.response?.status

    // 自动重试 5xx 错误
    const shouldRetry = !config._retry && (!error.response || status >= 500)
    if (shouldRetry) {
      config._retry = true
      try {
        return await dataService(config)
      } catch {
        // 继续处理错误
      }
    }

    const msg = getFriendlyMessage(status)
    return Promise.reject(new Error(msg))
  },
)

export default dataService