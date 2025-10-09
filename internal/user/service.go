package user

import (
	"context"
	"errors"
	"strings"

	"github.com/sebaactis/wallet-go-api/internal/validation"
	"gorm.io/gorm"
)

type Service struct {
	repository *Repository
	validator  validation.StructValidator
	db         *gorm.DB
}

func NewService(repository *Repository, v validation.StructValidator) *Service {
	return &Service{repository: repository, db: repository.db, validator: v}
}

func (s *Service) Create(ctx context.Context, user *UserCreate) (*User, error) {
	name := strings.TrimSpace(user.Name)
	email := strings.TrimSpace(user.Email)

	if fields, ok := s.validator.ValidateStruct(user); !ok {
		return nil, &validation.ValidationError{Fields: fields}
	}

	exists, err := s.repository.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrDuplicateEmail
	}

	newUser := &User{
		Name:  name,
		Email: email,
	}

	if err := s.repository.Create(ctx, newUser); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrDuplicateEmail
		}

		return nil, err
	}

	return newUser, nil
}

func (s *Service) GetByID(ctx context.Context, id uint) (*User, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.repository.FindByEmail(ctx, email)
}
