package mindee

type Client interface {
	GetPredictions(photoUrl string) (*Cheque, error)
}
