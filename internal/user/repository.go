package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByFirebaseUID(ctx context.Context, uid string) (*User, error)
	LinkFirebaseUID(ctx context.Context, id string, firebaseUID string) error
	CreateUser(ctx context.Context, user *User) error
	GetAllUsers() ([]*User, error)
	UpdateRole(id string, role string) error
	UpdateProfile(id string, username string, email string, avatarURL string) (*User, error)
	UpdatePassword(id string, passwordHash string) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return NewGormRepository(db)
}

var ErrEmailAlreadyExists = errors.New("email already registered")

var ErrFirebaseUIDAlreadyLinked = errors.New("profile is already linked to another Firebase account")

const uniqueViolation = "23505"

var _ UserRepository = (*GormRepository)(nil)

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindByID(ctx context.Context, id string) (*User, error) {
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

func (r *GormRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
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

func (r *GormRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
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

func (r *GormRepository) FindByFirebaseUID(ctx context.Context, uid string) (*User, error) {
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

func (r *GormRepository) LinkFirebaseUID(ctx context.Context, id string, firebaseUID string) error {
	result := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).
		Where("firebase_uid IS NULL OR firebase_uid = '' OR firebase_uid = ?", firebaseUID).Updates(map[string]any{
		"firebase_uid": firebaseUID,
		"updated_at":   time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrFirebaseUIDAlreadyLinked
	}
	return nil
}

func (r *GormRepository) CreateUser(ctx context.Context, item *User) error {
	if item == nil {
		return fmt.Errorf("create user: user is nil")
	}
	// Email-only registration still needs a unique username for profile APIs.
	if item.Username == "" {
		item.Username = item.ID
	}
	query := r.db.WithContext(ctx)
	// Unlinked accounts must store NULL rather than sharing an empty unique UID.
	if item.FirebaseUID == "" {
		query = query.Omit("FirebaseUID")
	}
	err := query.Create(item).Error
	if err == nil {
		return nil
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == uniqueViolation {
		if postgresError.ConstraintName == "idx_users_email" || postgresError.ConstraintName == "users_email_key" {
			return ErrEmailAlreadyExists
		}
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrEmailAlreadyExists
	}
	return fmt.Errorf("create user: %w", err)
}

func (r *GormRepository) GetAllUsers() ([]*User, error) {
	var users []*User
	if err := r.db.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *GormRepository) UpdateRole(id string, role string) error {
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

func (r *GormRepository) UpdateProfile(id string, username string, email string, avatarURL string) (*User, error) {
	updates := map[string]any{"updated_at": time.Now()}
	if username != "" {
		updates["username"] = username
	}
	if email != "" {
		updates["email"] = email
	}
	if avatarURL != "" {
		updates["avatar_url"] = avatarURL
	}
	result := r.db.Model(&User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}
	return r.FindByID(context.Background(), id)
}

func (r *GormRepository) UpdatePassword(id string, passwordHash string) error {
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
