package models

// Item represents the ITEM table schema.
type Item struct {
	ItemNo     string  `json:"itemno"`
	ItemUPC    string  `json:"itemupc"`
	ItemName   string  `json:"itemname"`
	CategoryID int     `json:"categoryid"`
	Price      float64 `json:"price"`
}
