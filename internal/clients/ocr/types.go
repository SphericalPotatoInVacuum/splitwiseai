package ocr

type Cheque struct {
	Date  string
	Items []Item
	Total *float32
}

type Item struct {
	Name      *string
	Price     *float32
	Quantity  *float32
	Total     *float32
	OrderedBy string
}
