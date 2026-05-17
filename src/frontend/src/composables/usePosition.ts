import { ref, computed, onMounted, onUnmounted } from 'vue'
import backend from '@/api/backend'
import dataApi from '@/api/data'
import { safeGetWebSocket } from '@/utils/websocket'

export interface PositionItem {
  id: number
  stockCode: string
  stockName: string
  quantity: number
  availableQty: number
  avgCost: number
  currentPrice: number
  marketValue: number
  profitLoss: number
  profitRate: number
}

export function usePosition() {
  const positions = ref<PositionItem[]>([])
  const loading = ref(false)
  let refreshTimer: number | null = null

  // 计算盈亏
  function calcPnL(pos: PositionItem) {
    pos.marketValue = pos.currentPrice * pos.quantity
    pos.profitLoss = (pos.currentPrice - pos.avgCost) * pos.quantity
    pos.profitRate = pos.avgCost > 0
      ? pos.profitLoss / (pos.avgCost * pos.quantity)
      : 0
  }

  // 获取单只股票实时价格（带超时保护）
  async function fetchPrice(pos: PositionItem) {
    try {
      const res: any = await dataApi.get(`/market/${pos.stockCode}`)
      const price = res?.current_price ?? res?.data?.current_price
      if (price) pos.currentPrice = price
    } catch {
      // 获取失败保留当前价
    }
  }

  // 带超时的批量获取
  async function fetchPricesWithTimeout() {
    const fetchPromise = Promise.all(positions.value.map(fetchPrice))
    const timeoutPromise = new Promise(r => setTimeout(r, 5000))

    await Promise.race([fetchPromise, timeoutPromise])
    // 不管是否完成都继续
  }

  // 加载持仓列表
  async function fetchPositions() {
    loading.value = true
    try {
      const res: any = await backend.get('/position')
      const list: PositionItem[] = res ?? res?.data ?? []

      positions.value = list.map((item: any) => ({
        id: item.id || 0,
        stockCode: item.stock_code || item.stockCode || '',
        stockName: item.stock_name || item.stockName || '',
        quantity: item.quantity || 0,
        availableQty: item.available_qty || item.availableQty || item.quantity || 0,
        avgCost: item.avg_cost || item.avgCost || 0,
        currentPrice: item.current_price || item.currentPrice || item.avg_cost || item.avgCost || 0,
        marketValue: 0,
        profitLoss: 0,
        profitRate: 0,
      }))

      // 获取实时价格（不阻塞）
      await fetchPricesWithTimeout()

      // 计算盈亏
      positions.value.forEach(calcPnL)

      // 按盈亏比例降序
      positions.value.sort((a, b) => b.profitRate - a.profitRate)
    } catch (e) {
      console.error('[Position] fetchPositions failed', e)
    } finally {
      loading.value = false
    }
  }

  // 定时刷新价格（30 秒）
  async function refreshPrices() {
    await fetchPricesWithTimeout()
    positions.value.forEach(calcPnL)
    positions.value.sort((a, b) => b.profitRate - a.profitRate)
  }

  // 汇总数据
  const totalMarketValue = computed(() =>
    positions.value.reduce((sum, p) => sum + p.marketValue, 0),
  )

  const totalProfitLoss = computed(() =>
    positions.value.reduce((sum, p) => sum + p.profitLoss, 0),
  )

  const winRate = computed(() => {
    if (positions.value.length === 0) return '0.0'
    const wins = positions.value.filter((p) => p.profitRate > 0).length
    return ((wins / positions.value.length) * 100).toFixed(1)
  })

  // 保存回调引用以便清理
  const positionUpdateHandler = (data: any) => {
    if (Array.isArray(data)) {
      positions.value = data.map((item: any) => ({
        id: item.id || 0,
        stockCode: item.stock_code || item.stockCode || '',
        stockName: item.stock_name || item.stockName || '',
        quantity: item.quantity || 0,
        availableQty: item.available_qty || item.availableQty || item.quantity || 0,
        avgCost: item.avg_cost || item.avgCost || 0,
        currentPrice: item.current_price || item.currentPrice || 0,
        marketValue: 0,
        profitLoss: 0,
        profitRate: 0,
      }))
      positions.value.forEach(calcPnL)
    }
  }

  onMounted(() => {
    fetchPositions()
    safeGetWebSocket().on('position_update', positionUpdateHandler)
    // 定时刷新价格
    refreshTimer = window.setInterval(refreshPrices, 30000)
  })

  onUnmounted(() => {
    safeGetWebSocket().off('position_update', positionUpdateHandler)
    if (refreshTimer) clearInterval(refreshTimer)
  })

  return {
    positions,
    loading,
    fetchPositions,
    refreshPrices,
    totalMarketValue,
    totalProfitLoss,
    winRate,
  }
}
