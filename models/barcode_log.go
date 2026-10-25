package models

// BarcodeLog represents a recorded barcode log in BARCODE_LOGS table.
type BarcodeLog struct {
	ID              string `json:"id"`
	ItemNo          string `json:"item_no"`
	ItemName        string `json:"item_name"`
	Barcode         string `json:"barcode"`
	PreviousBarcode string `json:"previous_barcode"`
	UserID          string `json:"user_id"`
	Username        string `json:"username"`
	RecordedAt      string `json:"recorded_at"`
}

// StaffBarcodeStat represents the recording progress of a single staff member.
type StaffBarcodeStat struct {
	Username   string  `json:"username"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// BarcodeLogResponse wraps the list of logs and analytical summary statistics.
type BarcodeLogResponse struct {
	Logs             []BarcodeLog       `json:"logs"`
	TotalPeriod      int                `json:"total_period"`
	TopRecorder      string             `json:"top_recorder"`
	StaffStats       []StaffBarcodeStat `json:"staff_stats"`
	TotalStoreItems  int                `json:"total_store_items"`
	StoreWithBarcode int                `json:"store_with_barcode"`
	StoreProgressPct float64            `json:"store_progress_pct"`
}
