<template>
  <div class="backtest-page">
    <div class="page-header">
      <h2>回测分析</h2>
    </div>

    <!-- 发起回测表单 -->
    <el-card class="backtest-form-card">
      <el-form :model="form" label-width="100px" inline>
        <el-form-item label="股票代码">
          <el-input
            v-model="form.stockCode"
            placeholder="如 000001.SZ"
            style="width: 160px"
            clearable
          />
        </el-form-item>
        <el-form-item label="开始日期">
          <el-date-picker
            v-model="form.startDate"
            type="date"
            value-format="YYYY-MM-DD"
            style="width: 160px"
          />
        </el-form-item>
        <el-form-item label="结束日期">
          <el-date-picker
            v-model="form.endDate"
            type="date"
            value-format="YYYY-MM-DD"
            style="width: 160px"
          />
        </el-form-item>
        <el-form-item label="初始资金">
          <el-input-number
            v-model="form.capital"
            :min="10000"
            :step="10000"
            :max="100000000"
            controls-position="right"
            style="width: 180px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="runBacktest" :loading="running">
            <el-icon v-if="!running"><VideoPlay /></el-icon>
            {{ running ? `回测中 ${progress}%` : '发起回测' }}
          </el-button>
        </el-form-item>
      </el-form>
      <!-- 进度条 -->
      <el-progress
        v-if="running"
        :percentage="progress"
        :stroke-width="6"
        :show-text="true"
        style="margin-top: 8px"
      />
    </el-card>

    <!-- 回测加载骨架屏 -->
    <template v-if="running && !result">
      <el-row :gutter="12">
        <el-col :xs="12" :sm="8" :md="4" v-for="n in 6" :key="n">
          <el-card shadow="hover" class="metric-card">
            <el-skeleton :rows="0" animated>
              <template #template>
                <div class="skeleton-label" />
                <div class="skeleton-value" />
              </template>
            </el-skeleton>
          </el-card>
        </el-col>
      </el-row>
      <el-row :gutter="12">
        <el-col :xs="24" :md="8" v-for="n in 3" :key="n">
          <el-card shadow="hover">
            <template #header>
              <el-skeleton :rows="0" animated style="width: 80px">
                <template #template><el-skeleton-item variant="text" /></template>
              </el-skeleton>
            </template>
            <el-skeleton :rows="6" animated />
          </el-card>
        </el-col>
      </el-row>
    </template>

    <!-- 回测结果区域 -->
    <template v-if="result">
      <!-- 核心指标卡片 -->
      <el-row :gutter="12" class="metrics-row">
        <el-col :xs="12" :sm="8" :md="4" v-for="m in result.metrics" :key="m.label">
          <el-card shadow="hover" class="metric-card">
            <div class="metric-label">{{ m.label }}</div>
            <div class="metric-value" :class="m.color">{{ m.value }}</div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 图表区域 -->
      <el-row :gutter="12" class="charts-row">
        <el-col :xs="24" :md="8">
          <el-card shadow="hover">
            <template #header>资金曲线</template>
            <div ref="equityChartRef" style="height: 300px" />
          </el-card>
        </el-col>
        <el-col :xs="24" :md="8">
          <el-card shadow="hover">
            <template #header>回撤曲线</template>
            <div ref="drawdownChartRef" style="height: 300px" />
          </el-card>
        </el-col>
        <el-col :xs="24" :md="8">
          <el-card shadow="hover">
            <template #header>收益分布</template>
            <div ref="returnDistChartRef" style="height: 300px" />
          </el-card>
        </el-col>
      </el-row>

      <!-- 成交记录表 -->
      <el-card class="trades-card">
        <template #header>
          <span>成交记录（共 {{ result.trades.length }} 条）</span>
        </template>
        <el-table :data="result.trades" stripe max-height="400" style="width: 100%">
          <el-table-column prop="time" label="时间" width="170" />
          <el-table-column prop="stockCode" label="代码" width="120" />
          <el-table-column prop="stockName" label="名称" width="100" />
          <el-table-column prop="direction" label="方向" width="80">
            <template #default="{ row }">
              <el-tag :type="row.direction === 'buy' ? 'danger' : 'success'" size="small">
                {{ row.direction === 'buy' ? '买入' : '卖出' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="price" label="价格" width="110">
            <template #default="{ row }">¥{{ row.price?.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column prop="quantity" label="数量" width="80" />
          <el-table-column prop="amount" label="金额" width="130">
            <template #default="{ row }">¥{{ row.amount?.toLocaleString() }}</template>
          </el-table-column>
        </el-table>
        <el-empty v-if="result.trades.length === 0" description="暂无成交记录" />
      </el-card>
    </template>

    <!-- 空状态 -->
    <el-card v-else-if="!running" class="empty-card">
      <el-empty description="发起回测后，结果将在此展示" :image-size="120" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { VideoPlay } from '@element-plus/icons-vue'
import { useBacktest } from '@/composables/useBacktest'

const {
  form,
  running,
  progress,
  result,
  runBacktest,
  renderCharts,
  disposeCharts,
} = useBacktest()

// 图表 DOM refs
const equityChartRef = ref<HTMLElement | null>(null)
const drawdownChartRef = ref<HTMLElement | null>(null)
const returnDistChartRef = ref<HTMLElement | null>(null)

// 监听 result 变化，自动渲染图表
watch(
  () => result.value,
  async (val) => {
    if (val) {
      await nextTick()
      renderCharts(equityChartRef, drawdownChartRef, returnDistChartRef)
    }
  },
  { once: true },
)

// 窗口 resize 时重新渲染图表
function onResize() {
  renderCharts(equityChartRef, drawdownChartRef, returnDistChartRef)
}

onMounted(() => {
  window.addEventListener('resize', onResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  disposeCharts()
})
</script>

<style scoped>
.backtest-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-header h2 {
  margin: 0;
}

.backtest-form-card {
  border-radius: 8px;
}

.metrics-row {
  margin: 0;
}

.metric-card {
  text-align: center;
  border-radius: 8px;
  cursor: default;
}

.metric-label {
  font-size: 13px;
  color: #909399;
  margin-bottom: 8px;
}

.metric-value {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
}

.metric-value.positive {
  color: #F56C6C;
}

.metric-value.negative {
  color: #67C23A;
}

.charts-row {
  margin: 0;
}

.trades-card {
  border-radius: 8px;
}

.empty-card {
  border-radius: 8px;
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.skeleton-label {
  height: 14px;
  background: #e5e7eb;
  border-radius: 4px;
  margin-bottom: 8px;
}

.skeleton-value {
  height: 28px;
  background: #e5e7eb;
  border-radius: 4px;
  width: 60%;
  margin: 0 auto;
}
</style>
