<template>
  <div class="position-page">
    <div class="page-header">
      <h2>持仓</h2>
      <el-button :icon="Refresh" @click="fetchPositions" :loading="loading" text>
        刷新
      </el-button>
    </div>

    <!-- 持仓表格 -->
    <el-table
      :data="positions"
      v-loading="loading"
      stripe
      :default-sort="{ prop: 'profitRate', order: 'descending' }"
      style="width: 100%"
    >
      <el-table-column label="代码" prop="stockCode" width="110" fixed />
      <el-table-column label="名称" prop="stockName" width="100" fixed>
        <template #default="{ row }">
          <span class="stock-name">{{ row.stockName }}</span>
        </template>
      </el-table-column>
      <el-table-column label="持仓量" prop="quantity" width="90" align="right" />
      <el-table-column label="可用量" prop="availableQty" width="90" align="right" />
      <el-table-column label="成本价" width="100" align="right">
        <template #default="{ row }">¥{{ row.avgCost.toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="现价" width="100" align="right">
        <template #default="{ row }">
          <span :class="row.currentPrice >= row.avgCost ? 'price-up' : 'price-down'">
            ¥{{ row.currentPrice.toFixed(2) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="市值" width="120" align="right">
        <template #default="{ row }">{{ formatMoney(row.marketValue) }}</template>
      </el-table-column>
      <el-table-column label="盈亏" prop="profitLoss" width="120" align="right" sortable>
        <template #default="{ row }">
          <span :class="row.profitLoss >= 0 ? 'pnl-up' : 'pnl-down'">
            {{ row.profitLoss >= 0 ? '+' : '' }}{{ formatMoney(row.profitLoss) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="盈亏%" prop="profitRate" width="100" align="right" sortable>
        <template #default="{ row }">
          <span :class="row.profitRate >= 0 ? 'pnl-up' : 'pnl-down'">
            {{ row.profitRate >= 0 ? '+' : '' }}{{ (row.profitRate * 100).toFixed(2) }}%
          </span>
        </template>
      </el-table-column>
      <el-table-column label="占比" width="100" align="right">
        <template #default="{ row }">
          <el-progress
            v-if="totalMarketValue > 0"
            :percentage="Number(((row.marketValue / totalMarketValue) * 100).toFixed(1))"
            :stroke-width="14"
            :text-inside="true"
            :show-text="false"
            :color="row.profitRate >= 0 ? '#F56C6C' : '#67C23A'"
          />
        </template>
      </el-table-column>
    </el-table>

    <!-- 空状态 -->
    <el-empty v-if="!loading && positions.length === 0" description="暂无持仓" />

    <!-- 底部汇总 -->
    <div class="position-summary" v-if="positions.length > 0">
      <div class="summary-item">
        <span class="summary-label">持仓数量</span>
        <span class="summary-value">{{ positions.length }} 只</span>
      </div>
      <div class="summary-item">
        <span class="summary-label">总市值</span>
        <span class="summary-value">¥{{ totalMarketValue.toFixed(2) }}</span>
      </div>
      <div class="summary-item">
        <span class="summary-label">总盈亏</span>
        <span class="summary-value" :class="totalProfitLoss >= 0 ? 'pnl-up' : 'pnl-down'">
          {{ totalProfitLoss >= 0 ? '+' : '' }}¥{{ totalProfitLoss.toFixed(2) }}
        </span>
      </div>
      <div class="summary-item">
        <span class="summary-label">胜率</span>
        <span class="summary-value">{{ winRate }}%</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Refresh } from '@element-plus/icons-vue'
import { usePosition } from '@/composables/usePosition'

const {
  positions,
  loading,
  fetchPositions,
  totalMarketValue,
  totalProfitLoss,
  winRate,
} = usePosition()

function formatMoney(v: number): string {
  return v >= 0 ? `¥${v.toFixed(2)}` : `-¥${Math.abs(v).toFixed(2)}`
}
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.stock-name {
  font-weight: 600;
  color: #303133;
}

/* A 股红涨绿跌 */
.price-up,
.pnl-up {
  color: #F56C6C;
  font-weight: 600;
}

.price-down,
.pnl-down {
  color: #67C23A;
  font-weight: 600;
}

.position-summary {
  display: flex;
  gap: 32px;
  padding: 16px 20px;
  margin-top: 16px;
  background: #fff;
  border-radius: 8px;
  border: 1px solid #e8e8e8;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.summary-label {
  font-size: 13px;
  color: #909399;
}

.summary-value {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}
</style>
