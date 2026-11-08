package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const uniqueViolation = "23505"

type Repository interface {
	CreateUser(context.Context, *User) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) CreateUser(ctx context.Context, item *User) error {
	if item == nil {
		return fmt.Errorf("create user: user is nil")
	}
	err := r.db.WithContext(ctx).Create(item).Error
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrEmailAlreadyExists
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == uniqueViolation {
		return ErrEmailAlreadyExists
	}
	return fmt.Errorf("create user: %w", err)
}
