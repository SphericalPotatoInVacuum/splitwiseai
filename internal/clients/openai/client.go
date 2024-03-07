package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/SphericalPotatoInVacuum/splitwiseai/internal/clients/ocr"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"go.uber.org/zap"
)

type Config struct {
	Enabled        bool   `env:"OAI_ENABLED" envDefault:"false"`
	ApiToken       string `env:"OAI_API_TOKEN"`
	ApiEndpoint    string `env:"OAI_API_ENDPOINT"`
	WhisperModelId string `env:"OAI_WHISPER_MODEL_ID"`
}

type client struct {
	openaiClient   *azopenai.Client
	whisperModelId string
}

func NewClient(cfg Config) (Client, error) {
	if !cfg.Enabled {
		zap.S().Debug("OpenAI client is disabled")
		return nil, nil
	}

	zap.S().Debug("Creating OpenAI client")
	keyCredential := azcore.NewKeyCredential(cfg.ApiToken)

	oaiClient, err := azopenai.NewClientForOpenAI(
		cfg.ApiEndpoint,
		keyCredential,
		&azopenai.ClientOptions{
			ClientOptions: azcore.ClientOptions{
				Logging: policy.LogOptions{
					IncludeBody: true,
				},
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create azopenai client: %w", err)
	}

	return &client{
		openaiClient:   oaiClient,
		whisperModelId: cfg.WhisperModelId,
	}, nil
}

func (c *client) GetTranscription(filePath string, prompt string) (*string, error) {
	voiceBytes, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	temp := float32(0.0)

	resp, err := c.openaiClient.GetAudioTranscription(context.TODO(), azopenai.AudioTranscriptionOptions{
		File: voiceBytes,

		ResponseFormat: to.Ptr(azopenai.AudioTranscriptionFormatText),

		DeploymentName: &c.whisperModelId,

		Temperature: &temp,

		Prompt: &prompt,
	}, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to get audio transcription: %w", err)
	}

	return resp.Text, nil
}

var (
	systemPrompt string = `
You are a robot that transcribes restaurant checks and outputs data in a json format.
Follow the following format:
{
  "Date": "2021-09-01",
  "Total": 100.0,
  "Items": [
    {
      "Name": "pizza",
      "Price": 20.0,
      "Quantity": 1,
      "Total": 20.0
    },
    {
      "Name": "coke",
      "Price": 5.0,
      "Quantity": 13,
      "Total": 65.0
    }
  ]
}
Strip down any markdown formatting and only return the json.
`
)

func (c *client) GetChequeTranscription(photoUrl string) (*ocr.Cheque, error) {
	resp, err := c.openaiClient.GetChatCompletions(
		context.TODO(),
		azopenai.ChatCompletionsOptions{
			Messages: []azopenai.ChatRequestMessageClassification{
				&azopenai.ChatRequestSystemMessage{
					Content: &systemPrompt,
				},
				&azopenai.ChatRequestUserMessage{
					Content: azopenai.NewChatRequestUserMessageContent([]azopenai.ChatCompletionRequestMessageContentPartClassification{
						&azopenai.ChatCompletionRequestMessageContentPartImage{
							ImageURL: &azopenai.ChatCompletionRequestMessageContentPartImageURL{
								Detail: to.Ptr(azopenai.ChatCompletionRequestMessageContentPartImageURLDetailHigh),
								URL:    &photoUrl,
							},
						},
					}),
				},
			},
			DeploymentName: to.Ptr("gpt-4-vision-preview"),
			MaxTokens:      to.Ptr[int32](1000),
		},
		&azopenai.GetChatCompletionsOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat completions: %w", err)
	}
	cheque := ocr.Cheque{}
	err = json.Unmarshal([]byte(*resp.ChatCompletions.Choices[0].Message.Content), &cheque)
	if _, ok := err.(*json.SyntaxError); ok {
		// TODO: bad json, ask chatgpt to correct itself
	}
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}
	return &cheque, nil
}
