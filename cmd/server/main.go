// Package main is the entry point for ZeroQuant 2.0 backend server.
// 核心功能：行情展示、自选股管理、模拟交易、简单回测
package main

import (
	"os"

	"github.com/Uncle0206061/zeroquant2/backend/internal/config"
	"github.com/Uncle0206061/zeroquant2/backend/internal/model"
	"github.com/Uncle0206061/zeroquant2/backend/internal/router"
	"github.com/Uncle0206061/zeroquant2/backend/internal/websocket"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/logger"
)

func main() {
	// 初始化日志（默认级别）
	logger.Init("info", "")
	logger.Info("Starting ZeroQuant 2.0 Backend Server...")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config: %v", err)
	}

	// 初始化数据库连接
	if err := config.InitDB(cfg); err != nil {
		logger.Error("Failed to connect database: %v", err)
		os.Exit(1)
	}
	defer config.CloseDB()

	// 自动迁移表结构（模拟盘+实盘）
	if err := config.DB.AutoMigrate(
		&model.User{},
		&model.UserProfile{},
		&model.Strategy{},
		&model.StrategyRule{},
		&model.Backtest{},
		&model.Portfolio{},
		&model.Order{},
		&model.Position{},
		&model.RealOrder{},
		&model.RealPosition{},
		&model.RealPortfolio{},
		&model.RealTradeLog{},
	); err != nil {
		logger.Error("Failed to auto migrate: %v", err)
		os.Exit(1)
	}
	logger.Info("Database auto migrated")

	// 初始化 Redis 连接
	if err := config.InitRedis(cfg); err != nil {
		logger.Error("Failed to connect Redis: %v", err)
		os.Exit(1)
	}
	defer config.CloseRedis()

	// 初始化 WebSocket Hub
	websocket.InitHub()

	// 注册路由并启动服务
	r := router.Setup(cfg, config.DB)
	addr := ":" + cfg.ServerPort
	logger.Info("Server listening on " + addr)
	if err := r.Run(addr); err != nil {
		logger.Error("Server failed: %v", err)
		os.Exit(1)
	}
}