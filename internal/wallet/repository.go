package wallet

import (
	"context"

	"github.com/sebaactis/wallet-go-api/internal/account"
	ledger "github.com/sebaactis/wallet-go-api/internal/legder"
	"github.com/sebaactis/wallet-go-api/internal/transaction"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) withTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

func (r *Repository) CreateTx(ctx context.Context, t *transaction.Transaction) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *Repository) FindTxByReference(ctx context.Context, ref string) (*transaction.Transaction, error) {
	var t transaction.Transaction

	if err := r.db.WithContext(ctx).Where("reference = ?", ref).First(&t).Error; err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *Repository) CreateEntries(ctx context.Context, entries ...*ledger.LedgerEntry) error {
	return r.db.WithContext(ctx).Create(&entries).Error
}

func (r *Repository) GetAccount(ctx context.Context, accountId uint, currency string) (*account.Account, error) {
	var a account.Account

	if err := r.db.WithContext(ctx).First(&a, accountId, currency).Error; err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *Repository) UpdateBalance(ctx context.Context, id uint, newBalance float64) error {
	return r.db.WithContext(ctx).Model(&account.Account{}).Where("id = ?", id).Update("balance", newBalance).Error
}
