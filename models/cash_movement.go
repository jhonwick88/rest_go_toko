package models

// CashMovement represents a cash movement record in the drawer (e.g. petty cash out, cash in).
type CashMovement struct {
	ID          string  `json:"id"`
	Date        string  `json:"date"`
	CashierName string  `json:"cashier_name"`
	Type        string  `json:"type"`     // "OUT" or "IN"
	Category    string  `json:"category"` // e.g. "Sumbangan", "Pengamen", "Operasional", "Konsumsi", "Lain-lain"
	Amount      float64 `json:"amount"`
	Notes       string  `json:"notes"`
}
