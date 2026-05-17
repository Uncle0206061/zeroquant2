import { ref, computed, onMounted, onUnmounted } from 'vue'
import backend from '@/api/backend'
import { safeGetWebSocket } from '@/utils/websocket'

export interface OrderItem {
  order_id: string
  created_at: string
  stock_code: string
  stock_name: string
  direction: number // 1=buy, 2=sell
  price: number
  quantity: number
  filled_qty: number
  status: string // pending/filled/cancelled/rejected
  commission?: number
  avg_price?: number
  reject_reason?: string
}

/** 订单状态配置 */
export const statusConfig: Record<string, { label: string; type: string }> = {
  pending: { label: '待成交', type: 'warning' },
  filled: { label: '已成交', type: 'success' },
  partial_filled: { label: '部分成交', type: '' },
  cancelled: { label: '已撤单', type: 'info' },
  rejected: { label: '已拒绝', type: 'danger' },
}

export function useOrders() {
  const allOrders = ref<OrderItem[]>([])
  const loading = ref(false)

  // 筛选
  const filterStatus = ref('')
  const filterDirection = ref('')
  const currentPage = ref(1)
  const pageSize = ref(20)

  // 筛选后列表
  const filteredOrders = computed(() => {
    let result = allOrders.value
    if (filterStatus.value) {
      result = result.filter((o) => o.status === filterStatus.value)
    }
    if (filterDirection.value) {
      result = result.filter((o) => o.direction.toString() === filterDirection.value)
    }
    return result
  })

  // 分页
  const paginatedOrders = computed(() => {
    const start = (currentPage.value - 1) * pageSize.value
    return filteredOrders.value.slice(start, start + pageSize.value)
  })

  const total = computed(() => filteredOrders.value.length)

  // 加载订单列表
  async function fetchOrders() {
    loading.value = true
    try {
      const res: any = await backend.get('/order/list', {
        params: { page: 1, page_size: 200 },
      })
      const list = res ?? res?.data ?? []
      allOrders.value = (Array.isArray(list) ? list : []).map(normalizeOrder)
    } catch (e) {
      console.error('[Orders] fetchOrders failed', e)
    } finally {
      loading.value = false
    }
  }

  // 字段名规范化（snake_case → camelCase 兼容）
  function normalizeOrder(item: any): OrderItem {
    return {
      order_id: item.order_id ?? '',
      created_at: item.created_at ?? item.createdAt ?? '',
      stock_code: item.stock_code ?? item.stockCode ?? '',
      stock_name: item.stock_name ?? item.stockName ?? '',
      direction: item.direction ?? 0,
      price: item.price ?? 0,
      quantity: item.quantity ?? 0,
      filled_qty: item.filled_qty ?? item.filledQty ?? 0,
      status: item.status ?? 'pending',
      commission: item.commission ?? undefined,
      avg_price: item.avg_price ?? item.avgPrice ?? undefined,
      reject_reason: item.reject_reason ?? item.rejectReason ?? undefined,
    }
  }

  // 撤单
  async function cancelOrder(orderId: string) {
    await backend.post(`/order/cancel/${orderId}`)
    const order = allOrders.value.find((o) => o.order_id === orderId)
    if (order) order.status = 'cancelled'
  }

  // 获取订单详情
  async function getOrderDetail(orderId: string): Promise<OrderItem> {
    const res: any = await backend.get(`/order/${orderId}`)
    return normalizeOrder(res ?? res?.data ?? {})
  }

  // 重置筛选
  function resetFilter() {
    filterStatus.value = ''
    filterDirection.value = ''
    currentPage.value = 1
  }

  // 筛选变化时回到第一页
  function onFilterChange() {
    currentPage.value = 1
  }

  // 保存回调引用以便清理
  const orderUpdateHandler = (data: any) => {
    const updated = normalizeOrder(data)
    const idx = allOrders.value.findIndex((o) => o.order_id === updated.order_id)
    if (idx >= 0) {
      allOrders.value[idx] = { ...allOrders.value[idx], ...updated }
    } else {
      allOrders.value.unshift(updated)
    }
  }

  onMounted(() => {
    fetchOrders()
    safeGetWebSocket().on('order_update', orderUpdateHandler)
  })

  onUnmounted(() => {
    safeGetWebSocket().off('order_update', orderUpdateHandler)
  })

  return {
    allOrders,
    filteredOrders,
    paginatedOrders,
    total,
    loading,
    filterStatus,
    filterDirection,
    currentPage,
    pageSize,
    fetchOrders,
    cancelOrder,
    getOrderDetail,
    resetFilter,
    onFilterChange,
    statusConfig,
  }
}
