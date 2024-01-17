package mindee

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"splitwiseai/internal/clients/mindee/mindeeapi"
)

type Config struct {
	APIKey string `env:"MINDEE_API_TOKEN"`
}

type client struct {
	mindeeClient *mindeeapi.ClientWithResponses
}

var _ Client = (*client)(nil)

func NewClient(cfg Config) (Client, error) {
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

func (c *client) GetPredictions(filename string) (*mindeeapi.MindeeExpenseReceipts5DocPrediction, error) {
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	// Depending on the type of the document, you'll handle it differently.
	// Here's how you might handle a file object:
	file, err := os.Open("test.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	fileWriter, err := multipartWriter.CreateFormFile("document", file.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
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
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("api request failed with status code %d", resp.StatusCode())
	}
	prediction := resp.JSON201.Document.Inference.Prediction
	return prediction, nil
}
