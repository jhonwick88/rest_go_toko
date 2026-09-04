package models

// Item represents the ITEM table schema.
type Item struct {
	ItemNo        string  `json:"itemno"`
	ItemUPC       string  `json:"itemupc"`
	ItemName      string  `json:"itemname"`
	ItemNameShort string  `json:"itemname_short"`
	CategoryID    int     `json:"categoryid"`
	Attribute     string  `json:"attribute"`
	Unit1         string  `json:"unit1"`
	Unit2         string  `json:"unit2"`
	Unit3         string  `json:"unit3"`
	Ratio2        float64 `json:"ratio2"`
	Ratio3        float64 `json:"ratio3"`
	DefUnitPrice1 float64 `json:"def_unitprice1"`
	DefUnitPrice2 float64 `json:"def_unitprice2"`
	DefUnitPrice3 float64 `json:"def_unitprice3"`
	ObQuantity    float64 `json:"obquantity"`
	Notes         string  `json:"notes"`
}
