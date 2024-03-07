package ocr

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/ocr/mindeeapi"

	"go.uber.org/zap"
)

type Config struct {
	Enabled bool   `env:"MINDEE_ENABLED" envDefault:"false"`
	APIKey  string `env:"MINDEE_API_TOKEN"`
}

type client struct {
	mindeeClient *mindeeapi.ClientWithResponses
}

var _ Client = (*client)(nil)

func NewClient(cfg Config) (Client, error) {
	if !cfg.Enabled {
		zap.S().Debug("Mindee client is disabled")
		return nil, nil
	}

	zap.S().Debug("Creating Mindee client")

	mindeeClient, err := mindeeapi.NewClientWithResponses(
		"https://api.mindee.net/v1/",
		mindeeapi.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Token "+cfg.APIKey)
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	return &client{mindeeClient: mindeeClient}, nil
}

func (c *client) GetChequeTranscription(photoUrl string) (*Cheque, error) {
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	fileWriter, err := multipartWriter.CreateFormField("document")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = fileWriter.Write([]byte(photoUrl))
	if err != nil {
		return nil, fmt.Errorf("failed to write form file: %w", err)
	}

	// Close the multipart writer to finalize the form
	multipartWriter.Close()

	resp, err := c.mindeeClient.PostProductsMindeeExpenseReceiptsVersionPredictWithBodyWithResponse(
		context.TODO(),
		mindeeapi.V5,
		"multipart/form-data",
		&requestBody,
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make api request: %w", err)
	}
	if resp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("api request failed with status code %d", resp.StatusCode())
	}
	zap.S().Debug("Completed API request")
	prediction := resp.JSON201.Document.Inference.Prediction

	cheque := Cheque{
		Date:  prediction.Date.Value.Format("2006-01-02"),
		Total: prediction.TotalAmount.Value,
		Items: make([]Item, 0, len(*prediction.LineItems)),
	}

	for _, lineItem := range *prediction.LineItems {
		item := Item{
			Name:     lineItem.Description,
			Price:    lineItem.UnitPrice,
			Quantity: lineItem.Quantity,
			Total:    lineItem.TotalAmount,
		}
		cheque.Items = append(cheque.Items, item)
	}

	return &cheque, nil
}
