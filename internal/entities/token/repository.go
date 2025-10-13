package token

import (
	"context"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, token *Token) (*Token, error) {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return nil, err
	}
	return token, nil
}

func (r *Repository) GetAll(ctx context.Context) ([]*Token, error) {
	var tokens []*Token
	if err := r.db.WithContext(ctx).Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}
