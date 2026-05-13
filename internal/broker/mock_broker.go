// Package broker 提供 Mock 券商实现（Phase 1）
// Phase 2 将替换为真实券商 SDK（如同花顺/东方财富）
package broker

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/Uncle0206061/zeroquant2/backend/internal/model"
)

// MockBroker Phase 1 模拟券商（返回 mock 数据，接口结构完整）
type MockBroker struct{}

// NewMockBroker 创建 MockBroker 实例
func NewMockBroker() *MockBroker {
	return &MockBroker{}
}

// SubmitOrder 模拟下单（Phase 1：返回 mock 数据）
func (m *MockBroker) SubmitOrder(req *RealOrderRequest) (*RealOrderResponse, error) {
	// 生成模拟订单号
	mockOrderID := "mock_" + uuid.New().String()[:8]
	mockBrokerID := "broker_mock_" + uuid.New().String()[:8]

	return &RealOrderResponse{
		OrderID:       mockOrderID,
		Status:        "mock_submitted",
		BrokerOrderID: mockBrokerID,
		MockData:      true,
	}, nil
}

// GetPosition 模拟持仓查询（Phase 1：返回空持仓）
func (m *MockBroker) GetPosition(userID int64) ([]model.RealPosition, error) {
	// Phase 1：返回空列表
	return []model.RealPosition{}, nil
}

// GetAccountInfo 模拟账户查询（Phase 1：返回初始资金）
func (m *MockBroker) GetAccountInfo(userID int64) (*RealAccountResponse, error) {
	return &RealAccountResponse{
		UserID:     userID,
		Balance:    1000000,
		TotalAsset: 1000000,
		MockData:   true,
	}, nil
}

// CancelOrder 模拟撤单（Phase 1：直接成功）
func (m *MockBroker) CancelOrder(orderID string) error {
	// Phase 1：模拟撤单成功
	if orderID == "" {
		return fmt.Errorf("order_id 不能为空")
	}
	return nil
}
