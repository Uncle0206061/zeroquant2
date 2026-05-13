package handler

import (
	"github.com/Uncle0206061/zeroquant2/backend/internal/service"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	svc *service.AuthService
}

func NewAdminHandler(svc *service.AuthService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// RequireAdmin admin 权限检查中间件
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			response.Forbidden(c, "需要 admin 权限")
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetAllUsers 获取所有用户（admin 专用）
// GET /api/v1/admin/users
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	users, err := h.svc.GetAllUsers()
	if err != nil {
		response.ServerError(c, "获取失败")
		return
	}
	response.Success(c, users)
}
