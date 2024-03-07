package ocr

type Client interface {
	GetChequeTranscription(photoUrl string) (*Cheque, error)
}
