package transaction

import (
	"time"

	"github.com/sebaactis/wallet-go-api/internal/entities/account"
)

type Transaction struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	Type          string          `json:"type" gorm:"size:20;not null"`
	Reference     *string          `json:"reference" gorm:"size:100;index:idx_tx_ref,unique,where:reference IS NOT NULL"`
	FromAccountID *uint           `json:"from_account_id"`
	ToAccountID   *uint           `json:"to_account_id"`
	FromAccount   *account.Account `json:"from_account" gorm:"foreignKey:FromAccountID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ToAccount     *account.Account `json:"to_account" gorm:"foreignKey:ToAccountID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Amount        float64         `json:"amount" gorm:"type:decimal(10,2);not null"`
	Currency      string          `json:"currency" gorm:"size:3;not null"`
	CreatedAt     time.Time
}
