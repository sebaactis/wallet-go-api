package user

import (
	"context"
	"errors"
	"strings"

	"github.com/sebaactis/wallet-go-api/internal/validation"
	"golang.org/x/crypto/bcrypt"
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
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(user.Password)), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

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
		Name:     name,
		Email:    email,
		Password: string(passwordHash),
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

func (s *Service) FindAll(ctx context.Context) ([]*User, error) {
	return s.repository.FindAll(ctx)
}

func (s *Service) IncrementLoginAttempt(ctx context.Context, id uint) (int, error) {
	return s.repository.IncrementLoginAttempt(ctx, id)
}

func (s *Service) UnlockUser(ctx context.Context, id uint) error {
	return s.repository.UnlockUser(ctx, id)
}

func (s *Service) UpdatePasswordByRecovery(ctx context.Context, req UserRecoveryPassword) (*User, error) {

	if fields, ok := s.validator.ValidateStruct(req); !ok {
		return nil, &validation.ValidationError{Fields: fields}
	}

	user, err := s.repository.FindByEmail(ctx, req.Email)

	if err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(req.Password)), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	err = s.repository.UpdatePassword(ctx, user.ID, string(passwordHash))

	if err != nil {
		return nil, err
	}

	return user, nil
}
