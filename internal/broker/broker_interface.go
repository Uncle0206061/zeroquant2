// Package broker 提供券商接口定义与实现
// Phase 1：MockBroker（返回模拟数据，框架完整）
// Phase 2：接入真实券商 SDK（如同花顺/东方财富）
package broker

import "github.com/Uncle0206061/zeroquant2/backend/internal/model"

// RealOrderRequest 实盘下单请求
type RealOrderRequest struct {
	UserID    int64   `json:"user_id"`
	StockCode string  `json:"stock_code"`
	StockName string  `json:"stock_name"`
	Direction int8    `json:"direction"` // 1=买入 2=卖出
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	OrderType int8    `json:"order_type"` // 1=市价 2=限价
}

// RealOrderResponse 实盘下单响应
type RealOrderResponse struct {
	OrderID       string `json:"order_id"`
	Status        string `json:"status"` // mock_submitted / submitted / filled
	BrokerOrderID string `json:"broker_order_id"`
	MockData      bool   `json:"mock_data"` // true 表示 Phase 1 模拟数据
}

// RealAccountResponse 实盘账户查询响应
type RealAccountResponse struct {
	UserID     int64   `json:"user_id"`
	Balance    float64 `json:"balance"`
	TotalAsset float64 `json:"total_asset"`
	MockData   bool    `json:"mock_data"`
}

// BrokerInterface 券商接口（所有券商 SDK 必须实现）
type BrokerInterface interface {
	// SubmitOrder 提交委托
	SubmitOrder(req *RealOrderRequest) (*RealOrderResponse, error)
	// GetPosition 查询实盘持仓
	GetPosition(userID int64) ([]model.RealPosition, error)
	// GetAccountInfo 查询实盘账户
	GetAccountInfo(userID int64) (*RealAccountResponse, error)
	// CancelOrder 撤销委托
	CancelOrder(orderID string) error
}
