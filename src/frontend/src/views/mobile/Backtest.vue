<template>
  <div class="mobile-backtest">
    <!-- 回测发起表单（简化版） -->
    <div class="backtest-form">
      <van-cell-group inset>
        <van-field v-model="form.stockCode" label="股票代码" placeholder="如 000001.SZ" />
        <van-field v-model="form.startDate" label="开始日期" placeholder="2024-01-01" />
        <van-field v-model="form.endDate" label="结束日期" placeholder="2024-12-31" />
        <van-field v-model="form.capitalStr" label="初始资金" type="number" placeholder="1000000" />
      </van-cell-group>
      <div class="form-actions">
        <van-button
          type="primary"
          block
          round
          :loading="running"
          :disabled="!isOnline"
          @click="runBacktest"
        >
          {{ running ? `回测中 ${progress}%` : (isOnline ? '发起回测' : '离线不可用') }}
        </van-button>
      </div>
      <van-progress v-if="running" :percentage="progress" :show-pivot="false" color="#1989fa" />
    </div>

    <!-- 回测结果 -->
    <template v-if="result">
      <!-- 指标卡片 -->
      <div class="metrics-grid">
        <div class="metric-card" v-for="m in result.metrics" :key="m.label">
          <div class="metric-label">{{ m.label }}</div>
          <div class="metric-value" :class="m.color">{{ m.value }}</div>
        </div>
      </div>

      <!-- 资金曲线 -->
      <div class="chart-section">
        <div class="section-title">资金曲线</div>
        <div ref="equityChartRef" style="height: 220px; width: 100%" />
      </div>

      <!-- 回撤曲线 -->
      <div class="chart-section">
        <div class="section-title">回撤曲线</div>
        <div ref="drawdownChartRef" style="height: 200px; width: 100%" />
      </div>

      <!-- 收益分布 -->
      <div class="chart-section">
        <div class="section-title">收益分布</div>
        <div ref="returnDistChartRef" style="height: 200px; width: 100%" />
      </div>

      <!-- 成交记录 -->
      <div class="trades-section">
        <div class="section-title">成交记录（共 {{ result.trades.length }} 条）</div>
        <van-cell-group inset>
          <van-cell
            v-for="(trade, idx) in result.trades.slice(0, 20)"
            :key="idx"
          >
            <template #title>
              <div class="trade-row">
                <span class="trade-time">{{ trade.time?.substring(0, 10) }}</span>
                <span class="trade-stock">{{ trade.stockCode }}</span>
                <van-tag
                  :type="trade.direction === 'buy' ? 'danger' : 'success'"
                  size="mini"
                >
                  {{ trade.direction === 'buy' ? '买' : '卖' }}
                </van-tag>
                <span class="trade-price">¥{{ trade.price?.toFixed(2) }}</span>
                <span class="trade-qty">×{{ trade.quantity }}</span>
              </div>
            </template>
          </van-cell>
        </van-cell-group>
        <div v-if="result.trades.length > 20" class="trades-more">
          还有 {{ result.trades.length - 20 }} 条记录...
        </div>
      </div>
    </template>

    <!-- 空状态 -->
    <van-empty
      v-if="!result && !running"
      description="发起回测后，结果将在此展示"
      :image-size="100"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import backend from '@/api/backend'
import { safeGetWebSocket } from '@/utils/websocket'
import { useNetworkStatus } from '@/composables/useNetworkStatus'
import { useBacktestStore } from '@/stores/backtest'
import { showToast } from 'vant'
import echarts from '@/utils/echarts'

const { isOnline } = useNetworkStatus()
const backtestStore = useBacktestStore()

const form = reactive({
  stockCode: '000001.SZ',
  startDate: '2024-01-01',
  endDate: '2024-12-31',
  capital: 1000000,
  capitalStr: '1000000',
})

const running = ref(false)
const progress = ref(0)
const result = ref<any>(null)

// 从 Pinia 恢复上次结果
if (backtestStore.lastResult) {
  result.value = backtestStore.lastResult
}

// 图表 refs
const equityChartRef = ref<HTMLElement | null>(null)
const drawdownChartRef = ref<HTMLElement | null>(null)
const returnDistChartRef = ref<HTMLElement | null>(null)

let eqChart: any = null
let ddChart: any = null
let rdChart: any = null

// 监听 result 渲染图表
watch(result, async (val) => {
  if (val) {
    await nextTick()
    backtestStore.setResult(val)
    initCharts()
  }
}, { once: true })

