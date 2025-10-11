package auth

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,min=6,max=32"`
	Password string `json:"password" validate:"required,min=8,max=30"`
}

type LoginResponse struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type UnlockUserReq struct {
	UserId uint `json:"userId"`
}


type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)
