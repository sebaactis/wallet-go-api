package ledger

import (
	"time"

	"github.com/sebaactis/wallet-go-api/internal/account"
	"github.com/sebaactis/wallet-go-api/internal/transaction"
)

type LedgerEntry struct {
	ID            uint                    `json:"id" gorm:"primaryKey"`
	TransactionID uint                    `json:"transaction_id" gorm:"not null;index"`
	Transaction   *transaction.Transaction `json:"transaction" gorm:"foreignKey:TransactionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	AccountID     uint                    `json:"account_id" gorm:"not null;index"`
	Account       *account.Account         `json:"account" gorm:"foreignKey:AccountID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Amount        int64                   `json:"amount" gorm:"not null"`
	CreatedAt     time.Time
}
