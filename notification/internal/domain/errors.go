package domain

import "errors"

var (
	ErrInvalidPayload   = errors.New("invalid event payload")
	ErrUsersFetch       = errors.New("failed to fetch users")
	ErrSendEmail        = errors.New("failed to send email")
	ErrPublishMessage   = errors.New("failed to publish message")
	ErrUnknownEventType = errors.New("unknown event type")
)
