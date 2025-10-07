package user

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type Service struct {
	repository *Repository
	db         *gorm.DB
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository, db: repository.db}
}

func (s *Service) Create(ctx context.Context, user *UserCreate) (*User, error) {
	name := strings.TrimSpace(user.Name)
	email := strings.TrimSpace(user.Email)

	if name == "" || email == "" {
		return nil, errors.New("name and email are required")
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
