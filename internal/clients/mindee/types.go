package mindee

type Cheque struct {
	Date  string
	Time  string
	Items []Item
	Total float64
}

type Item struct {
	Name      string
	Price     float64
	Quantity  int
	Total     float64
	OrderedBy string
}
