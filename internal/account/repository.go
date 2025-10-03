package account

import (
	"context"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, account *Account) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*Account, error) {
	var acc Account

	if err := r.db.WithContext(ctx).First(&acc, id).Error; err != nil {
		return nil, err
	}

	return &acc, nil
}

func (r *Repository) FindByUserAndCurrency(ctx context.Context, userID uint, currency string) (*Account, error) {
	var acc Account
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND currency = ?", userID, currency).
		First(&acc).Error
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *Repository) ExistsByUserAndCurrency(ctx context.Context, userID uint, currency string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&Account{}).
		Where("user_id = ? AND currency = ?", userID, currency).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) UpdateBalance(ctx context.Context, id uint, newBalance int64) error {
	return r.db.WithContext(ctx).Model(&Account{}).
		Where("id = ?", id).
		Update("balance", newBalance).Error
}
