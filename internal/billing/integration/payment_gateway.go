package integration

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)

// PaymentGateway defines the interface for external payment systems
type PaymentGateway interface {
	GeneratePaymentLink(txID uuid.UUID, amount float64, returnURL string) (string, error)
	VerifySignature(payload []byte, signature string) bool
}

// MockClickGateway implements PaymentGateway for Click
type MockClickGateway struct {
	SecretKey string
}

func NewMockClickGateway(secret string) *MockClickGateway {
	return &MockClickGateway{SecretKey: secret}
}

func (m *MockClickGateway) GeneratePaymentLink(txID uuid.UUID, amount float64, returnURL string) (string, error) {
	// Mock URL generation
	return fmt.Sprintf("https://mock.click.uz/pay?tx_id=%s&amount=%.2f", txID.String(), amount), nil
}

func (m *MockClickGateway) VerifySignature(payload []byte, signature string) bool {
	// Simple HMAC-SHA1 verification for Click
	h := hmac.New(sha1.New, []byte(m.SecretKey))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// MockPaymeGateway implements PaymentGateway for Payme
type MockPaymeGateway struct {
	SecretKey string
}

func NewMockPaymeGateway(secret string) *MockPaymeGateway {
	return &MockPaymeGateway{SecretKey: secret}
}

func (m *MockPaymeGateway) GeneratePaymentLink(txID uuid.UUID, amount float64, returnURL string) (string, error) {
	// Mock URL generation
	return fmt.Sprintf("https://mock.payme.uz/pay?tx_id=%s&amount=%.2f", txID.String(), amount), nil
}

func (m *MockPaymeGateway) VerifySignature(payload []byte, signature string) bool {
	// Simple HMAC-SHA1 verification for Payme (mocked same as Click for simplicity)
	h := hmac.New(sha1.New, []byte(m.SecretKey))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}
