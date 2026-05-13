// Package router 提供路由注册和管理
package router

import (
	"github.com/Uncle0206061/zeroquant2/backend/internal/broker"
	"github.com/Uncle0206061/zeroquant2/backend/internal/config"
	"github.com/Uncle0206061/zeroquant2/backend/internal/handler"
	"github.com/Uncle0206061/zeroquant2/backend/internal/middleware"
	"github.com/Uncle0206061/zeroquant2/backend/internal/repository"
	"github.com/Uncle0206061/zeroquant2/backend/internal/service"
	"github.com/Uncle0206061/zeroquant2/backend/internal/websocket"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Setup 注册所有路由
func Setup(cfg *config.Config, db *gorm.DB) *gin.Engine {
	// 创建 Gin 引擎
	r := gin.Default()

	// 注册中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// 健康检查路由（无需认证）
	r.GET("/api/v1/health", handler.HealthCheck)
	r.GET("/api/v1/ping", handler.Ping)

	// WebSocket 路由
	r.GET("/api/v1/ws", websocket.HandleWS)

	// WebSocket 推送回调（供 Python 数据服务调用，无需 JWT）
	r.POST("/api/v1/ws/push", handler.WSPushHandler)
	r.GET("/api/v1/ws/stats", handler.WSStatsHandler)

	// ============ 依赖注入 ============
	// Repository → Service → Handler 链
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo)

	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(authSvc)
	adminHandler := handler.NewAdminHandler(authSvc)

	// ============ 公开路由（无需认证）============
	r.POST("/api/v1/auth/register", authHandler.Register)
	r.POST("/api/v1/auth/login", authHandler.Login)

	// ============ API 路由组（需 JWT 认证）============
	apiV1 := r.Group("/api/v1")
	apiV1.Use(middleware.JWTAuth())
	{
		// auth
		apiV1.GET("/auth/me", authHandler.Me)

		// user 画像
		apiV1.GET("/user/profile", userHandler.GetProfile)
		apiV1.PUT("/user/profile", userHandler.UpdateProfile)

		// strategy 策略管理
		strategyRepo := repository.NewStrategyRepository(db)
		strategySvc := service.NewStrategyService(strategyRepo, cfg.DataServiceURL)
		strategyHandler := handler.NewStrategyHandler(strategySvc)
		apiV1.GET("/strategy/list", strategyHandler.ListStrategy)
		apiV1.POST("/strategy/create", strategyHandler.CreateStrategy)
		apiV1.GET("/strategy/:id", strategyHandler.GetStrategy)
		apiV1.PUT("/strategy/:id", strategyHandler.UpdateStrategy)
		apiV1.DELETE("/strategy/:id", strategyHandler.DeleteStrategy)
		apiV1.POST("/strategy/:id/submit", strategyHandler.SubmitStrategy)
		apiV1.GET("/strategy/:id/backtests", strategyHandler.GetBacktests)

		// trade 交易（模拟账户/订单/持仓/撮合引擎）
		orderRepo := repository.NewOrderRepository(db)
		positionRepo := repository.NewPositionRepository(db)
		portfolioRepo := repository.NewPortfolioRepository(db)
		orderSvc := service.NewOrderService(orderRepo, positionRepo, portfolioRepo, cfg.DataServiceURL)
		tradeHandler := handler.NewTradeHandler(orderSvc)
		tradeHandler.RegisterRoutes(apiV1)

		// real trade 实盘交易（受 TradeModeMiddleware 控制）
		realOrderRepo := repository.NewRealOrderRepository(db)
		realPositionRepo := repository.NewRealPositionRepository(db)
		realPortfolioRepo := repository.NewRealPortfolioRepository(db)
		realTradeLogRepo := repository.NewRealTradeLogRepository(db)
		mockBroker := broker.NewMockBroker()
		realOrderSvc := service.NewRealOrderService(realOrderRepo, realPositionRepo, realPortfolioRepo, realTradeLogRepo, mockBroker)
		realTradeHandler := handler.NewRealTradeHandler(realOrderSvc)
		realTradeHandler.RegisterRoutes(apiV1)

		// admin 路由
		admin := apiV1.Group("/admin")
		admin.Use(handler.RequireAdmin())
		{
			admin.GET("/users", adminHandler.GetAllUsers)
		}
	}

	return r
}