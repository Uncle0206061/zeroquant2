// Package cache 提供 Redis 缓存服务
// 盘口数据缓存、用户Session缓存、策略结果缓存
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Uncle0206061/zeroquant2/backend/internal/config"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/logger"
)

const (
	// 盘口数据缓存 TTL（秒）
	QuoteCacheTTL = 3 * time.Second

	// 用户Session缓存 TTL
	UserCacheTTL = 30 * time.Minute

	// 策略结果缓存 TTL
	StrategyCacheTTL = 5 * time.Minute
)

// QuoteData 盘口缓存数据
type QuoteData struct {
	StockCode  string  `json:"stock_code"`
	StockName  string  `json:"stock_name"`
	Bid1Price  float64 `json:"bid1_price"`
	Bid1Qty    int     `json:"bid1_qty"`
	Ask1Price  float64 `json:"ask1_price"`
	Ask1Qty    int     `json:"ask1_qty"`
	PrevClose  float64 `json:"prev_close"`
	ChangeRate float64 `json:"change_rate"`
	UpdatedAt  int64   `json:"updated_at"` // 更新时间戳
}

// quoteKey 生成盘口缓存 key
func quoteKey(stockCode string) string {
	return fmt.Sprintf("quote:%s", stockCode)
}

// userKey 生成用户缓存 key
func userKey(userID int64) string {
	return fmt.Sprintf("user:%d", userID)
}

// strategyKey 生成策略缓存 key
func strategyKey(strategyID int64) string {
	return fmt.Sprintf("strategy:%d", strategyID)
}

// GetQuote 获取盘口缓存
func GetQuote(ctx context.Context, stockCode string) (*QuoteData, error) {
	rdb := config.GetRedis()
	if rdb == nil {
		return nil, fmt.Errorf("redis not initialized")
	}

	data, err := rdb.Get(ctx, quoteKey(stockCode)).Bytes()
	if err != nil {
		return nil, err // redis.Nil 或其他错误
	}

	var quote QuoteData
	if err := json.Unmarshal(data, &quote); err != nil {
		return nil, err
	}
	return &quote, nil
}

// SetQuote 写入盘口缓存（TTL 3秒）
func SetQuote(ctx context.Context, quote *QuoteData) error {
	rdb := config.GetRedis()
	if rdb == nil {
		return fmt.Errorf("redis not initialized")
	}

	data, err := json.Marshal(quote)
	if err != nil {
		return err
	}

	return rdb.Set(ctx, quoteKey(quote.StockCode), data, QuoteCacheTTL).Err()
}

// GetUserCache 获取用户缓存
func GetUserCache(ctx context.Context, userID int64) (map[string]interface{}, error) {
	rdb := config.GetRedis()
	if rdb == nil {
		return nil, fmt.Errorf("redis not initialized")
	}

	data, err := rdb.Get(ctx, userKey(userID)).Bytes()
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetUserCache 写入用户缓存
func SetUserCache(ctx context.Context, userID int64, data map[string]interface{}) error {
	rdb := config.GetRedis()
	if rdb == nil {
		return fmt.Errorf("redis not initialized")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return rdb.Set(ctx, userKey(userID), jsonData, UserCacheTTL).Err()
}

// DeleteUserCache 清除用户缓存
func DeleteUserCache(ctx context.Context, userID int64) error {
	rdb := config.GetRedis()
	if rdb == nil {
		return nil
	}
	return rdb.Del(ctx, userKey(userID)).Err()
}

// GetStrategyCache 获取策略缓存
func GetStrategyCache(ctx context.Context, strategyID int64) ([]byte, error) {
	rdb := config.GetRedis()
	if rdb == nil {
		return nil, fmt.Errorf("redis not initialized")
	}

	return rdb.Get(ctx, strategyKey(strategyID)).Bytes()
}

// SetStrategyCache 写入策略缓存
func SetStrategyCache(ctx context.Context, strategyID int64, data []byte) error {
	rdb := config.GetRedis()
	if rdb == nil {
		return fmt.Errorf("redis not initialized")
	}

	return rdb.Set(ctx, strategyKey(strategyID), data, StrategyCacheTTL).Err()
}

// Ping 检查 Redis 连接
func Ping(ctx context.Context) error {
	rdb := config.GetRedis()
	if rdb == nil {
		return fmt.Errorf("redis not initialized")
	}
	return rdb.Ping(ctx).Err()
}

// init 日志提示
func init() {
	logger.Info("Cache module loaded (quote TTL=3s, user TTL=30m, strategy TTL=5m)")
}
