package main

// Структура запроса API Gateway v1
type APIGatewayRequest struct {
	OperationID string `json:"operationId"`
	Resource    string `json:"resource"`

	HTTPMethod string `json:"httpMethod"`

	Path           string            `json:"path"`
	PathParameters map[string]string `json:"pathParameters"`

	Headers           map[string]string   `json:"headers"`
	MultiValueHeaders map[string][]string `json:"multiValueHeaders"`

	QueryStringParameters           map[string]string   `json:"queryStringParameters"`
	MultiValueQueryStringParameters map[string][]string `json:"multiValueQueryStringParameters"`

	Parameters           map[string]string   `json:"parameters"`
	MultiValueParameters map[string][]string `json:"multiValueParameters"`

	Body            string `json:"body"`
	IsBase64Encoded bool   `json:"isBase64Encoded,omitempty"`

	RequestContext interface{} `json:"requestContext"`
}

// Структура ответа API Gateway v1
type APIGatewayResponse struct {
	StatusCode        int                 `json:"statusCode"`
	Headers           map[string]string   `json:"headers"`
	MultiValueHeaders map[string][]string `json:"multiValueHeaders"`
	Body              string              `json:"body"`
	IsBase64Encoded   bool                `json:"isBase64Encoded,omitempty"`
}

type Event struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	EventMetadata EventMetadata `json:"event_metadata"`
	Details       Details       `json:"details"`
}

type EventMetadata struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	CreatedAt string `json:"created_at"`
	CloudID   string `json:"cloud_id"`
	FolderID  string `json:"folder_id"`
}

type Details struct {
	QueueID string         `json:"queue_id"`
	Message MessageDetails `json:"message"`
}

type MessageDetails struct {
	MessageID              string               `json:"message_id"`
	MD5OfBody              string               `json:"md5_of_body"`
	Body                   string               `json:"body"`
	Attributes             map[string]string    `json:"attributes"`
	MessageAttributes      map[string]Attribute `json:"message_attributes"`
	MD5OfMessageAttributes string               `json:"md5_of_message_attributes"`
}

type Attribute struct {
	DataType    string `json:"data_type"`
	StringValue string `json:"string_value"`
}
