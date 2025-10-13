package wallet

import (
	"context"
	"errors"
	"strings"

	ledger "github.com/sebaactis/wallet-go-api/internal/entities/legder"
	"github.com/sebaactis/wallet-go-api/internal/entities/transaction"
	"gorm.io/gorm"
)

var (
	ErrAccountNotFound   = errors.New("account not found")
	ErrCurrencyMismatch  = errors.New("currency mismatch")
	ErrNegativeAmount    = errors.New("amount must be > 0")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrSameAccount       = errors.New("from and to accounts are the same")
)

type Service struct {
	db   *gorm.DB
	repo *Repository
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db, repo: NewRepository(db)}
}

func (s *Service) Deposit(ctx context.Context, depositRequest *DepositRequest, ref string) (*transaction.Transaction, error) {
	depositRequest.Currency = strings.ToUpper(depositRequest.Currency)

	if depositRequest.Amount <= 0 {
		return nil, ErrNegativeAmount
	}

	if ref != "" {
		if t, err := s.repo.FindTxByReference(ctx, ref); err == nil {
			return t, nil
		}
	}

	var out *transaction.Transaction

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		r := s.repo.withTx(tx)

		acc, err := r.GetAccount(ctx, depositRequest.AccountID, depositRequest.Currency)
		if err != nil {
			return ErrAccountNotFound
		}

		if acc.Currency != depositRequest.Currency {
			return ErrCurrencyMismatch
		}

		t := &transaction.Transaction{
			Type:        "deposit",
			Reference:   toRefPtr(ref),
			ToAccountID: &acc.ID,
			Amount:      depositRequest.Amount,
			Currency:    depositRequest.Currency,
		}

		if err := r.CreateTx(ctx, t); err != nil {

			if errors.Is(err, gorm.ErrDuplicatedKey) && ref != "" {
				if prev, e := r.FindTxByReference(ctx, ref); e == nil {
					out = prev
					return nil
				}
			}
			return err
		}

		entry := &ledger.LedgerEntry{
			TransactionID: t.ID,
			AccountID:     acc.ID,
			Amount:        depositRequest.Amount,
		}

		if err := r.CreateEntries(ctx, entry); err != nil {
			return err
		}

		newBal := acc.Balance + depositRequest.Amount
		if err := r.UpdateBalance(ctx, acc.ID, newBal); err != nil {
			return err
		}

		out = t
		return nil
	})

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Service) Withdraw(ctx context.Context, withdrawRequest *WithdrawRequest, ref string) (*transaction.Transaction, error) {
	withdrawRequest.Currency = strings.ToUpper(withdrawRequest.Currency)

	if withdrawRequest.Amount <= 0 {
		return nil, ErrNegativeAmount
	}

	if ref != "" {
		if t, err := s.repo.FindTxByReference(ctx, ref); err == nil {
			return t, nil
		}
	}

	var out *transaction.Transaction

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		r := s.repo.withTx(tx)

		acc, err := r.GetAccount(ctx, withdrawRequest.AccountID, withdrawRequest.Currency)

		if err != nil {
			return ErrAccountNotFound
		}
		if acc.Currency != withdrawRequest.Currency {
			return ErrCurrencyMismatch
		}
		if acc.Balance < withdrawRequest.Amount {
			return ErrInsufficientFunds
		}

		t := &transaction.Transaction{
			Type:          "withdraw",
			Reference:     toRefPtr(ref),
			FromAccountID: &acc.ID,
			Amount:        withdrawRequest.Amount,
			Currency:      withdrawRequest.Currency,
		}

		if err := r.CreateTx(ctx, t); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) && ref != "" {
				if prev, e := r.FindTxByReference(ctx, ref); e == nil {
					out = prev
					return nil
				}
			}
			return err
		}

		entry := &ledger.LedgerEntry{
			TransactionID: t.ID,
			AccountID:     acc.ID,
			Amount:        -withdrawRequest.Amount,
		}
		if err := r.CreateEntries(ctx, entry); err != nil {
			return err
		}

		newBal := acc.Balance - withdrawRequest.Amount
		if err := r.UpdateBalance(ctx, acc.ID, newBal); err != nil {
			return err
		}

		out = t
		return nil

	})

	if err != nil {
		return nil, err
	}

	return out, nil

}

func (s *Service) Transfer(ctx context.Context, transferRequest *TransferRequest, ref string) (*transaction.Transaction, error) {
	transferRequest.Currency = strings.ToUpper(strings.TrimSpace(transferRequest.Currency))

	if transferRequest.Amount <= 0 {
		return nil, ErrNegativeAmount
	}
	if transferRequest.FromAccountID == transferRequest.ToAccountID {
		return nil, ErrSameAccount
	}

	if ref != "" {
		if t, err := s.repo.FindTxByReference(ctx, ref); err == nil {
			return t, nil
		}
	}

	var out *transaction.Transaction

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		r := s.repo.withTx(tx)

		from, err := r.GetAccount(ctx, transferRequest.FromAccountID, transferRequest.Currency)
		if err != nil {
			return ErrAccountNotFound
		}

		to, err := r.GetAccount(ctx, transferRequest.ToAccountID, transferRequest.Currency)
		if err != nil {
			return ErrAccountNotFound
		}

		if from.Currency != transferRequest.Currency || to.Currency != transferRequest.Currency {
			return ErrCurrencyMismatch
		}
		if from.Balance < transferRequest.Amount {
			return ErrInsufficientFunds
		}

		t := &transaction.Transaction{
			Type:          "transfer",
			Reference:     toRefPtr(ref),
			FromAccountID: &from.ID,
			ToAccountID:   &to.ID,
			Amount:        transferRequest.Amount,
			Currency:      transferRequest.Currency,
		}

		if err := r.CreateTx(ctx, t); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) && ref != "" {
				if prev, e := r.FindTxByReference(ctx, ref); e == nil {
					out = prev
					return nil
				}
			}
			return err
		}

		debit := &ledger.LedgerEntry{TransactionID: t.ID, AccountID: from.ID, Amount: -transferRequest.Amount}
		credit := &ledger.LedgerEntry{TransactionID: t.ID, AccountID: to.ID, Amount: +transferRequest.Amount}

		if err := r.CreateEntries(ctx, debit, credit); err != nil {
			return err
		}

		if err := r.UpdateBalance(ctx, from.ID, from.Balance-transferRequest.Amount); err != nil {
			return err
		}
		if err := r.UpdateBalance(ctx, to.ID, to.Balance+transferRequest.Amount); err != nil {
			return err
		}

		out = t
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func toRefPtr(ref string) *string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil
	}
	return &ref
}
