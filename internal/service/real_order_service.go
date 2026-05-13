// Package service 提供实盘交易业务逻辑层
// 核心安全机制：二次确认 + Mutex 锁 + 风控前置 + 操作日志全量记录
package service

import (
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/Uncle0206061/zeroquant2/backend/internal/broker"
	"github.com/Uncle0206061/zeroquant2/backend/internal/middleware"
	"github.com/Uncle0206061/zeroquant2/backend/internal/model"
	"github.com/Uncle0206061/zeroquant2/backend/internal/repository"
	ws "github.com/Uncle0206061/zeroquant2/backend/internal/websocket"
)

// RealOrderService 实盘交易服务
type RealOrderService struct {
	realOrderRepo    *repository.RealOrderRepository
	realPositionRepo *repository.RealPositionRepository
	realPortfolioRepo *repository.RealPortfolioRepository
	realTradeLogRepo *repository.RealTradeLogRepository
	brokerImpl       broker.BrokerInterface
	orderMutex       sync.Mutex // 下单串行锁（与模拟盘一致）
}

// NewRealOrderService 创建 RealOrderService
func NewRealOrderService(
	realOrderRepo *repository.RealOrderRepository,
	realPositionRepo *repository.RealPositionRepository,
	realPortfolioRepo *repository.RealPortfolioRepository,
	realTradeLogRepo *repository.RealTradeLogRepository,
	brokerImpl broker.BrokerInterface,
) *RealOrderService {
	return &RealOrderService{
		realOrderRepo:    realOrderRepo,
		realPositionRepo: realPositionRepo,
		realPortfolioRepo: realPortfolioRepo,
		realTradeLogRepo: realTradeLogRepo,
		brokerImpl:       brokerImpl,
	}
}

// CreateRealAccount 创建实盘账户（需 REAL_TRADE=true）
func (s *RealOrderService) CreateRealAccount(userID int64) (*model.RealPortfolio, error) {
	if !middleware.IsRealTradeMode() {
		return nil, errors.New("实盘模式未开启，请设置 REAL_TRADE=true")
	}

	// 查重
	existing, err := s.realPortfolioRepo.FindByUserID(userID)
	if err == nil && existing != nil {
		return existing, nil
	}

	p := &model.RealPortfolio{
		UserID:     userID,
		Balance:    1000000, // Phase 1 初始资金，实盘由券商同步
		Frozen:     0,
		TotalAsset: 1000000,
	}

	if err := s.realPortfolioRepo.Create(p); err != nil {
		return nil, err
	}

	// 记录操作日志
	s.logTrade(userID, "", "CREATE_ACCOUNT", "", 0, 0, "account", "confirmed", "", "")

	return p, nil
}

// GetRealAccount 获取实盘账户
func (s *RealOrderService) GetRealAccount(userID int64) (*model.RealPortfolio, error) {
	return s.realPortfolioRepo.FindByUserID(userID)
}

// SubmitRealOrder 实盘下单（第一次提交 → pending 等待确认）
func (s *RealOrderService) SubmitRealOrder(userID int64, stockCode, stockName string,
	direction int8, price float64, quantity int, orderType int8) (*model.RealOrder, error) {

	if !middleware.IsRealTradeMode() {
		return nil, errors.New("实盘模式未开启")
	}

	// 加全局锁
	s.orderMutex.Lock()
	defer s.orderMutex.Unlock()

	// 检查账户
	portfolio, err := s.realPortfolioRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("实盘账户不存在，请先创建账户")
	}
	if portfolio.FrozenTrading {
		return nil, errors.New("当日交易已冻结（日亏损超5%）")
	}

	// 风控前置检查
	if err := s.realRiskCheck(portfolio, direction, stockName, price, quantity); err != nil {
		return nil, err
	}

	// 生成委托单号
	orderID := uuid.New().String()

	// 创建订单（pending 状态，等待二次确认）
	order := &model.RealOrder{
		UserID:      userID,
		OrderID:     orderID,
		StockCode:   stockCode,
		StockName:   stockName,
		Direction:   direction,
		OrderType:   orderType,
		Price:       price,
		Quantity:    quantity,
		Status:      "pending",
		IsConfirmed: false,
	}

	if err := s.realOrderRepo.Create(order); err != nil {
		return nil, err
	}

	// 记录操作日志
	s.logTrade(userID, orderID, stockCode, directionStr(direction), price, quantity,
		orderTypeStr(orderType), "pending", "", "")

	return order, nil
}

