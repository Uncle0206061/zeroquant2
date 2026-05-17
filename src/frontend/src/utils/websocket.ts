/** WebSocket 全局封装
 * - 心跳间隔：10 秒
 * - 断开自动重连：最多 10 次，间隔 2 秒
 * - 全局唯一实例，跨页面复用
 */

import { ref } from 'vue'

type EventCallback = (data: any) => void

/** 全局 WebSocket 连接状态（供 UI 组件使用） */
export const wsStatus = ref<'connected' | 'connecting' | 'disconnected' | null>(null)

// WebSocket 开关 - 暂时禁用，等后端 WebSocket 服务就绪后再开启
const WS_ENABLED = false

class WebSocketClient {
  private ws: WebSocket | null = null
  private url: string
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private reconnectCount = 0
  private maxReconnect = 10
  private reconnectInterval = 2000
  private heartbeatInterval = 10000
  private listeners: Map<string, Set<EventCallback>> = new Map()
  private isManualClose = false

  constructor(url: string) {
    this.url = url
  }

  connect(): void {
    // 未登录时不连接
    const token = localStorage.getItem('zq_token')
    if (!token) {
      wsStatus.value = null
      return
    }

    if (this.ws?.readyState === WebSocket.OPEN) return
    if (this.ws?.readyState === WebSocket.CONNECTING) return

    wsStatus.value = 'connecting'
    this.isManualClose = false
    const wsUrl = token ? `${this.url}?token=${token}` : this.url

    try {
      this.ws = new WebSocket(wsUrl)
    } catch {
      console.error('[WS] 连接创建失败')
      wsStatus.value = 'disconnected'
      return
    }

    this.ws.onopen = () => {
      console.log('[WS] 已连接')
      wsStatus.value = 'connected'
      this.reconnectCount = 0
      this.startHeartbeat()
    }

    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)

        // 心跳响应
        if (msg.type === 'pong') return

        // 事件分发
        const eventType = msg.type || msg.event
        if (eventType && this.listeners.has(eventType)) {
          this.listeners.get(eventType)!.forEach((cb) => cb(msg.data ?? msg))
        }
      } catch {
        // 非 JSON 消息，忽略
      }
    }

    this.ws.onclose = (event) => {
      console.warn(`[WS] 已断开 (code: ${event.code})`)
      wsStatus.value = 'disconnected'
      this.stopHeartbeat()
      // 断开后不自动重连，等待用户手动刷新或页面重新加载
    }

    this.ws.onerror = (error) => {
      console.error('[WS] 错误', error)
      wsStatus.value = 'disconnected'
    }
  }

  /** 注册事件监听 */
  on(event: string, callback: EventCallback): void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }
    this.listeners.get(event)!.add(callback)
  }

  /** 取消事件监听 */
  off(event: string, callback: EventCallback): void {
    this.listeners.get(event)?.delete(callback)
  }

  /** 发送消息 */
  send(data: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(typeof data === 'string' ? data : JSON.stringify(data))
    } else {
      console.warn('[WS] 未连接，无法发送')
    }
  }

  /** 手动断开 */
  disconnect(): void {
    this.isManualClose = true
    this.stopHeartbeat()
    this.clearReconnect()
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  /** 获取连接状态 */
  get connected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  // --- 内部方法 ---

  private startHeartbeat(): void {
    this.stopHeartbeat()
    this.heartbeatTimer = setInterval(() => {
      this.send({ type: 'ping', timestamp: Date.now() })
    }, this.heartbeatInterval)
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  private tryReconnect(): void {
    if (this.reconnectCount >= this.maxReconnect) {
      console.error(`[WS] 重连失败，已达最大次数 (${this.maxReconnect})`)
      return
    }
    this.reconnectCount++
    const delay = this.reconnectInterval * this.reconnectCount
    console.log(`[WS] 第 ${this.reconnectCount} 次重连，${delay / 1000}s 后...`)
    this.reconnectTimer = setTimeout(() => {
      this.connect()
    }, delay)
  }

  private clearReconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    this.reconnectCount = 0
  }
}

// 全局唯一实例
let instance: WebSocketClient | null = null

// 安全获取 WebSocket（返回空对象防止报错）
export function safeGetWebSocket() {
  const ws = getWebSocket()
  if (!ws) {
    return {
      on: () => {},
      off: () => {},
      send: () => {},
    }
  }
  return ws
}

export function getWebSocket(url?: string): WebSocketClient {
  // WebSocket 已禁用
  if (!WS_ENABLED) {
    return null as any
  }

  // 未登录不创建实例
  if (!localStorage.getItem('zq_token')) {
    wsStatus.value = null
    return null as any
  }

  if (!instance) {
    instance = new WebSocketClient(url || 'ws://localhost:8080/api/v1/ws')
  }
  // 只有未连接时才尝试连接
  if (instance.ws?.readyState !== WebSocket.OPEN && instance.ws?.readyState !== WebSocket.CONNECTING) {
    instance.connect()
  }
  return instance
}

// 懒初始化，不在模块加载时自动连接
export default null

export { WebSocketClient }
