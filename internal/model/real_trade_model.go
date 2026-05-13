// Package model 提供实盘交易数据模型定义
// 表命名规范：biz_real_ 前缀（与模拟盘 biz_ 前缀完全隔离）
package model

import "time"

// RealPortfolio 实盘账户（每个用户一条记录，与模拟盘 Portfolio 完全隔离）
type RealPortfolio struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	UserID        int64     `gorm:"uniqueIndex;not null" json:"user_id"`
	Balance       float64   `gorm:"not null;default:1000000" json:"balance"`
	Frozen        float64   `gorm:"not null;default:0" json:"frozen"`
	TotalAsset    float64   `gorm:"not null;default:1000000" json:"total_asset"`
	DailyLoss     float64   `gorm:"not null;default:0" json:"daily_loss"`
	DailyTrades   int       `gorm:"not null;default:0" json:"daily_trades"`
	FrozenTrading bool      `gorm:"not null;default:false" json:"frozen_trading"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName 指定表名
func (RealPortfolio) TableName() string { return "biz_real_portfolio" }

// RealOrder 实盘逐笔委托（与模拟盘 Order 完全隔离）
type RealOrder struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	UserID       int64     `gorm:"index;not null" json:"user_id"`
	OrderID      string    `gorm:"size:64;uniqueIndex" json:"order_id"`
	StockCode    string    `gorm:"size:10;index" json:"stock_code"`
	StockName    string    `gorm:"size:50" json:"stock_name"`
	Direction    int8      `gorm:"not null" json:"direction"`           // 1=买入 2=卖出
	OrderType    int8      `gorm:"not null;default:1" json:"order_type"` // 1=市价 2=限价
	Price        float64   `json:"price"`
	Quantity     int       `gorm:"not null" json:"quantity"`
	FilledQty    int       `gorm:"default:0" json:"filled_qty"`
	AvgPrice     float64   `json:"avg_price"`
	Status       string    `gorm:"size:16;default:pending" json:"status"` // pending/confirmed/filled/cancelled/failed
	RejectReason string    `gorm:"size:256" json:"reject_reason"`
	BrokerOrderID string   `gorm:"size:64" json:"broker_order_id"`        // 券商委托号
	IsConfirmed  bool      `gorm:"not null;default:false" json:"is_confirmed"` // 二次确认状态
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (RealOrder) TableName() string { return "biz_real_order" }

// RealPosition 实盘持仓（与模拟盘 Position 完全隔离）
type RealPosition struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	UserID       int64     `gorm:"index;not null" json:"user_id"`
	StockCode    string    `gorm:"size:10;index" json:"stock_code"`
	StockName    string    `gorm:"size:50" json:"stock_name"`
	Quantity     int       `gorm:"not null;default:0" json:"quantity"`
	AvailableQty int       `gorm:"not null;default:0" json:"available_qty"`
	AvgCost      float64   `gorm:"not null;default:0" json:"avg_cost"`
	MarketValue  float64   `json:"market_value"`
	ProfitLoss   float64   `json:"profit_loss"`
	ProfitRate   float64   `json:"profit_rate"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (RealPosition) TableName() string { return "biz_real_position" }

// RealTradeLog 实盘操作日志（所有实盘操作全量记录）
type RealTradeLog struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	UserID        int64     `gorm:"index;not null" json:"user_id"`
	OrderID       string    `gorm:"size:64;index" json:"order_id"`
	StockCode     string    `gorm:"size:10;not null" json:"stock_code"`
	Direction     string    `gorm:"size:4;not null" json:"direction"`    // buy/sell
	Price         float64   `gorm:"not null" json:"price"`
	Quantity      int       `gorm:"not null" json:"quantity"`
	OrderType     string    `gorm:"size:10;not null" json:"order_type"`  // limit/market
	Status        string    `gorm:"size:20;not null" json:"status"`      // submitted/confirmed/filled/cancelled/failed
	BrokerOrderID string    `gorm:"size:64" json:"broker_order_id"`
	ErrorMessage  string    `gorm:"type:text" json:"error_message"`
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
}

// TableName 指定表名
func (RealTradeLog) TableName() string { return "biz_real_trade_log" }
