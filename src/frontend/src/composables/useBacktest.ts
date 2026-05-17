import { ref, reactive, nextTick, type Ref } from 'vue'
import echarts from '@/utils/echarts'
import backend from '@/api/backend'
import { safeGetWebSocket } from '@/utils/websocket'
import { ElMessage } from 'element-plus'

export interface BacktestForm {
  stockCode: string
  startDate: string
  endDate: string
  capital: number
}

export interface BacktestResult {
  metrics: {
    label: string
    value: string
    color?: string
  }[]
  equity_curve: { date: string; equity: number }[]
  drawdown_curve: { date: string; drawdown: number }[]
  return_distribution: number[]
  trades: {
    time: string
    stockCode: string
    stockName: string
    direction: 'buy' | 'sell'
    price: number
    quantity: number
    amount: number
  }[]
}

export function useBacktest() {
  const form = reactive<BacktestForm>({
    stockCode: '000001.SZ',
    startDate: '2024-01-01',
    endDate: '2024-12-31',
    capital: 1000000,
  })

  const running = ref(false)
  const progress = ref(0)
  const result = ref<BacktestResult | null>(null)

  // ECharts 实例
  let equityChart: echarts.ECharts | null = null
  let drawdownChart: echarts.ECharts | null = null
  let returnDistChart: echarts.ECharts | null = null

  /** 初始化 ECharts 实例（传入 DOM ref） */
  function initCharts(
    equityRef: Ref<HTMLElement | null>,
    drawdownRef: Ref<HTMLElement | null>,
    returnDistRef: Ref<HTMLElement | null>,
  ) {
    if (equityRef.value && !equityChart) equityChart = echarts.init(equityRef.value)
    if (drawdownRef.value && !drawdownChart) drawdownChart = echarts.init(drawdownRef.value)
    if (returnDistRef.value && !returnDistChart) returnDistChart = echarts.init(returnDistRef.value)
  }

  /** 绘制资金曲线 */
  function drawEquityCurve(data: { date: string; equity: number }[]) {
    if (!equityChart) return
    equityChart.setOption({
      tooltip: { trigger: 'axis', formatter: (p: any) => `${p[0].axisValue}<br/>资金：¥${(p[0].value / 10000).toFixed(2)}万` },
      grid: { left: 70, right: 20, top: 20, bottom: 40 },
      xAxis: { type: 'category', data: data.map((d) => d.date), axisLabel: { fontSize: 11 } },
      yAxis: { type: 'value', axisLabel: { formatter: (v: number) => `${(v / 10000).toFixed(0)}万` } },
      series: [{
        name: '资金',
        type: 'line',
        data: data.map((d) => d.equity),
        smooth: true,
        lineStyle: { width: 2 },
        areaStyle: { opacity: 0.15 },
      }],
    })
  }

  /** 绘制回撤曲线 */
  function drawDrawdown(data: { date: string; drawdown: number }[]) {
    if (!drawdownChart) return
    drawdownChart.setOption({
      tooltip: { trigger: 'axis', formatter: (p: any) => `${p[0].axisValue}<br/>回撤：${(p[0].value * 100).toFixed(2)}%` },
      grid: { left: 70, right: 20, top: 20, bottom: 40 },
      xAxis: { type: 'category', data: data.map((d) => d.date), axisLabel: { fontSize: 11 } },
      yAxis: { type: 'value', axisLabel: { formatter: (v: number) => `${(v * 100).toFixed(1)}%` } },
      series: [{
        name: '回撤',
        type: 'line',
        data: data.map((d) => d.drawdown),
        smooth: true,
        lineStyle: { width: 2, color: '#F56C6C' },
        areaStyle: { color: 'rgba(245,108,108,0.15)' },
      }],
    })
  }

  /** 绘制收益分布柱状图 */
  function drawReturnDist(returns: number[]) {
    if (!returnDistChart) return
    const barData = returns.map((r) => ({
      value: r,
      itemStyle: { color: r >= 0 ? '#F56C6C' : '#67C23A' },
    }))
    returnDistChart.setOption({
      tooltip: { trigger: 'axis', formatter: (p: any) => `收益：${(p[0].value * 100).toFixed(2)}%` },
      grid: { left: 70, right: 20, top: 20, bottom: 40 },
      xAxis: { type: 'category', data: returns.map((_, i) => `${i}`) },
      yAxis: { type: 'value', axisLabel: { formatter: (v: number) => `${(v * 100).toFixed(1)}%` } },
      series: [{ name: '收益', type: 'bar', data: barData }],
    })
  }

  /** 发起回测 */
  async function runBacktest() {
    if (!form.stockCode.trim()) {
      ElMessage.warning('请输入股票代码')
      return
    }

    running.value = true
    progress.value = 0
    result.value = null

    try {
      // 1. 提交回测任务
      const res: any = await backend.post('/strategy/backtest/submit', {
        stock_code: form.stockCode,
        start_date: form.startDate,
        end_date: form.endDate,
        initial_capital: form.capital,
      })
      const taskId = res?.task_id || res?.data?.task_id

      // 2. 监听 WebSocket 进度
      safeGetWebSocket().on('backtest_progress', (data: any) => {
        if (data?.task_id === taskId) {
          progress.value = data.progress || 0
        }
      })

      // 3. 监听回测完成
      safeGetWebSocket().on('backtest_result', (data: any) => {
        if (data?.task_id === taskId) {
          handleResult(data)
        }
      })

      // 4. 备用：轮询结果（WS 可能丢失）
      if (taskId) {
        const pollResult = await pollBacktestResult(taskId)
        if (pollResult) handleResult(pollResult)
      }
    } catch (e: any) {
      ElMessage.error(e.message || '回测失败')
    } finally {
      running.value = false
    }
  }

  /** 轮询回测结果 */
  async function pollBacktestResult(taskId: string): Promise<any> {
    const maxAttempts = 60
    const interval = 2000
    for (let i = 0; i < maxAttempts; i++) {
      await new Promise((r) => setTimeout(r, interval))
      try {
        const res: any = await backend.get(`/strategy/backtest/result/${taskId}`)
        if (res?.status === 'completed') return res
      } catch { /* 继续轮询 */ }
    }
    return null
  }

  /** 处理回测结果 */
  async function handleResult(data: any) {
    if (result.value) return // 避免重复处理

    result.value = {
      metrics: [
        { label: '年化收益率', value: `${(data.annual_return * 100).toFixed(2)}%`, color: data.annual_return >= 0 ? 'positive' : 'negative' },
        { label: '最大回撤', value: `${(data.max_drawdown * 100).toFixed(2)}%`, color: 'negative' },
        { label: '夏普比率', value: data.sharpe_ratio?.toFixed(2) || '--' },
        { label: '胜率', value: `${(data.win_rate * 100).toFixed(1)}%` },
        { label: '盈亏比', value: data.profit_loss_ratio?.toFixed(2) || '--' },
        { label: '总交易次数', value: data.total_trades?.toString() || '0' },
      ],
      equity_curve: data.equity_curve || [],
      drawdown_curve: data.drawdown_curve || [],
      return_distribution: data.return_distribution || [],
      trades: data.trades || [],
    }
  }

  /** 渲染图表（需要 DOM ref） */
  async function renderCharts(
    equityRef: Ref<HTMLElement | null>,
    drawdownRef: Ref<HTMLElement | null>,
    returnDistRef: Ref<HTMLElement | null>,
  ) {
    await nextTick()
    initCharts(equityRef, drawdownRef, returnDistRef)
    if (result.value?.equity_curve?.length) drawEquityCurve(result.value.equity_curve)
    if (result.value?.drawdown_curve?.length) drawDrawdown(result.value.drawdown_curve)
    if (result.value?.return_distribution?.length) drawReturnDist(result.value.return_distribution)
  }

  /** 销毁图表 */
  function disposeCharts() {
    equityChart?.dispose()
    drawdownChart?.dispose()
    returnDistChart?.dispose()
    equityChart = null
    drawdownChart = null
    returnDistChart = null
  }

  return {
    form,
    running,
    progress,
    result,
    runBacktest,
    renderCharts,
    disposeCharts,
  }
}
