package user

import "time"

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
}

func ToResponse(u *User) *UserResponse {
	return &UserResponse{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		LockedUntil:   u.Locked_until,
		LoginAttempts: u.LoginAttempt,
	}
}

func ToResponseMany(users []*User) []*UserResponse {
	response := make([]*UserResponse, len(users))

	for i, u := range users {
		response[i] = ToResponse(u)
	}

	return response
}