// ConfirmRealOrder 实盘下单二次确认（pending → confirmed → 提交券商）
func (s *RealOrderService) ConfirmRealOrder(userID int64, orderID string) (*model.RealOrder, error) {
	if !middleware.IsRealTradeMode() {
		return nil, errors.New("实盘模式未开启")
	}

	s.orderMutex.Lock()
	defer s.orderMutex.Unlock()

	// 查订单
	order, err := s.realOrderRepo.FindByOrderID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.UserID != userID {
		return nil, errors.New("无权操作此订单")
	}
	if order.Status != "pending" {
		return nil, errors.New("仅 pending 状态的订单可确认")
	}
	if order.IsConfirmed {
		return nil, errors.New("订单已确认，请勿重复操作")
	}

	// 标记已确认
	order.IsConfirmed = true
	order.Status = "confirmed"

	// 提交券商
	req := &broker.RealOrderRequest{
		UserID:    userID,
		StockCode: order.StockCode,
		StockName: order.StockName,
		Direction: order.Direction,
		Price:     order.Price,
		Quantity:  order.Quantity,
		OrderType: order.OrderType,
	}

	resp, brokerErr := s.brokerImpl.SubmitOrder(req)
	if brokerErr != nil {
		order.Status = "failed"
		order.RejectReason = brokerErr.Error()
		s.realOrderRepo.Update(order)
		s.logTrade(userID, orderID, order.StockCode, directionStr(order.Direction),
			order.Price, order.Quantity, orderTypeStr(order.OrderType), "failed", "", brokerErr.Error())
		return order, brokerErr
	}

	// 更新券商委托号
	order.BrokerOrderID = resp.BrokerOrderID
	if resp.Status == "mock_submitted" {
		order.Status = "confirmed" // Phase 1：mock 状态停留在 confirmed
	} else {
		order.Status = resp.Status // Phase 2：使用券商返回的真实状态
	}

	s.realOrderRepo.Update(order)

	// 记录操作日志
	s.logTrade(userID, orderID, order.StockCode, directionStr(order.Direction),
		order.Price, order.Quantity, orderTypeStr(order.OrderType), order.Status, resp.BrokerOrderID, "")

	// Phase 1 模拟成交：MockBroker 返回后直接标记为 filled
	if resp.MockData {
		order.Status = "filled"
		order.FilledQty = order.Quantity
		order.AvgPrice = order.Price
		if order.AvgPrice == 0 {
			order.AvgPrice = 10.0 // Phase 1 默认价格
		}
		s.realOrderRepo.Update(order)
		s.updateRealPosition(order)
		s.updateRealPortfolioAfterFill(order)
		s.logTrade(userID, orderID, order.StockCode, directionStr(order.Direction),
			order.AvgPrice, order.FilledQty, orderTypeStr(order.OrderType), "filled", resp.BrokerOrderID, "")
	}

	// 推送 WebSocket
	s.pushRealUpdates(userID, order)

	return order, nil
}

// CancelRealOrder 撤销实盘订单
func (s *RealOrderService) CancelRealOrder(userID int64, orderID string) error {
	if !middleware.IsRealTradeMode() {
		return errors.New("实盘模式未开启")
	}

	s.orderMutex.Lock()
	defer s.orderMutex.Unlock()

	order, err := s.realOrderRepo.FindByOrderID(orderID)
	if err != nil {
		return errors.New("订单不存在")
	}
	if order.UserID != userID {
		return errors.New("无权操作此订单")
	}
	if order.Status != "pending" && order.Status != "confirmed" {
		return errors.New("仅可撤销 pending/confirmed 状态的订单")
	}

	// 如果已提交券商，先撤券商
	if order.BrokerOrderID != "" {
		if err := s.brokerImpl.CancelOrder(order.BrokerOrderID); err != nil {
			s.logTrade(userID, orderID, order.StockCode, directionStr(order.Direction),
				order.Price, order.Quantity, orderTypeStr(order.OrderType), "cancel_failed", order.BrokerOrderID, err.Error())
			return fmt.Errorf("券商撤单失败：%v", err)
		}
	}

	order.Status = "cancelled"
	s.realOrderRepo.Update(order)

	s.logTrade(userID, orderID, order.StockCode, directionStr(order.Direction),
		order.Price, order.Quantity, orderTypeStr(order.OrderType), "cancelled", order.BrokerOrderID, "")

	return nil
}

// ListRealOrders 查询实盘订单列表
func (s *RealOrderService) ListRealOrders(userID int64, status string, page, pageSize int) ([]model.RealOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.realOrderRepo.ListByUser(userID, status, page, pageSize)
}

// GetRealOrder 获取实盘订单详情
func (s *RealOrderService) GetRealOrder(userID int64, orderID string) (*model.RealOrder, error) {
	order, err := s.realOrderRepo.FindByOrderID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.UserID != userID {
		return nil, errors.New("无权查看此订单")
	}
	return order, nil
}

// ListRealPositions 查询实盘持仓
func (s *RealOrderService) ListRealPositions(userID int64, stockCode string) ([]model.RealPosition, error) {
	if stockCode != "" {
		pos, err := s.realPositionRepo.FindByUserAndStock(userID, stockCode)
		if err != nil {
			return nil, nil
		}
		return []model.RealPosition{*pos}, nil
	}
	return s.realPositionRepo.ListByUser(userID)
}

// ListRealTradeLogs 查询实盘操作日志
func (s *RealOrderService) ListRealTradeLogs(userID int64, page, pageSize int) ([]model.RealTradeLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.realTradeLogRepo.ListByUser(userID, page, pageSize)
}

