package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/artsgoz/artsgoz-backend/internal/user/domain"
	"github.com/artsgoz/artsgoz-backend/internal/user/usecase"
)

const uniqueViolation = "23505"

type PostgresUserRepository struct {
	database *gorm.DB
}

var _ usecase.UserRepository = (*PostgresUserRepository)(nil)

type userRecord struct {
	ID           string    `gorm:"column:id;type:uuid;primaryKey"`
	Email        string    `gorm:"column:email;uniqueIndex;not null"`
	PasswordHash string    `gorm:"column:password_hash;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

func (userRecord) TableName() string { return "users" }

func NewPostgresUserRepository(database *gorm.DB) *PostgresUserRepository {
	return &PostgresUserRepository{database: database}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user domain.User) error {
	record := userRecord{
		ID: user.ID(), Email: user.Email(), PasswordHash: user.PasswordHash(), CreatedAt: user.CreatedAt(),
	}
	err := r.database.WithContext(ctx).Create(&record).Error
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return usecase.ErrEmailAlreadyExists
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == uniqueViolation {
		return usecase.ErrEmailAlreadyExists
	}
	return fmt.Errorf("create user: %w", err)
}
