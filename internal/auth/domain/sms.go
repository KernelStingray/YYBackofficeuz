package domain

import (
	"context"
)

// SMSProvider defines an interface for external SMS gateways like Eskiz.uz.
type SMSProvider interface {
	// SendSMS sends a message to the provided phone number.
	SendSMS(ctx context.Context, phone string, message string) error
}
