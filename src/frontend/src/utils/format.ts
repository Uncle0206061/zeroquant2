// 格式化工具函数

/**
 * 格式化金额
 * 正数: ¥100.00
 * 负数: -¥100.00
 */
export function formatMoney(v: number): string {
  return v >= 0 ? `¥${v.toFixed(2)}` : `-¥${Math.abs(v).toFixed(2)}`
}

/**
 * 格式化日期时间
 * 2024-01-01T10:30:00Z -> 01-01 10:30
 */
export function formatDateTime(d: string): string {
  if (!d) return '-'
  return d.replace('T', ' ').substring(0, 19)
}

/**
 * 格式化日期
 * 2024-01-01 -> 01-01
 */
export function formatDate(d: string): string {
  if (!d) return '-'
  return d.substring(5, 10)
}

/**
 * 格式化涨跌幅
 * 0.05 -> +5.00%
 */
export function formatRate(v: number): string {
  const sign = v >= 0 ? '+' : ''
  return `${sign}${(v * 100).toFixed(2)}%`
}