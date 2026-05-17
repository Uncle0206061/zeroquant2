<template>
  <div class="market">
    <h2>实时行情</h2>

    <!-- 股票搜索 -->
    <el-row :gutter="16" class="search-row">
      <el-col :span="8">
        <el-input
          v-model="stockCode"
          placeholder="输入股票代码 (如 600519)"
          @keyup.enter="fetchOrderbook"
          clearable
        >
          <template #append>
            <el-button @click="fetchOrderbook">查询</el-button>
          </template>
        </el-input>
      </el-col>
      <el-col :span="4">
        <el-select v-model="selectedStock" placeholder="热门股票" @change="onStockSelect">
          <el-option label="浦发银行" value="600000" />
          <el-option label="贵州茅台" value="600519" />
          <el-option label="平安银行" value="000001" />
          <el-option label="宁德时代" value="300750" />
        </el-select>
      </el-col>
    </el-row>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading">
      <el-icon class="is-loading"><Loading /></el-icon>
      加载中...
    </div>

    <!-- 错误提示 -->
    <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" class="error-alert" />

    <!-- 盘口数据 -->
    <el-card v-if="orderbook" class="orderbook-card">
      <template #header>
        <div class="card-header">
          <span class="stock-name">{{ selectedStockName || stockCode }}</span>
          <span class="stock-code">{{ stockCode }}</span>
          <el-tag :type="priceChange >= 0 ? 'success' : 'danger'" size="small">
            {{ priceChange >= 0 ? '+' : '' }}{{ priceChange.toFixed(2) }}
          </el-tag>
        </div>
      </template>

      <el-row :gutter="20">
        <!-- 卖盘（右侧） -->
        <el-col :span="12">
          <div class="orderbook-section">
            <h4 class="section-title ask">卖出</h4>
            <div class="orderbook-header">
              <span>价格</span>
              <span>数量</span>
            </div>
            <div class="orderbook-list ask-list">
              <div v-for="(item, index) in orderbook.asks" :key="'ask-'+index" class="orderbook-row ask-row">
                <span class="price">{{ item[0].toFixed(2) }}</span>
                <span class="volume">{{ formatVolume(item[1]) }}</span>
                <div class="volume-bar ask-bar" :style="{ width: getBarWidth(item[1]) + '%' }"></div>
              </div>
            </div>
          </div>
        </el-col>

        <!-- 买盘（左侧） -->
        <el-col :span="12">
          <div class="orderbook-section">
            <h4 class="section-title bid">买入</h4>
            <div class="orderbook-header">
              <span>价格</span>
              <span>数量</span>
            </div>
            <div class="orderbook-list bid-list">
              <div v-for="(item, index) in orderbook.bids" :key="'bid-'+index" class="orderbook-row bid-row">
                <span class="price">{{ item[0].toFixed(2) }}</span>
                <span class="volume">{{ formatVolume(item[1]) }}</span>
                <div class="volume-bar bid-bar" :style="{ width: getBarWidth(item[1]) + '%' }"></div>
              </div>
            </div>
          </div>
        </el-col>
      </el-row>

      <!-- 最新价和时间 -->
      <div class="price-info">
        <span class="current-price">最新价: ¥{{ orderbook.price?.toFixed(2) || '--' }}</span>
        <span class="update-time">更新时间: {{ orderbook.timestamp || '--' }}</span>
        <el-tag size="small" type="info">{{ orderbook.source || 'unknown' }}</el-tag>
      </div>

      <!-- 刷新按钮 -->
      <div class="refresh-row">
        <el-button type="primary" @click="fetchOrderbook" :loading="loading">
          刷新盘口
        </el-button>
        <el-checkbox v-model="autoRefresh" @change="toggleAutoRefresh">
          自动刷新 ({{ refreshInterval }}s)
        </el-checkbox>
      </div>
    </el-card>

    <!-- 空状态 -->
    <el-empty v-else-if="!loading && !error" description="请输入股票代码查询盘口数据" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Loading } from '@element-plus/icons-vue'
import dataService from '@/api/data'

