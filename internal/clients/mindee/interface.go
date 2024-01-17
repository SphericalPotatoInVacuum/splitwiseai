package mindee

import "splitwiseai/internal/clients/mindee/mindeeapi"

type Client interface {
	GetPredictions(filename string) (*mindeeapi.MindeeExpenseReceipts5DocPrediction, error)
}
