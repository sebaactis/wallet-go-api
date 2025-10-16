package account

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrAccountExists = errors.New("account already exists for user+currency")
	ErrUserNotFound  = errors.New("user not found")
	ErrCurrencyISO = errors.New("currency must be 3-letter ISO code")
)

type Service struct {
	repo *Repository
	db   *gorm.DB
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo, db: repo.db}
}


func (s *Service) Create(ctx context.Context, accountCreate *CreateAccountRequest) (*Account, error) {
	accountCreate.Currency = strings.ToUpper(strings.TrimSpace(accountCreate.Currency))

	if len(accountCreate.Currency) != 3 {
		return nil, ErrCurrencyISO
	}

	exists, err := s.repo.ExistsByUserAndCurrency(ctx, accountCreate.UserID, accountCreate.Currency)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAccountExists
	}

	acc := &Account{
		UserID:   accountCreate.UserID,
		Currency: accountCreate.Currency,
		Balance:  0,
	}
	if err := s.repo.Create(ctx, acc); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrAccountExists
		}
		return nil, err
	}
	return acc, nil
}

func (s *Service) GetBalance(ctx context.Context, accountID uint) (float64, error) {
	acc, err := s.repo.FindByID(ctx, accountID)
	if err != nil {
		return 0, err
	}
	return acc.Balance, nil
}
