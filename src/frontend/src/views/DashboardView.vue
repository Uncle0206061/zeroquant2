<template>
  <div class="dashboard">
    <h2>仪表盘</h2>
    <el-row :gutter="16">
      <el-col :span="6" v-for="s in stats" :key="s.label">
        <el-card shadow="hover">
          <template #header>{{ s.label }}</template>
          <div class="stat-value" :class="getValueClass(s.value, s.format)">
            {{ formatValue(s.value, s.format) }}
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import backend from '@/api/backend'
import { usePosition } from '@/composables/usePosition'

const { positions, totalMarketValue, totalProfitLoss } = usePosition()

// 统计数据
const stats = computed(() => [
  { label: '总资产', value: totalMarketValue.value, format: 'money' },
  { label: '今日盈亏', value: totalProfitLoss.value, format: 'money' },
  { label: '持仓数量', value: positions.value.length, format: 'count' },
  { label: '运行策略', value: 0, format: 'count' }, // TODO: 从 API 获取
])

// 格式化显示
function formatValue(v: number, fmt: string) {
  if (fmt === 'money') {
    const sign = v >= 0 ? '+' : ''
    return `${sign}¥${v.toFixed(2)}`
  }
  return v
}

function getValueClass(v: number, fmt: string) {
  if (fmt === 'money') {
    return v >= 0 ? 'value-up' : 'value-down'
  }
  return ''
}
</script>

<style scoped>
.dashboard h2 {
  margin-bottom: 20px;
}
.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #001529;
}
.stat-value.value-up {
  color: #F56C6C;
}
.stat-value.value-down {
  color: #67C23A;
}
</style>
