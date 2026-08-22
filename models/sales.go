package models

// Sales represents the SALES table schema.
type Sales struct {
	ID            string  `json:"id"`
	ReceiptNo     string  `json:"receipt_no"`
	SaleDate      string  `json:"sale_date"`
	CashierID     string  `json:"cashier_id"`
	Subtotal      float64 `json:"subtotal"`
	Discount      float64 `json:"discount"`
	Total         float64 `json:"total"`
	Tax           float64 `json:"tax"`
	ServiceCharge float64 `json:"service_charge,omitempty"`
	PaymentMethod string  `json:"payment_method"`
	CashTendered  float64 `json:"cash_tendered"`
	ChangeDue     float64 `json:"change_due"`
	Status        string  `json:"status"`
	VoidReason    string  `json:"void_reason,omitempty"`
	VoidedAt      string  `json:"voided_at,omitempty"`
	VoidedBy      string  `json:"voided_by,omitempty"`
	RefundReason  string  `json:"refund_reason,omitempty"`
	RefundedAt    string  `json:"refunded_at,omitempty"`
	RefundedBy    string  `json:"refunded_by,omitempty"`
	CustomerID    *int    `json:"customer_id,omitempty"`
	IsVoid        int     `json:"is_void"`
	Items         []SalesItem `json:"items"`
}

// SalesItem represents the SALES_ITEMS table schema.
type SalesItem struct {
	ID       int     `json:"id"`
	SaleID   string  `json:"sale_id"`
	ItemNo   string  `json:"itemno"`
	ItemName string  `json:"itemname"`
	ItemUPC  string  `json:"itemupc"`
	Qty      float64 `json:"qty"`
	Price    float64 `json:"price"`
	Discount float64 `json:"discount"`
	Total    float64 `json:"total"`
	Note     string  `json:"note"`
}

// CreateSalesRequest represents the payload for creating a new sale with its items.
type CreateSalesRequest struct {
	Sales
}
