package user

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetByEmail(email string) (*User, error)
	GetByUID(uid string) (*User, error)
	CreateUser(user *User) error
	GetAllUsers() ([]*User, error)
	UpdateRole(uid string, role string) error
	UpdateProfile(uid string, email string) (*User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUID(uid string) (*User, error) {
	var user User
	err := r.db.Where("firebase_uid = ?", uid).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CreateUser(user *User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetAllUsers() ([]*User, error) {
	var users []*User
	if err := r.db.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	return users, nil
}

func (r *userRepository) UpdateRole(uid string, role string) error {
	result := r.db.Model(&User{}).Where("firebase_uid = ?", uid).Updates(map[string]any{
		"role": role, "updated_at": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *userRepository) UpdateProfile(uid string, email string) (*User, error) {
	result := r.db.Model(&User{}).Where("firebase_uid = ?", uid).Updates(map[string]any{
		"email": email, "updated_at": time.Now(),
	})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}
	return r.GetByUID(uid)
}
