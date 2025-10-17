package user

import (
	"time"

	"github.com/sebaactis/wallet-go-api/internal/httputil"
)

type UserCreate struct {
	Name            string `json:"name"  validate:"required,min=5,max=30"`
	Email           string `json:"email" validate:"required,min=5,max=30,email"`
	Password        string `json:"password" validate:"required,min=8,max=30"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,min=8,max=30,eqfield=Password"`
}

type UserResponse struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	LoginAttempts int       `json:"login_attempt"`
	LockedUntil   time.Time `json:"locked_until"`
	CreatedAt     string    `json:"created_at"`
	UpdatedAt     string    `json:"updated_at"`
}

type UserRecoveryPassword struct {
	Email           string `json:"email" validate:"required,email,min=6,max=32"`
	Token           string `json:"token" validate:"required,min=1,max=1000"`
	Password        string `json:"password" validate:"required,min=8,max=30"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,min=8,max=30,eqfield=Password"`
}

func ToResponse(u *User) *UserResponse {
	return &UserResponse{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		LockedUntil:   u.Locked_until,
		LoginAttempts: u.LoginAttempt,
		CreatedAt:     httputil.FormatDate(&u.CreatedAt),
		UpdatedAt:     httputil.FormatDate(&u.UpdatedAt),
	}
}

func ToResponseMany(users []*User) []*UserResponse {
	response := make([]*UserResponse, len(users))

	for i, u := range users {
		response[i] = ToResponse(u)
	}

	return response
}
