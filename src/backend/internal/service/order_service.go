// Package service 提供业务逻辑层
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/Uncle0206061/zeroquant2/backend/internal/model"
	"github.com/Uncle0206061/zeroquant2/backend/internal/repository"
	ws "github.com/Uncle0206061/zeroquant2/backend/internal/websocket"
)

// 撮合方向常量
const (
	DirectionBuy  = 1
	DirectionSell = 2
)

// OrderService 订单与撮合服务
type OrderService struct {
	orderRepo     *repository.OrderRepository
	positionRepo *repository.PositionRepository
	portfolioRepo *repository.PortfolioRepository
	quoteURL      string // Python数据服务盘口地址
	orderMutex   sync.Mutex
}

// NewOrderService 创建 OrderService
func NewOrderService(
	orderRepo *repository.OrderRepository,
	positionRepo *repository.PositionRepository,
	portfolioRepo *repository.PortfolioRepository,
	quoteURL string,
) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		positionRepo:  positionRepo,
		portfolioRepo: portfolioRepo,
		quoteURL:      quoteURL,
	}
}

// QuoteResponse Python数据服务盘口响应
type QuoteResponse struct {
	StockCode  string  `json:"stock_code"`
	StockName  string  `json:"stock_name"`
	Bid1Price  float64 `json:"bid1_price"`  // 买一价
	Bid1Qty    int     `json:"bid1_qty"`    // 买一量
	Ask1Price  float64 `json:"ask1_price"`  // 卖一价
	Ask1Qty    int     `json:"ask1_qty"`    // 卖一量
	PrevClose  float64 `json:"prev_close"`  // 昨收价
	ChangeRate float64 `json:"change_rate"` // 涨跌幅%
}

// CreateSimAccount 创建模拟账户（初始资金100万）
func (s *OrderService) CreateSimAccount(userID int64) (*model.Portfolio, error) {
	// 查重：每个用户只能有一个模拟账户
	existing, err := s.portfolioRepo.FindByUserID(userID)
	if err == nil && existing != nil {
		return existing, nil // 已存在则返回现有账户
	}

	p := &model.Portfolio{
		UserID:      userID,
		Balance:     1000000,
		Frozen:      0,
		TotalAsset:  1000000,
		DailyLoss:   0,
		DailyTrades: 0,
	}

	if err := s.portfolioRepo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

// SubmitOrder 下单（撮合引擎入口，持Mutex锁）
func (s *OrderService) SubmitOrder(userID int64, taskID, stockCode, stockName string,
	direction int8, price float64, quantity int) (*model.Order, error) {

	// 1. 加全局锁，禁止并发下单
	s.orderMutex.Lock()
	defer s.orderMutex.Unlock()

	// 2. 检查交易冻结
	portfolio, err := s.portfolioRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("模拟账户不存在，请先创建账户")
	}

	if portfolio.FrozenTrading {
		return nil, errors.New("当日交易已冻结（日亏损超5%）")
	}

	// 3. 生成委托单号
	orderID := uuid.New().String()

	// 4. 风控前置检查
	order := &model.Order{
		TaskID:    taskID,
		UserID:    userID,
		OrderID:   orderID,
		StockCode: stockCode,
		StockName: stockName,
		Direction: direction,
		OrderType: 1, // 市价委托
		Price:     price,
		Quantity:  quantity,
		Status:    "pending",
	}

	// 风控检查（若失败直接返回，不写订单）
	if err := s.riskCheck(portfolio, direction, stockName, price, quantity); err != nil {
		order.Status = "rejected"
		order.RejectReason = err.Error()
		s.orderRepo.Create(order)
		return order, err
	}

	// 5. 写订单（pending）
	if err := s.orderRepo.Create(order); err != nil {
		return nil, err
	}

	// 6. 撮合
	filled, avgPrice, matchErr := s.match(order, portfolio)

	if filled > 0 {
		// 6a. 成交
		order.Status = "filled"
		order.FilledQty = filled
		order.AvgPrice = avgPrice
		s.orderRepo.Update(order)

		// 更新持仓
		s.updatePosition(order, portfolio)

		// 更新账户
		s.updatePortfolioAfterFill(portfolio, order)

		// 推送 WebSocket
		s.pushUpdates(userID, order)
	} else {
		// 6b. 未成交 → pending，超时任务另处理
		if matchErr != nil {
			order.Status = "rejected"
			order.RejectReason = matchErr.Error()
		}
		s.orderRepo.Update(order)
	}

	// 7. 日交易次数+1
	portfolio.DailyTrades++
	s.portfolioRepo.Save(portfolio)

	return order, nil
}

