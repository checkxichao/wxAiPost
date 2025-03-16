package repositories

import (
	"gorm.io/gorm"
	"publicPost/src/models"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	SetUserPower(username string) error
	FindUserByUsername(username string) (*models.User, error)
	GetAllUser() ([]*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	CheckUserPower(username string) bool
	UpdateUserInfo(username string, pwd string, power int) error
	DeleteUser(id int) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}
func (r *userRepository) UpdateUserInfo(username string, pwd string, power int) error {

	return r.db.Table("users").Where("username", username).Update("password", pwd).Update("power", power).Error
}

func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}
func (r *userRepository) DeleteUser(id int) error {
	return r.db.Table("users").Where("id", id).Update("show", 0).Error
}
func (r *userRepository) CheckUserPower(username string) bool {
	user, err := r.GetUserByUsername(username)
	if err != nil {
		return false
	}
	if user.Power != 2 {
		return false
	}
	return true
}
func (r *userRepository) GetAllUser() ([]*models.User, error) {
	result := make([]*models.User, 0)
	if err := r.db.Where("show", 1).Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}
func (r *userRepository) GetUserByUsername(username string) (*models.User, error) {
	result := &models.User{}
	if err := r.db.Where("username = ?", username).Find(result).Error; err != nil {
		return nil, err
	}
	return result, nil

}
func (r *userRepository) SetUserPower(username string) error {
	updates := map[string]interface{}{
		"power": 1,
	}
	return r.db.Where("username", username).Updates(updates).Error
}

func (r *userRepository) FindUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
