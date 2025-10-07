package account

import (
	"time"

	"github.com/sebaactis/wallet-go-api/internal/user"
)

type Account struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	UserID    uint       `json:"user_id" gorm:"not null;index"`
	User      *user.User `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Currency  string     `json:"currency" gorm:"size:3;not null"`
	Balance   float64    `json:"balance" gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
