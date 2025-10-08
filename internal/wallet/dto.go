package wallet

type DepositRequest struct {
	AccountID uint    `json:"accountId" validate:"required"`
	Amount    float64 `json:"amount"    validate:"required,gt=0"`
	Currency  string  `json:"currency"  validate:"required,iso4217"`
}

type WithdrawRequest struct {
	AccountID uint    `json:"accountId" validate:"required"`
	Amount    float64 `json:"amount"    validate:"required,gt=0"`
	Currency  string  `json:"currency"  validate:"required,iso4217"`
}

type TransferRequest struct {
	FromAccountID uint    `json:"fromAccountId" validate:"required,nefield=ToAccountID"`
	ToAccountID   uint    `json:"toAccountId"   validate:"required"`
	Amount        float64 `json:"amount"        validate:"required,gt=0"`
	Currency      string  `json:"currency"      validate:"required,iso4217"`
}

type TxResponse struct {
	TransactionID uint    `json:"transactionId"`
	Type          string  `json:"type"`
	Reference     *string `json:"reference"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
}
