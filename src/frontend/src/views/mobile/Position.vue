<template>
  <div class="mobile-position">
    <!-- 账户概览卡片 -->
    <div class="overview-card">
      <div class="overview-label">总市值</div>
      <div class="overview-value">¥{{ totalMarketValue.toFixed(2) }}</div>
      <div class="overview-pnl" :class="totalProfitLoss >= 0 ? 'pnl-up' : 'pnl-down'">
        {{ totalProfitLoss >= 0 ? '+' : '' }}¥{{ totalProfitLoss.toFixed(2) }}
        <span class="overview-rate">
          ({{ totalProfitLossRate }})
        </span>
      </div>
      <div class="overview-stats">
        <div class="stat-item">
          <span class="stat-label">持仓</span>
          <span class="stat-value">{{ positions.length }}只</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">胜率</span>
          <span class="stat-value">{{ winRate }}%</span>
        </div>
      </div>
    </div>

    <!-- 持仓列表 -->
    <van-pull-refresh v-model="refreshing" @refresh="onRefresh" :disabled="!isOnline">
      <van-list v-model:loading="loading" :finished="true" finished-text="">
        <div
          v-for="pos in positions"
          :key="pos.id"
          class="pos-card"
          @click="showDetail(pos)"
        >
          <div class="pos-header">
            <span class="pos-name">{{ pos.stockName }}</span>
            <span class="pos-code">{{ pos.stockCode }}</span>
          </div>
          <div class="pos-body">
            <div class="pos-price-info">
              <div class="pos-current" :class="pos.profitRate >= 0 ? 'pnl-up' : 'pnl-down'">
                ¥{{ pos.currentPrice.toFixed(2) }}
              </div>
              <div class="pos-cost">成本 ¥{{ pos.avgCost.toFixed(2) }}</div>
            </div>
            <div class="pos-pnl-info">
              <div class="pos-rate" :class="pos.profitRate >= 0 ? 'pnl-up' : 'pnl-down'">
                {{ pos.profitRate >= 0 ? '+' : '' }}{{ (pos.profitRate * 100).toFixed(2) }}%
              </div>
              <div class="pos-amount" :class="pos.profitLoss >= 0 ? 'pnl-up' : 'pnl-down'">
                {{ pos.profitLoss >= 0 ? '+' : '' }}{{ formatMoney(pos.profitLoss) }}
              </div>
            </div>
          </div>
          <div class="pos-footer">
            <span>持仓 {{ pos.quantity }} 股</span>
            <span>市值 {{ formatMoney(pos.marketValue) }}</span>
          </div>
        </div>
        <van-empty v-if="!loading && positions.length === 0" description="暂无持仓" />
      </van-list>
    </van-pull-refresh>

    <!-- 持仓详情弹窗 -->
    <van-popup v-model:show="detailVisible" position="bottom" round :style="{ height: '50%' }">
      <div class="detail-panel" v-if="currentPos">
        <div class="detail-header">
          <span class="detail-name">{{ currentPos.stockName }}</span>
          <span class="detail-code">{{ currentPos.stockCode }}</span>
        </div>
        <van-cell-group inset>
          <van-cell title="持仓数量" :value="currentPos.quantity + ' 股'" />
          <van-cell title="可用数量" :value="(currentPos.availableQty || 0) + ' 股'" />
          <van-cell title="成本价" :value="'¥' + currentPos.avgCost.toFixed(2)" />
          <van-cell title="当前价" :value="'¥' + currentPos.currentPrice.toFixed(2)" />
          <van-cell title="市值" :value="formatMoney(currentPos.marketValue)" />
          <van-cell title="盈亏金额" :value="formatMoney(currentPos.profitLoss)" />
          <van-cell title="盈亏比例" :value="(currentPos.profitRate * 100).toFixed(2) + '%'" />
        </van-cell-group>
      </div>
    </van-popup>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { usePosition } from '@/composables/usePosition'
import { useNetworkStatus } from '@/composables/useNetworkStatus'

const { isOnline } = useNetworkStatus()
const { positions, loading, fetchPositions, totalMarketValue, totalProfitLoss, winRate } =
  usePosition()

const refreshing = ref(false)
const detailVisible = ref(false)
const currentPos = ref<any>(null)

// 总盈亏比例（粗略：总盈亏/总市值）
const totalProfitLossRate = computed(() => {
  if (totalMarketValue.value <= 0) return '0.00%'
  const rate = totalProfitLoss.value / (totalMarketValue.value - totalProfitLoss.value) * 100
  return (rate >= 0 ? '+' : '') + rate.toFixed(2) + '%'
})

function formatMoney(v: number): string {
  return v >= 0 ? `¥${v.toFixed(2)}` : `-¥${Math.abs(v).toFixed(2)}`
}

async function onRefresh() {
  if (!isOnline.value) {
    refreshing.value = false
    return
  }
  await fetchPositions()
  refreshing.value = false
}

function showDetail(pos: any) {
  currentPos.value = pos
  detailVisible.value = true
}
</script>

<style scoped>
.mobile-position {
  padding: 0 0 16px 0;
}

/* 概览卡片 */
.overview-card {
  margin: 12px;
  padding: 20px;
  background: linear-gradient(135deg, #1a1a2e, #16213e);
  border-radius: 12px;
  color: #fff;
}
.overview-label {
  font-size: 13px;
  color: #ffffffa6;
  margin-bottom: 4px;
}
.overview-value {
  font-size: 28px;
  font-weight: 700;
  margin-bottom: 8px;
}
.overview-pnl {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 12px;
}
.overview-rate {
  font-size: 14px;
  margin-left: 4px;
  opacity: 0.9;
}
.overview-stats {
  display: flex;
  gap: 32px;
}
.stat-label {
  font-size: 12px;
  color: #ffffff80;
}
.stat-value {
  font-size: 15px;
  font-weight: 600;
}

/* A 股红涨绿跌 */
.pnl-up { color: #ee0a24; }
.pnl-down { color: #07c160; }

/* 持仓卡片 */
.pos-card {
  margin: 0 12px 8px 12px;
  padding: 14px 16px;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}
.pos-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}
.pos-name {
  font-size: 15px;
  font-weight: 600;
  color: #323233;
}
.pos-code {
  font-size: 12px;
  color: #969799;
  font-family: 'Consolas', monospace;
}
.pos-body {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 8px;
}
.pos-current {
  font-size: 20px;
  font-weight: 700;
}
.pos-cost {
  font-size: 12px;
  color: #969799;
  margin-top: 2px;
}
.pos-rate {
  font-size: 18px;
  font-weight: 700;
  text-align: right;
}
.pos-amount {
  font-size: 12px;
  text-align: right;
  margin-top: 2px;
}
.pos-footer {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #969799;
  padding-top: 8px;
  border-top: 1px solid #f5f5f5;
}

/* 详情面板 */
.detail-panel {
  padding: 8px 0;
}
.detail-header {
  padding: 16px;
  text-align: center;
}
.detail-name {
  font-size: 18px;
  font-weight: 600;
  color: #323233;
}
.detail-code {
  margin-left: 8px;
  font-size: 14px;
  color: #969799;
}
</style>
