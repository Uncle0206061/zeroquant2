package service

import (
	"errors"

	"github.com/Uncle0206061/zeroquant2/backend/internal/model"
	"github.com/Uncle0206061/zeroquant2/backend/internal/repository"
	"github.com/Uncle0206061/zeroquant2/backend/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserExists   = errors.New("用户名已存在")
	ErrInvalidCreds = errors.New("用户名或密码错误")
	ErrUserDisabled = errors.New("账号已被禁用")
)

// UpdateProfileReq 更新画像请求
type UpdateProfileReq struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Bio      string `json:"bio"`
	Phone    string `json:"phone"`
}

type AuthService struct {
	repo *repository.UserRepository
}

func NewAuthService(repo *repository.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Register 注册
// 1. 检查用户名是否已存在
// 2. bcrypt 加密密码
// 3. 创建用户
// 4. 生成 JWT Token
// 返回 user + token
func (s *AuthService) Register(username, password, email, role string) (*model.User, string, error) {
	_, err := s.repo.FindByUsername(username)
	if err == nil {
		return nil, "", ErrUserExists
	}
	if err != gorm.ErrRecordNotFound {
		return nil, "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return nil, "", err
	}

	if role == "" {
		role = "user"
	}

	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Nickname:     username,
		Role:         role,
		Status:       1,
	}
	if err := s.repo.Create(user); err != nil {
		return nil, "", err
	}

	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login 登录
// 1. 查找用户
// 2. bcrypt 验证密码
// 3. 检查账号状态
// 4. 生成 JWT Token
func (s *AuthService) Login(username, password string) (string, int64, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return "", 0, ErrInvalidCreds
	}
	if user.Status != 1 {
		return "", 0, ErrUserDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", 0, ErrInvalidCreds
	}

	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return "", 0, err
	}

	return token, user.ID, nil
}

// GetProfile 获取用户完整信息（基本信息 + 画像）
func (s *AuthService) GetProfile(userID int64) (*model.User, *model.UserProfile, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, nil, err
	}
	profile, err := s.repo.FindProfile(userID)
	if err != nil {
		return nil, nil, err
	}
	return user, profile, nil
}

// UpdateProfile 更新用户画像
func (s *AuthService) UpdateProfile(userID int64, req *UpdateProfileReq) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if err := s.repo.Update(user); err != nil {
		return err
	}

	// FindProfile 已自动创建空白画像
	profile, err := s.repo.FindProfile(userID)
	if err != nil {
		return err
	}
	if req.Avatar != "" {
		profile.Avatar = req.Avatar
	}
	if req.Bio != "" {
		profile.Bio = req.Bio
	}
	if req.Phone != "" {
		profile.Phone = req.Phone
	}
	return s.repo.UpdateProfile(profile)
}

// GetAllUsers 获取所有用户（admin）
func (s *AuthService) GetAllUsers() ([]model.User, error) {
	return s.repo.GetAllUsers()
}
