package token

import (
	"context"

	"github.com/sebaactis/wallet-go-api/internal/validation"
)

type Service struct {
	repository *Repository
	validator  validation.StructValidator
}

func NewService(repository *Repository, v validation.StructValidator) *Service {
	return &Service{repository: repository, validator: v}
}

func (s *Service) Create(ctx context.Context, tokenRequest *TokenRequest) (*Token, error) {
	if fields, ok := s.validator.ValidateStruct(tokenRequest); !ok {
		return nil, &validation.ValidationError{Fields: fields}
	}

	tokenCreate := &Token{
		TokenType: tokenRequest.TokenType,
		Token:     tokenRequest.Token,
	}

	token, err := s.repository.Create(ctx, tokenCreate)

	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *Service) GetAll(ctx context.Context) ([]*Token, error) {
	tokens, err := s.repository.GetAll(ctx)

	if err != nil {
		return nil, err
	}
	return tokens, nil
}