// riskCheck 风控前置检查
// direction: 1=buy, 2=sell
func (s *OrderService) riskCheck(p *model.Portfolio, direction int8, stockName string, price float64, quantity int) error {
	orderValue := price * float64(quantity)

	// 日交易次数上限 50
	if p.DailyTrades >= 50 {
		return errors.New("当日交易次数已达上限（50次）")
	}

	// ST/*ST禁止买入
	if direction == DirectionBuy {
		if containsST(stockName) {
			return errors.New("禁止买入 ST/*ST 股")
		}

		// 单股仓位上限 30%
		// 粗估：当前资金*30% vs 订单金额
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

// containsST 检查股票名称是否含 ST/*ST
func containsST(name string) bool {
	return contains(name, "*ST") || contains(name, "ST ") || contains(name, " ST")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// match 撮合：查盘口 → 判断成交
// 返回：成交数量、成交均价、错误信息
func (s *OrderService) match(order *model.Order, portfolio *model.Portfolio) (int, float64, error) {
	// 查询 Python 数据服务的盘口
	quote, err := s.fetchQuote(order.StockCode)
	if err != nil {
		// 数据服务不可用 → 订单维持 pending（不成交）
		return 0, 0, nil
	}

	// 获取成交价（若未指定价格，按盘口档位撮合）
	execPrice := order.Price
	if execPrice == 0 {
		// 市价：买入→卖一价，卖出→买一价
		if order.Direction == DirectionBuy {
			execPrice = quote.Ask1Price
		} else {
			execPrice = quote.Bid1Price
		}
	}

	// 判断是否成交
	var canMatch bool
	if order.Direction == DirectionBuy {
		// 买入：卖一价 ≤ 委托价 → 成交
		canMatch = quote.Ask1Price > 0 && quote.Ask1Price <= execPrice
	} else {
		// 卖出：买一价 ≥ 委托价 → 成交
		canMatch = quote.Bid1Price > 0 && quote.Bid1Price >= execPrice
	}

	if !canMatch {
		return 0, 0, nil // 不成交
	}

	// 更新订单股票名称（以盘口为准）
	if order.StockName == "" {
		order.StockName = quote.StockName
	}

	// 撮合逻辑：买入→卖一价成交（取Ask1Qty），卖出→买一价成交（取Bid1Qty）
	var filledQty int
	if order.Direction == DirectionBuy {
		filledQty = minInt(order.Quantity, quote.Ask1Qty)
	} else {
		filledQty = minInt(order.Quantity, quote.Bid1Qty)
	}

	if filledQty <= 0 {
		// 无对手盘，无法成交
		return 0, 0, nil
	}

	if filledQty <= 0 {
		return 0, 0, nil
	}

	return filledQty, execPrice, nil
}

// fetchQuote 查询 Python 数据服务的盘口
func (s *OrderService) fetchQuote(stockCode string) (*QuoteResponse, error) {
	url := s.quoteURL + "/data/v1/quote/" + stockCode
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("盘口服务返回状态码 %d", resp.StatusCode)
	}

	var quote QuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
		return nil, err
	}

	return &quote, nil
}

// updatePosition 更新持仓（成交后）
func (s *OrderService) updatePosition(order *model.Order, portfolio *model.Portfolio) {
	pos, err := s.positionRepo.FindByUserAndStock(order.UserID, order.StockCode)
	if err != nil {
		// 持仓不存在，新建
		pos = &model.Position{
			UserID:       order.UserID,
			StockCode:    order.StockCode,
			StockName:    order.StockName,
			Quantity:     0,
			AvailableQty: 0,
			AvgCost:      0,
		}
	}

	avgPrice := order.AvgPrice
	totalCost := avgPrice * float64(order.FilledQty)

	if order.Direction == DirectionBuy {
		// 买入：加仓（计算新的成本价）
		oldValue := pos.AvgCost * float64(pos.Quantity)
		newQty := pos.Quantity + order.FilledQty
		if newQty > 0 {
			pos.AvgCost = (oldValue + totalCost) / float64(newQty)
		}
		pos.Quantity = newQty
		pos.AvailableQty = newQty
	} else {
		// 卖出：减仓
		pos.Quantity -= order.FilledQty
		pos.AvailableQty -= order.FilledQty
		if pos.Quantity < 0 {
			pos.Quantity = 0
			pos.AvailableQty = 0
		}
	}

	// 更新市值和盈亏（简化版：用成本价估算）
	pos.MarketValue = pos.AvgCost * float64(pos.Quantity)
	pos.ProfitLoss = 0
	pos.ProfitRate = 0
	pos.StockName = order.StockName

	if pos.ID == 0 {
		s.positionRepo.Create(pos)
	} else {
		s.positionRepo.Save(pos)
	}
}

// updatePortfolioAfterFill 更新账户资金（成交后）
func (s *OrderService) updatePortfolioAfterFill(p *model.Portfolio, order *model.Order) {
	amount := order.AvgPrice * float64(order.FilledQty)

	if order.Direction == DirectionBuy {
		p.Balance -= amount
	} else {
		p.Balance += amount
	}

	// 更新总资产（简化：资金+持仓市值）
	positions, _ := s.positionRepo.ListByUser(p.UserID)
	holdingsValue := 0.0
	for _, pos := range positions {
		holdingsValue += pos.MarketValue
	}
	p.TotalAsset = p.Balance + holdingsValue

	// 日亏损检测（懒重置：首次下单时检查是否新交易日）
	if s.isNewDay(p) {
		p.DailyLoss = 0
		p.DailyTrades = 0
		p.FrozenTrading = false
	}

	// 日亏损 ≥5% 则冻结
	if p.TotalAsset < 1000000*(1-0.05) {
		p.DailyLoss = 1000000 - p.TotalAsset
		if p.DailyLoss >= 1000000*0.05 {
			p.FrozenTrading = true
		}
	}

	s.portfolioRepo.Save(p)
}

// isNewDay 检查是否是新股交易日（用于重置日状态）
func (s *OrderService) isNewDay(p *model.Portfolio) bool {
	now := time.Now()
	today := now.Format("2006-01-02")

	// 从数据库获取最新记录判断（简化：用 UpdatedAt）
	if p.UpdatedAt.IsZero() {
		return true
	}

	lastDay := p.UpdatedAt.Format("2006-01-02")
	return today != lastDay
}

// pushUpdates 推送 WebSocket 更新
func (s *OrderService) pushUpdates(userID int64, order *model.Order) {
	// 获取全局 Hub 实例（通过 ws 包）
	if hub := ws.GetHub(); hub != nil {
		// 推送订单更新
		hub.BroadcastToUser(userID, ws.EventOrderUpdate, order)
		// 推送持仓更新
		positions, _ := s.positionRepo.ListByUser(userID)
		hub.BroadcastToUser(userID, ws.EventPositionUpdate, positions)
	}
}

// CancelOrder 撤单（仅 pending 可撤）
func (s *OrderService) CancelOrder(userID int64, orderID string) error {
	s.orderMutex.Lock()
	defer s.orderMutex.Unlock()

	order, err := s.orderRepo.FindByOrderID(orderID)
	if err != nil {
		return errors.New("订单不存在")
	}

	if order.UserID != userID {
		return errors.New("无权操作此订单")
	}

	if order.Status != "pending" {
		return errors.New("仅可撤销 pending 状态的订单")
	}

	order.Status = "cancelled"
	return s.orderRepo.Update(order)
}

// ListOrders 查询订单列表
func (s *OrderService) ListOrders(userID int64, status string, page, pageSize int) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.orderRepo.ListByUser(userID, status, page, pageSize)
}

// GetOrder 获取订单详情
func (s *OrderService) GetOrder(userID int64, orderID string) (*model.Order, error) {
	order, err := s.orderRepo.FindByOrderID(orderID)
	if err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.UserID != userID {
		return nil, errors.New("无权查看此订单")
	}
	return order, nil
}

// ListPositions 查询持仓列表
func (s *OrderService) ListPositions(userID int64, stockCode string) ([]model.Position, error) {
	if stockCode != "" {
		pos, err := s.positionRepo.FindByUserAndStock(userID, stockCode)
		if err != nil {
			return nil, nil
		}
		return []model.Position{*pos}, nil
	}
	return s.positionRepo.ListByUser(userID)
}

// GetAccount 获取模拟账户
func (s *OrderService) GetAccount(userID int64) (*model.Portfolio, error) {
	return s.portfolioRepo.FindByUserID(userID)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// avoid unused