onMounted(() => {
  if (result.value) {
    nextTick(() => initCharts())
  }
  window.addEventListener('resize', onResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  eqChart?.dispose()
  ddChart?.dispose()
  rdChart?.dispose()
})

function initCharts() {
  if (equityChartRef.value && !eqChart) {
    eqChart = echarts.init(equityChartRef.value)
    if (result.value?.equity_curve?.length) {
      eqChart.setOption({
        grid: { left: 50, right: 10, top: 10, bottom: 30 },
        xAxis: { type: 'category', data: result.value.equity_curve.map((d: any) => d.date), axisLabel: { fontSize: 10 } },
        yAxis: { type: 'value', axisLabel: { fontSize: 10, formatter: (v: number) => `${(v / 10000).toFixed(0)}万` } },
        series: [{ type: 'line', data: result.value.equity_curve.map((d: any) => d.equity), smooth: true, lineStyle: { width: 2 }, areaStyle: { opacity: 0.15 } }],
      })
    }
  }
  if (drawdownChartRef.value && !ddChart) {
    ddChart = echarts.init(drawdownChartRef.value)
    if (result.value?.drawdown_curve?.length) {
      ddChart.setOption({
        grid: { left: 50, right: 10, top: 10, bottom: 30 },
        xAxis: { type: 'category', data: result.value.drawdown_curve.map((d: any) => d.date), axisLabel: { fontSize: 10 } },
        yAxis: { type: 'value', axisLabel: { fontSize: 10, formatter: (v: number) => `${(v * 100).toFixed(1)}%` } },
        series: [{ type: 'line', data: result.value.drawdown_curve.map((d: any) => d.drawdown), smooth: true, lineStyle: { width: 2, color: '#F56C6C' }, areaStyle: { color: 'rgba(245,108,108,0.12)' } }],
      })
    }
  }
  if (returnDistChartRef.value && !rdChart) {
    rdChart = echarts.init(returnDistChartRef.value)
    if (result.value?.return_distribution?.length) {
      rdChart.setOption({
        grid: { left: 50, right: 10, top: 10, bottom: 30 },
        xAxis: { type: 'category' },
        yAxis: { type: 'value', axisLabel: { fontSize: 10, formatter: (v: number) => `${(v * 100).toFixed(1)}%` } },
        series: [{ type: 'bar', data: result.value.return_distribution.map((r: number) => ({ value: r, itemStyle: { color: r >= 0 ? '#F56C6C' : '#67C23A' } })) }],
      })
    }
  }
}

function onResize() {
  eqChart?.resize()
  ddChart?.resize()
  rdChart?.resize()
}

async function runBacktest() {
  form.capital = Number(form.capitalStr) || 1000000
  if (!form.stockCode.trim()) {
    showToast('请输入股票代码')
    return
  }

  // 先移除旧监听器，避免重复
  safeGetWebSocket().off('backtest_progress', progressHandler)
  safeGetWebSocket().off('backtest_result', resultHandler)

  running.value = true
  progress.value = 0

  try {
    const res: any = await backend.post('/strategy/backtest/submit', {
      stock_code: form.stockCode,
      start_date: form.startDate,
      end_date: form.endDate,
      initial_capital: form.capital,
    })
    const taskId = res?.task_id || res?.data?.task_id

    safeGetWebSocket().on('backtest_progress', (data: any) => {
      if (data?.task_id === taskId) progress.value = data.progress || 0
    })

    safeGetWebSocket().on('backtest_result', (data: any) => {
      if (data?.task_id === taskId) handleResult(data)
    })

    // 轮询备用
    if (taskId) {
      for (let i = 0; i < 60; i++) {
        await new Promise((r) => setTimeout(r, 2000))
        try {
          const pollRes: any = await backend.get(`/strategy/backtest/result/${taskId}`)
          if (pollRes?.status === 'completed') {
            handleResult(pollRes)
            break
          }
        } catch { /* continue */ }
      }
    }
  } catch (e: any) {
    showToast({ message: e.message || '回测失败', type: 'fail' })
  } finally {
    running.value = false
  }
}

function handleResult(data: any) {
  if (result.value) return
  result.value = {
    metrics: [
      { label: '年化收益', value: `${(data.annual_return * 100).toFixed(2)}%`, color: data.annual_return >= 0 ? 'pnl-up' : 'pnl-down' },
      { label: '最大回撤', value: `${(data.max_drawdown * 100).toFixed(2)}%`, color: 'pnl-down' },
      { label: '夏普比率', value: data.sharpe_ratio?.toFixed(2) || '--' },
      { label: '胜率', value: `${(data.win_rate * 100).toFixed(1)}%` },
      { label: '盈亏比', value: data.profit_loss_ratio?.toFixed(2) || '--' },
      { label: '交易次数', value: String(data.total_trades || 0) },
    ],
    equity_curve: data.equity_curve || [],
    drawdown_curve: data.drawdown_curve || [],
    return_distribution: data.return_distribution || [],
    trades: data.trades || [],
  }
}
</script>

<style scoped>
.mobile-backtest {
  padding: 0 0 16px 0;
}

/* 表单 */
.backtest-form {
  margin-bottom: 12px;
}
.form-actions {
  padding: 12px 16px;
}

/* 指标卡片 */
.metrics-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 8px;
  margin: 0 12px 12px 12px;
}
.metric-card {
  background: #fff;
  border-radius: 8px;
  padding: 12px;
  text-align: center;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}
.metric-label {
  font-size: 12px;
  color: #969799;
  margin-bottom: 4px;
}
.metric-value {
  font-size: 16px;
  font-weight: 700;
  color: #323233;
}

/* 颜色 */
.pnl-up { color: #ee0a24; }
.pnl-down { color: #07c160; }

/* 图表区 */
.chart-section {
  margin: 0 12px 12px 12px;
  background: #fff;
  border-radius: 10px;
  padding: 12px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}
.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #323233;
  margin-bottom: 8px;
}

/* 成交记录 */
.trades-section {
  margin: 0 12px 12px 12px;
}
.trades-section .section-title {
  margin-bottom: 8px;
  padding-left: 4px;
}
.trade-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}
.trade-time {
  color: #969799;
  min-width: 80px;
}
.trade-stock {
  font-weight: 500;
  min-width: 80px;
}
.trade-price {
  font-weight: 600;
  margin-left: auto;
}
.trade-qty {
  color: #969799;
  min-width: 40px;
}
.trades-more {
  text-align: center;
  font-size: 13px;
  color: #969799;
  padding: 8px;
}
</style>
