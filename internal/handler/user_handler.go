package handler

import (
	"github.com/Uncle0206061/zeroquant2/backend/internal/service"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户画像 Handler
type UserHandler struct {
	svc *service.AuthService
}

func NewUserHandler(svc *service.AuthService) *UserHandler {
	return &UserHandler{svc: svc}
}

// GetProfile 获取用户画像
// GET /api/v1/user/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, profile, err := h.svc.GetProfile(userID.(int64))
	if err != nil {
		response.ServerError(c, "获取失败")
		return
	}

	response.Success(c, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"nickname": user.Nickname,
		"role":     user.Role,
		"avatar":   profile.Avatar,
		"bio":      profile.Bio,
		"phone":    profile.Phone,
	})
}

// UpdateProfile 更新用户画像
// PUT /api/v1/user/profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req service.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "参数错误")
		return
	}

	if err := h.svc.UpdateProfile(userID.(int64), &req); err != nil {
		response.ServerError(c, "更新失败")
		return
	}

	response.Success(c, nil)
}
