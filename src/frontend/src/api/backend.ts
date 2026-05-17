import axios from 'axios'

// Go 后端 API 封装
const backend = axios.create({
  baseURL: '/api/v1',
  timeout: 5000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 友好错误消息映射（不暴露技术细节）
const friendlyMessages: Record<number, string> = {
  400: '请求参数有误，请检查后重试',
  401: '登录已过期，请重新登录',
  403: '没有操作权限',
  404: '请求的资源不存在',
  429: '操作过于频繁，请稍后再试',
  500: '服务器开小差了，请稍后重试',
  502: '服务器维护中，请稍后重试',
  503: '服务暂不可用，请稍后重试',
}

function getFriendlyMessage(status?: number): string {
  return friendlyMessages[status ?? 0] || '网络请求失败，请稍后重试'
}

// 请求拦截器：自动附加 JWT
backend.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('zq_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error),
)

// 响应拦截器：统一处理 {code, message, data} + 自动重试
backend.interceptors.response.use(
  (response) => {
    const { code, message, data } = response.data
    if (code === 0) {
      return data
    }
    // Token 过期
    if (code === 40102 || code === 40103) {
      localStorage.removeItem('zq_token')
      window.location.href = '/login'
      return Promise.reject(new Error(message || 'Token 已过期'))
    }
    return Promise.reject(new Error(message || getFriendlyMessage(response.status)))
  },
  async (error) => {
    const config = error.config

    // 自动重试：仅对 5xx / 网络错误重试 1 次
    const shouldRetry =
      !config._retry &&
      (!error.response || error.response.status >= 500)

    if (shouldRetry) {
      config._retry = true
      try {
        return await backend(config)
      } catch {
        // 重试失败，走下方错误提示
      }
    }

    // HTTP 状态码错误处理
    const status = error.response?.status
    if (status === 401) {
      localStorage.removeItem('zq_token')
      window.location.href = '/login'
    }

    // 不在控制台打印敏感信息，返回友好提示
    const msg = getFriendlyMessage(status)
    return Promise.reject(new Error(msg))
  },
)

export default backend

// 注册
export async function register(username: string, password: string, email?: string) {
  return backend.post('/auth/register', { username, password, email })
}
