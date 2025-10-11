package user

import "time"

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"size:30;not null"`
	Email        string    `json:"email" gorm:"size:30;not null;uniqueIndex"`
	Password     string    `json:"password" gorm:"size:30;not null"`
	LoginAttempt int       `json:"login_attempt" gorm:"default:0"`
	Locked_until time.Time `json:"locked_until" gorm:"default:null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
