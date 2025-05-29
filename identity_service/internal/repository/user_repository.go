package repository

import (
	"identity_service/internal/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByEmail ищет пользователя по email в базе данных
func (r *UserRepository) FindByEmail(email string) (*models.Users, error) {
	var user models.Users
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// Create создает нового пользователя в базе данных
func (r UserRepository) Create(user *models.Users) error {
	return r.db.Create(user).Error
}
