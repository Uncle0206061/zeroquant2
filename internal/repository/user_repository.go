package repository

import (
	"github.com/Uncle0206061/zeroquant2/backend/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// FindByUsername 按用户名查找
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID 按 ID 查找
func (r *UserRepository) FindByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// FindProfile 查找用户画像（不存在则自动创建空记录）
func (r *UserRepository) FindProfile(userID int64) (*model.UserProfile, error) {
	var profile model.UserProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	if err == gorm.ErrRecordNotFound {
		profile = model.UserProfile{UserID: userID}
		if createErr := r.db.Create(&profile).Error; createErr != nil {
			return nil, createErr
		}
		return &profile, nil
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// UpdateProfile 更新画像
func (r *UserRepository) UpdateProfile(profile *model.UserProfile) error {
	return r.db.Save(profile).Error
}

// GetAllUsers 获取所有用户（admin 用）
func (r *UserRepository) GetAllUsers() ([]model.User, error) {
	var users []model.User
	err := r.db.Select("id, username, email, role, status, created_at").Find(&users).Error
	return users, err
}
