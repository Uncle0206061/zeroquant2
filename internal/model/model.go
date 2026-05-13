// Package model 提供数据模型定义
// 表命名规范：biz_ 前缀（如 biz_user, biz_watchlist）
package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID           int64      `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string     `gorm:"uniqueIndex;size:100" json:"email"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Nickname     string     `gorm:"size:50" json:"nickname"`
	Role        string     `gorm:"size:20;default:user" json:"role"` // admin, user
	Status      int        `gorm:"default:1" json:"status"` // 1=正常, 0=禁用
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// UserProfile 用户扩展信息
type UserProfile struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	UserID      int64     `gorm:"uniqueIndex;not null" json:"user_id"`
	Avatar      string    `gorm:"size:255" json:"avatar"`
	Bio         string    `gorm:"type:text" json:"bio"`
	Phone       string    `gorm:"size:20" json:"phone"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Watchlist 自选股
type Watchlist struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    int64    `gorm:"index;not null" json:"user_id"`
	StockCode string   `gorm:"size:10;index" json:"stock_code"` // 股票代码：000001
	StockName string   `gorm:"size:50" json:"stock_name"` // 股票名称：平安银行
	Notes     string   `gorm:"type:text" json:"notes"` // 备注
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Portfolio 模拟账户持仓（每个用户一条记录）
type Portfolio struct {
	ID              int64   `gorm:"primaryKey" json:"id"`
	UserID          int64   `gorm:"uniqueIndex;not null" json:"user_id"`
	Balance         float64 `gorm:"not null;default:1000000" json:"balance"`         // 可用资金
	Frozen          float64 `gorm:"not null;default:0" json:"frozen"`               // 冻结资金
	TotalAsset      float64 `gorm:"not null;default:1000000" json:"total_asset"`     // 总资产
	DailyLoss       float64 `gorm:"not null;default:0" json:"daily_loss"`            // 当日亏损额
	DailyTrades     int     `gorm:"not null;default:0" json:"daily_trades"`          // 当日交易次数
	FrozenTrading   bool    `gorm:"not null;default:false" json:"frozen_trading"`   // 当日交易冻结
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Order 逐笔委托
type Order struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	TaskID       string    `gorm:"size:64;index" json:"task_id"`               // 策略任务ID
	UserID       int64     `gorm:"index;not null" json:"user_id"`
	OrderID      string    `gorm:"size:64;uniqueIndex" json:"order_id"`       // 委托单号（UUID）
	StockCode    string    `gorm:"size:10;index" json:"stock_code"`
	StockName    string    `gorm:"size:50" json:"stock_name"`
	Direction    int8      `gorm:"not null" json:"direction"`                 // 1=买入 2=卖出
	OrderType    int8      `gorm:"not null;default:1" json:"order_type"`       // 1=市价委托
	Price        float64   `json:"price"`                                       // 委托价格（0=市价）
	Quantity     int       `gorm:"not null" json:"quantity"`                  // 委托数量（手×100）
	FilledQty    int       `gorm:"default:0" json:"filled_qty"`              // 成交数量
	AvgPrice     float64   `json:"avg_price"`                                   // 成交均价
	Status       string    `gorm:"size:16;default:pending" json:"status"`     // pending/filled/cancelled/rejected
	RejectReason string    `gorm:"size:256" json:"reject_reason"`              // 拒绝原因
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Position 持仓（按用户+股票唯一）
type Position struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	UserID       int64     `gorm:"index;not null" json:"user_id"`
	StockCode    string    `gorm:"size:10;index" json:"stock_code"`
	StockName    string    `gorm:"size:50" json:"stock_name"`
	Quantity     int       `gorm:"not null;default:0" json:"quantity"`         // 持仓数量
	AvailableQty int       `gorm:"not null;default:0" json:"available_qty"`   // 可卖数量
	AvgCost      float64   `gorm:"not null;default:0" json:"avg_cost"`        // 成本价
	MarketValue  float64   `json:"market_value"`                                // 市值
	ProfitLoss   float64   `json:"profit_loss"`                                 // 盈亏金额
	ProfitRate   float64   `json:"profit_rate"`                                 // 盈亏比例
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Strategy 策略
type Strategy struct {
	ID             int64     `gorm:"primaryKey" json:"id"`
	UserID         int64     `gorm:"index;not null" json:"user_id"`
	Name           string    `gorm:"size:100;not null" json:"name"`
	Description    string    `gorm:"type:text" json:"description"`
	Rules          string    `gorm:"type:jsonb" json:"rules"` // 策略规则 JSON（完整规则树）
	Status         int       `gorm:"default:1" json:"status"`   // 0=草稿 1=启用 2=暂停
	LastSubmitAt   *time.Time `json:"last_submit_at"`
	LastSubmitMode string    `gorm:"size:20" json:"last_submit_mode"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// StrategyRule 策略规则
type StrategyRule struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	StrategyID int64     `gorm:"index;not null" json:"strategy_id"`
	RuleType   string    `gorm:"size:20;not null" json:"rule_type"` // stock_filter/timing/risk
	RuleKey    string    `gorm:"size:50;not null" json:"rule_key"`
	RuleValue  string    `gorm:"type:jsonb" json:"rule_value"`
	Priority   int       `gorm:"default:0" json:"priority"`
	Enabled    bool      `gorm:"default:true" json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Backtest 回测/执行任务
type Backtest struct {
	ID            int64      `gorm:"primaryKey" json:"id"`
	StrategyID    int64      `gorm:"index;not null" json:"strategy_id"`
	UserID        int64      `gorm:"index;not null" json:"user_id"`
	Mode          string     `gorm:"size:20;not null" json:"mode"`       // realtime/historian
	Status        string     `gorm:"size:20;default:pending" json:"status"` // pending/running/completed/failed
	TaskID        string     `gorm:"size:36;uniqueIndex" json:"task_id"`
	ResultSummary string     `gorm:"type:jsonb" json:"result_summary"`
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Alert 告警
type Alert struct {
	ID        int64   `gorm:"primaryKey" json:"id"`
	UserID   int64  `gorm:"index;not null" json:"user_id"`
	Type    string `gorm:"size:50;not null" json:"type"` // price, change, volume
	StockCode string `gorm:"size:10;index" json:"stock_code"`
	Condition string `gorm:"size:100" json:"condition"` // 触发条件
	Message string  `gorm:"type:text" json:"message"` // 告警消息
	Status  int     `gorm:"default:1" json:"status"` // 1=启用, 0=停用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}