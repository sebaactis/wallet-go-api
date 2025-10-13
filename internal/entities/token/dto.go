package token

type TokenRequest struct {
	TokenType     string    `json:"token_type" validate:"required,max=30"`
	Token         string    `json:"token" validate:"required,max=1000"`
}