package models

// Supplier represents a supplier in the database.
type Supplier struct {
	SupplierID int    `json:"supplier_id"`
	Name       string `json:"name"`
	Contact    string `json:"contact"`
	Address    string `json:"address"`
	Notes      string `json:"notes"`
}
