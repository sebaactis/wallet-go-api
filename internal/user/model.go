package user

import "time"

type User struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name" gorm:"size:200;not null"`
	Email     string `json:"email" gorm:"size:200;not null;uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserCreate struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}
