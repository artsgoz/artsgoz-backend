package user

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository interface {
	GetByEmail(email string) (*User, error)
	FindByUID(ctx context.Context, uid string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
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

// FindByUID returns nil, nil when the user has not completed registration.
func (r *userRepository) FindByUID(ctx context.Context, uid string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("firebase_uid = ?", uid).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser inserts a profile without overwriting one created by a concurrent request.
func (r *userRepository) CreateUser(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "firebase_uid"}},
		DoNothing: true,
	}).Create(user).Error
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
	return r.FindByUID(context.Background(), uid)
}
