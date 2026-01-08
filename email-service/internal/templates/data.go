package templates

type NewUserData struct {
	Username string
	LoginURL string
}

type PaymentFailedData struct {
	Username string
	Reason   string
	RetryURL string
}

type PaymentSuccessData struct {
	Username      string
	Amount        string
	TransactionID string
	DashboardURL  string
}

type PaymentRefundData struct {
	Username string
	Amount   string
	RefundID string
}