// ============ 内部方法 ============

// realRiskCheck 实盘风控前置检查（复用模拟盘规则）
func (s *RealOrderService) realRiskCheck(p *model.RealPortfolio, direction int8, stockName string, price float64, quantity int) error {
	orderValue := price * float64(quantity)

	// 日交易次数上限 50
	if p.DailyTrades >= 50 {
		return errors.New("当日交易次数已达上限（50次）")
	}

	// ST/*ST 禁止买入
	if direction == DirectionBuy {
		if containsST(stockName) {
			return errors.New("禁止买入 ST/*ST 股")
		}
		maxPosition := p.TotalAsset * 0.3
		if orderValue > maxPosition {
			return fmt.Errorf("下单金额 %.2f 超过单股仓位上限（总资产30%%=%.2f）", orderValue, maxPosition)
		}
	}

	// 余额检查（买入）
	if direction == DirectionBuy {
		if p.Balance < orderValue {
			return fmt.Errorf("资金不足：需要 %.2f，可用 %.2f", orderValue, p.Balance)
		}
	}

	return nil
}

// updateRealPosition 更新实盘持仓
func (s *RealOrderService) updateRealPosition(order *model.RealOrder) {
	pos, err := s.realPositionRepo.FindByUserAndStock(order.UserID, order.StockCode)
	if err != nil {
		pos = &model.RealPosition{
			UserID:       order.UserID,
			StockCode:    order.StockCode,
			StockName:    order.StockName,
			Quantity:     0,
			AvailableQty: 0,
			AvgCost:      0,
		}
	}

	totalCost := order.AvgPrice * float64(order.FilledQty)

	if order.Direction == DirectionBuy {
		oldValue := pos.AvgCost * float64(pos.Quantity)
		newQty := pos.Quantity + order.FilledQty
		if newQty > 0 {
			pos.AvgCost = (oldValue + totalCost) / float64(newQty)
		}
		pos.Quantity = newQty
		pos.AvailableQty = newQty
	} else {
		pos.Quantity -= order.FilledQty
		pos.AvailableQty -= order.FilledQty
		if pos.Quantity < 0 {
			pos.Quantity = 0
			pos.AvailableQty = 0
		}
	}

	pos.MarketValue = pos.AvgCost * float64(pos.Quantity)
	pos.StockName = order.StockName

	if pos.ID == 0 {
		s.realPositionRepo.Create(pos)
	} else {
		s.realPositionRepo.Save(pos)
	}
}

// updateRealPortfolioAfterFill 更新实盘账户资金
func (s *RealOrderService) updateRealPortfolioAfterFill(order *model.RealOrder) {
	portfolio, err := s.realPortfolioRepo.FindByUserID(order.UserID)
	if err != nil {
		return
	}

	amount := order.AvgPrice * float64(order.FilledQty)

	if order.Direction == DirectionBuy {
		portfolio.Balance -= amount
	} else {
		portfolio.Balance += amount
	}

	// 更新总资产
	positions, _ := s.realPositionRepo.ListByUser(portfolio.UserID)
	holdingsValue := 0.0
	for _, pos := range positions {
		holdingsValue += pos.MarketValue
	}
	portfolio.TotalAsset = portfolio.Balance + holdingsValue

	// 日亏损检测
	if portfolio.TotalAsset < 1000000*(1-0.05) {
		portfolio.DailyLoss = 1000000 - portfolio.TotalAsset
		if portfolio.DailyLoss >= 1000000*0.05 {
			portfolio.FrozenTrading = true
		}
	}

	portfolio.DailyTrades++
	s.realPortfolioRepo.Save(portfolio)
}

// pushRealUpdates 推送实盘 WebSocket 更新
func (s *RealOrderService) pushRealUpdates(userID int64, order *model.RealOrder) {
	if hub := ws.GetHub(); hub != nil {
		hub.BroadcastToUser(userID, ws.EventOrderUpdate, order)
		positions, _ := s.realPositionRepo.ListByUser(userID)
		hub.BroadcastToUser(userID, ws.EventPositionUpdate, positions)
	}
}

// logTrade 记录操作日志
func (s *RealOrderService) logTrade(userID int64, orderID, stockCode, direction string,
	price float64, quantity int, orderType, status, brokerOrderID, errMsg string) {

	log := &model.RealTradeLog{
		UserID:        userID,
		OrderID:       orderID,
		StockCode:     stockCode,
		Direction:     direction,
		Price:         price,
		Quantity:      quantity,
		OrderType:     orderType,
		Status:        status,
		BrokerOrderID: brokerOrderID,
		ErrorMessage:  errMsg,
	}
	s.realTradeLogRepo.Create(log)
}

// directionStr 方向转字符串
func directionStr(d int8) string {
	if d == DirectionBuy {
		return "buy"
	}
	return "sell"
}

// orderTypeStr 委托类型转字符串
func orderTypeStr(t int8) string {
	if t == 2 {
		return "limit"
	}
	return "market"
}
