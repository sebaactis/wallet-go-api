package wallet


type DepositRequest struct {
	AccountID uint    `json:"accountId"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

type WithdrawRequest struct {
	AccountID uint    `json:"accountId"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

type TransferRequest struct {
	FromAccountID uint    `json:"fromAccountId"`
	ToAccountID   uint    `json:"toAccountId"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
}

type TxResponse struct {
	TransactionID uint    `json:"transactionId"`
	Type          string  `json:"type"`
	Reference     *string  `json:"reference"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
}
