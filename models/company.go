package models

// Company represents the COMPANY table schema.
// This is used for global shop settings.
type Company struct {
	CompanyName             string  `json:"shop_name"`
	AddressLine1            string  `json:"shop_address"`
	PhoneNo                 string  `json:"shop_phone"`
	ReceiptHeader           string  `json:"receipt_header"`
	ReceiptFooter           string  `json:"receipt_footer"`
	AdminPin                string  `json:"admin_pin"`
	EnableTax               bool    `json:"enable_tax"`
	TaxPercentage           float64 `json:"tax_percentage"`
	EnableServiceCharge     bool    `json:"enable_service_charge"`
	ServiceChargePercentage float64 `json:"service_charge_percentage"`
	// Additional fields like Printer IP can be added here if needed,
	// but are usually device-specific and saved locally.
}
