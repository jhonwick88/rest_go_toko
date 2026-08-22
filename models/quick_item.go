package models

// QuickItem represents the QUICK_ITEMS table schema.
type QuickItem struct {
	ID           string `json:"id"`
	ItemNo       string `json:"item_no"`
	ItemName     string `json:"item_name"`
	IconName     string `json:"icon_name,omitempty"`
	ColorHex     string `json:"color_hex,omitempty"`
	DisplayOrder int    `json:"display_order"`
	IsActive     int    `json:"is_active"`
}
