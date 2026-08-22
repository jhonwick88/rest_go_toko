package models

// Item represents the ITEM table schema.
type Item struct {
	ItemNo        string  `json:"itemno"`
	ItemUPC       string  `json:"itemupc"`
	ItemName      string  `json:"itemname"`
	ItemNameShort string  `json:"itemname_short"`
	CategoryID    int     `json:"categoryid"`
	DefUnitPrice1 float64 `json:"def_unitprice1"`
	ObQuantity    float64 `json:"obquantity"`
	Notes         string  `json:"notes"`
}
