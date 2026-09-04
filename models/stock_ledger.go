package models

// StockLedger represents a record in the stock ledger
type StockLedger struct {
	ID          int     `json:"id"`
	ItemNo      string  `json:"item_no"`
	TxType      string  `json:"tx_type"` // IN, OUT, ADJ, RET
	Qty         float64 `json:"qty"`
	Timestamp   string  `json:"timestamp"`
	Notes       string  `json:"notes"`
	ReferenceID string  `json:"reference_id"`
}
