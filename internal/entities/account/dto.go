package account

type CreateAccountRequest struct {
	UserID   uint   `json:"userId"   validate:"required"`
	Currency string `json:"currency" validate:"required,iso4217"`
}

type AccountResponse struct {
	ID       uint    `json:"id"`
	UserID   uint    `json:"userId"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"` // menor unidad
}

type BalanceResponse struct {
	AccountID uint    `json:"accountId"`
	Currency  string  `json:"currency"`
	Balance   float64 `json:"balance"`
}

func ToResponse(a *Account) *AccountResponse {
	return &AccountResponse{
		ID:       a.ID,
		UserID:   a.UserID,
		Currency: a.Currency,
		Balance:  a.Balance,
	}
}
