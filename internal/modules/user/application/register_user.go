package application

import (
	"context"
	"errors"

	"github.com/artsgoz/artsgoz-backend/internal/modules/user/domain"
	"github.com/artsgoz/artsgoz-backend/internal/platform/apperr"
	"github.com/artsgoz/artsgoz-backend/internal/platform/validator"
)

type RegisterUser struct {
	users domain.UserRepository
}

func NewRegisterUser(users domain.UserRepository) *RegisterUser {
	return &RegisterUser{users: users}
}

func (uc *RegisterUser) Execute(ctx context.Context, in RegisterUserInput) (RegisterUserOutput, error) {
	if err := validator.Struct(in); err != nil {
		return RegisterUserOutput{}, err
	}

	existing, err := uc.users.FindByEmail(ctx, in.Email)
	if err != nil {
		var appErr *apperr.Error
		if !errors.As(err, &appErr) || appErr.Kind != apperr.KindNotFound {
			return RegisterUserOutput{}, err
		}
	}
	if existing != nil {
		return RegisterUserOutput{}, apperr.Conflict("email already registered")
	}

	user, err := domain.NewUser(in.Email, in.Password)
	if err != nil {
		return RegisterUserOutput{}, err
	}
	if err := uc.users.Save(ctx, user); err != nil {
		return RegisterUserOutput{}, err
	}

	return RegisterUserOutput{ID: user.ID, Email: user.Email}, nil
}
