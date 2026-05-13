// Package main is the entry point for ZeroQuant 2.0 backend server.
// 核心功能：行情展示、自选股管理、模拟交易、实盘交易、简单回测
//
// @title ZeroQuant 2.0 Backend API
// @version 1.0.0
// @description A股全自动交易系统后端API - 用户注册登录、策略管理、模拟交易、实盘交易、WebSocket实时推送
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Uncle0206061/zeroquant2/backend/internal/config"
	"github.com/Uncle0206061/zeroquant2/backend/internal/model"
	"github.com/Uncle0206061/zeroquant2/backend/internal/router"
	"github.com/Uncle0206061/zeroquant2/backend/internal/websocket"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/logger"
)

func main() {
	// 初始化日志（默认级别，后续配置加载后可调整）
	logger.Init("info", "./logs/app.log")
	logger.Info("Starting ZeroQuant 2.0 Backend Server...")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config: %v", err)
	}

	// 重新初始化日志（使用配置文件中的级别和路径）
	logger.Init(cfg.LogLevel, cfg.LogPath)

	// 初始化数据库连接（含连接池优化）
	if err := config.InitDB(cfg); err != nil {
		logger.Fatal("Failed to connect database: %v", err)
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
		&model.Alert{},
		&model.RealOrder{},
		&model.RealPosition{},
		&model.RealPortfolio{},
		&model.RealTradeLog{},
	); err != nil {
		logger.Fatal("Failed to auto migrate: %v", err)
	}
	logger.Info("Database auto migrated (13 tables)")

	// 初始化 Redis 连接（含连接池优化）
	if err := config.InitRedis(cfg); err != nil {
		logger.Fatal("Failed to connect Redis: %v", err)
	}
	defer config.CloseRedis()

	// 初始化 WebSocket Hub（最大100连接）
	websocket.InitHub()

	// 注册路由
	r := router.Setup(cfg, config.DB)

	// ============ 优雅关闭 ============
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	// 启动服务（非阻塞）
	go func() {
		logger.Info("Server listening on %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("Received signal: %v, shutting down...", sig)

	// 优雅关闭（5秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced shutdown: %v", err)
	}

	// 关闭 WebSocket Hub
	if hub := websocket.GetHub(); hub != nil {
		logger.Info("WebSocket connections: %d", hub.ClientCount())
	}

	logger.Info("Server exited gracefully")
}
