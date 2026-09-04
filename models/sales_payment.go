package models

// SalesPayment represents a payment for a sale.
type SalesPayment struct {
	ID            int     `json:"id"`
	InvoiceNo     string  `json:"invoiceno"`
	PaymentMethod string  `json:"payment_method"`
	Amount        float64 `json:"amount"`
	PaymentDate   string  `json:"payment_date"`
}
