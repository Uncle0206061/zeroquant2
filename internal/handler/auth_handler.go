package handler

import (
	"github.com/Uncle0206061/zeroquant2/backend/internal/service"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// RegisterReq 注册请求
type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"omitempty,email"`
	Role     string `json:"role"`
}

// LoginReq 登录请求
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register 注册接口
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "参数错误: "+err.Error())
		return
	}

	user, token, err := h.svc.Register(req.Username, req.Password, req.Email, req.Role)
	if err != nil {
		if err == service.ErrUserExists {
			response.InvalidParam(c, "用户名已存在")
			return
		}
		response.ServerError(c, "注册失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"user_id": user.ID,
		"token":   token,
	})
}

// Login 登录接口
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidParam(c, "参数错误: "+err.Error())
		return
	}

	token, userID, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		if err == service.ErrInvalidCreds {
			response.Unauthorized(c, "用户名或密码错误")
			return
		}
		if err == service.ErrUserDisabled {
			response.Unauthorized(c, "账号已被禁用")
			return
		}
		response.ServerError(c, "登录失败")
		return
	}

	response.Success(c, gin.H{
		"token":   token,
		"user_id": userID,
	})
}

// Me 当前用户信息
// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, profile, err := h.svc.GetProfile(userID.(int64))
	if err != nil {
		response.ServerError(c, "获取用户信息失败")
		return
	}

	response.Success(c, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"nickname": user.Nickname,
		"role":     user.Role,
		"status":   user.Status,
		"avatar":   profile.Avatar,
		"bio":      profile.Bio,
		"phone":    profile.Phone,
	})
}
