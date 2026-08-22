package models

// CashReconciliation represents the CASH_RECONCILIATIONS table schema.
type CashReconciliation struct {
	ID             string  `json:"id"`
	Date           string  `json:"date"`
	Cashier        string  `json:"cashier_name"`
	ExpectedAmount float64 `json:"system_revenue"`
	ActualAmount   float64 `json:"actual_drawer_cash"`
	Difference     float64 `json:"difference"`
	AccuracyRate   float64 `json:"accuracy_rate,omitempty"`
	Note           string  `json:"notes"`
}
