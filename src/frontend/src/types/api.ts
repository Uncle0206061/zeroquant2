// 统一 API 响应类型

export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// ========== 认证 ==========
export interface LoginReq {
  username: string
  password: string
}

export interface LoginRes {
  token: string
  user_id: number
}

export interface RegisterReq {
  username: string
  password: string
  email?: string
}

// ========== 持仓 ==========
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

// ========== 订单 ==========
export interface OrderItem {
  order_id: string
  created_at: string
  stock_code: string
  stock_name: string
  direction: number
  price: number
  quantity: number
  filled_qty: number
  status: string
  commission?: number
  avg_price?: number
  reject_reason?: string
}

// ========== 回测 ==========
export interface BacktestForm {
  stockCode: string
  startDate: string
  endDate: string
  capital: number
}

export interface BacktestResult {
  task_id: string
  status: 'pending' | 'completed'
  annual_return?: number
  max_drawdown?: number
  sharpe_ratio?: number
  win_rate?: number
  profit_loss_ratio?: number
  total_trades?: number
  equity_curve?: { date: string; equity: number }[]
  drawdown_curve?: { date: string; drawdown: number }[]
  return_distribution?: number[]
  trades?: Trade[]
}

export interface Trade {
  time: string
  stockCode: string
  stockName: string
  direction: 'buy' | 'sell'
  price: number
  quantity: number
  amount: number
}

// ========== 策略 ==========
export interface Factor {
  key: string
  name: string
  category: 'financial' | 'technical' | 'sector'
  params: FactorParam[]
}

export interface FactorParam {
  key: string
  label: string
  type: 'number' | 'select' | 'tags'
  default?: any
  options?: { label: string; value: string }[]
  min?: number
  max?: number
  step?: number
  precision?: number
  unit?: string
}

export interface StrategyConfig {
  name: string
  stock_filter: Record<string, any>
  timing: { type: string; period?: number }
  risk: {
    max_position: number
    stop_loss: number
    stop_profit: number
    max_hold_days?: number
  }
}