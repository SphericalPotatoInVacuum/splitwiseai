package tgupdatesmq

import "context"

type Client interface {
	PublishMessage(ctx context.Context, message string) error
}
