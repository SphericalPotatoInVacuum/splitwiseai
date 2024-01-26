package openai

import "splitwiseai/internal/clients/ocr"

type Client interface {
	GetTranscription(filePath string, prompt string) (*string, error)
	GetChequeTranscription(photoUrl string) (*ocr.Cheque, error)
}
