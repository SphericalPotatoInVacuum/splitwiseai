package tgupdatesmq

import "context"

type Client interface {
	PublishMessage(ctx context.Context, message string, attributes map[string]string) error
}
