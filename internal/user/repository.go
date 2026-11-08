package user

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	GetAllUsers() ([]*User, error)
	UpdateRole(id string, role string) error
	UpdateProfile(id string, updates map[string]any) (*User, error)
	UpdatePassword(id string, passwordHash string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("username = ?", username).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ?", email).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CreateUser(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetAllUsers() ([]*User, error) {
	var users []*User
	if err := r.db.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) UpdateRole(id string, role string) error {
	result := r.db.Model(&User{}).Where("id = ?", id).Updates(map[string]any{
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

func (r *userRepository) UpdateProfile(id string, updates map[string]any) (*User, error) {
	if len(updates) == 0 {
		return r.FindByID(context.Background(), id)
	}
	updates["updated_at"] = time.Now()
	result := r.db.Model(&User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}
	return r.FindByID(context.Background(), id)
}

func (r *userRepository) UpdatePassword(id string, passwordHash string) error {
	result := r.db.Model(&User{}).Where("id = ?", id).Updates(map[string]any{
		"password_hash": passwordHash,
		"updated_at":    time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}