// 状态
const stockCode = ref('600519')
const selectedStock = ref('600519')
const loading = ref(false)
const error = ref('')
const orderbook = ref<any>(null)
const autoRefresh = ref(false)
const refreshInterval = ref(5)
let timer: ReturnType<typeof setInterval> | null = null

// 热门股票名称映射
const stockNames: Record<string, string> = {
  '600000': '浦发银行',
  '600519': '贵州茅台',
  '000001': '平安银行',
  '300750': '宁德时代',
}

const selectedStockName = computed(() => stockNames[stockCode.value] || '')

// 计算价格变化（模拟）
const priceChange = computed(() => {
  if (!orderbook.value?.price) return 0
  return orderbook.value.price - 50 // 模拟基准价
})

// 获取盘口数据
async function fetchOrderbook() {
  if (!stockCode.value) {
    error.value = '请输入股票代码'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const code = stockCode.value.trim()
    const data = await dataService.get(`/orderbook/${code}`)
    orderbook.value = data

    if (!data || !data.bids?.length) {
      error.value = '暂无盘口数据'
    }
  } catch (err: any) {
    error.value = err.message || '获取盘口数据失败'
    orderbook.value = null
  } finally {
    loading.value = false
  }
}

// 股票选择
function onStockSelect(val: string) {
  stockCode.value = val
  fetchOrderbook()
}

// 格式化成交量
function formatVolume(vol: number): string {
  if (vol >= 10000) {
    return (vol / 10000).toFixed(1) + '万'
  }
  return vol.toString()
}

// 获取进度条宽度
function getBarWidth(vol: number): number {
  if (!orderbook.value) return 0
  const maxVol = Math.max(
    ...orderbook.value.bids?.map((b: number[]) => b[1]) || [],
    ...orderbook.value.asks?.map((a: number[]) => a[1]) || []
  )
  return maxVol > 0 ? (vol / maxVol) * 100 : 0
}

// 自动刷新
function toggleAutoRefresh(val: boolean) {
  if (val) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

function startAutoRefresh() {
  stopAutoRefresh()
  timer = setInterval(() => {
    fetchOrderbook()
  }, refreshInterval.value * 1000)
}

function stopAutoRefresh() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

// 生命周期
onMounted(() => {
  fetchOrderbook()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.market {
  padding: 20px;
}

.market h2 {
  margin-bottom: 20px;
}

.search-row {
  margin-bottom: 20px;
}

.loading {
  text-align: center;
  padding: 40px;
  color: #909399;
}

.loading .el-icon {
  font-size: 24px;
  margin-right: 8px;
}

.error-alert {
  margin-bottom: 20px;
}

.orderbook-card {
  margin-top: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.stock-name {
  font-size: 18px;
  font-weight: bold;
}

.stock-code {
  color: #909399;
  font-size: 14px;
}

.orderbook-section {
  padding: 10px;
}

.section-title {
  margin: 0 0 10px 0;
  font-size: 14px;
  font-weight: bold;
}

.section-title.ask {
  color: #f56c6c;
}

.section-title.bid {
  color: #67c23a;
}

.orderbook-header {
  display: flex;
  justify-content: space-between;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
  font-size: 12px;
  color: #909399;
  margin-bottom: 8px;
}

.orderbook-list {
  font-size: 14px;
}

.orderbook-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 12px;
  position: relative;
  border-radius: 2px;
}

.orderbook-row .price {
  font-weight: bold;
  z-index: 1;
}

.orderbook-row .volume {
  color: #606266;
  z-index: 1;
}

.ask-row {
  background: linear-gradient(to left, rgba(245, 108, 108, 0.15), transparent);
}

.bid-row {
  background: linear-gradient(to left, rgba(103, 194, 58, 0.15), transparent);
}

.volume-bar {
  position: absolute;
  right: 0;
  top: 0;
  height: 100%;
  opacity: 0.3;
  z-index: 0;
}

.ask-bar {
  background: #f56c6c;
}

.bid-bar {
  background: #67c23a;
}

.price-info {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
}

.current-price {
  font-size: 20px;
  font-weight: bold;
  color: #409eff;
}

.update-time {
  color: #909399;
  font-size: 12px;
}

.refresh-row {
  margin-top: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
}
</style>
