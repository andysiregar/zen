package repositories

import (
	"time"
	"shared/pkg/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id string) error
	List(limit, offset int) ([]*models.User, error)
	Count() (int64, error)
	IncrementLoginAttempts(userID string) error
	ResetLoginAttempts(userID string) error
	UpdateLastLogin(userID string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.User{}).Error
}

func (r *userRepository) List(limit, offset int) ([]*models.User, error) {
	var users []*models.User
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

func (r *userRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

func (r *userRepository) IncrementLoginAttempts(userID string) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("login_attempts", gorm.Expr("login_attempts + ?", 1)).Error
}

func (r *userRepository) ResetLoginAttempts(userID string) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"login_attempts": 0,
			"locked_at":      nil,
		}).Error
}

func (r *userRepository) UpdateLastLogin(userID string) error {
	now := time.Now()
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at": &now,
			// TODO: Add IP tracking if needed
			// "last_login_ip": ip,
		}).Error
}